package agent

// reverse proxy for grpc infer and metadata endpoints

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
)

const (
	ReverseGRPCProxyPort          = 9998
	grpcProxyMaxConcurrentStreams = 100
	maxConnsPerHostGRPC           = 10
)

type reverseGRPCProxy struct {
	v2.UnimplementedGRPCInferenceServiceServer
	logger                log.FieldLogger
	stateManager          *LocalStateManager
	grpcServer            *grpc.Server
	serverReady           bool
	backendGRPCServerHost string
	backendGRPCServerPort uint
	v2GRPCClientPool      []v2.GRPCInferenceServiceClient
	port                  uint // service port
	mu                    sync.RWMutex
	metrics               metrics.AgentMetricsHandler
	callOptions           []grpc.CallOption
	modelLagStats         interfaces.ModelScalingStats
	modelLastUsedStats    interfaces.ModelScalingStats
}

func NewReverseGRPCProxy(
	metricsHandler metrics.AgentMetricsHandler,
	logger log.FieldLogger, backendGRPCServerHost string,
	backendGRPCServerPort uint,
	servicePort uint,
	modelLagStats interfaces.ModelScalingStats,
	modelLastUsedStats interfaces.ModelScalingStats,
) *reverseGRPCProxy {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	return &reverseGRPCProxy{
		logger:                logger.WithField("Source", "GRPCProxy"),
		backendGRPCServerHost: backendGRPCServerHost,
		backendGRPCServerPort: backendGRPCServerPort,
		port:                  servicePort,
		metrics:               metricsHandler,
		callOptions:           opts,
		modelLagStats:         modelLagStats,
		modelLastUsedStats:    modelLastUsedStats,
	}
}

func (rp *reverseGRPCProxy) SetState(sm interface{}) {
	rp.stateManager = sm.(*LocalStateManager)
}

func (rp *reverseGRPCProxy) Start() error {
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", rp.port))
	if err != nil {
		rp.logger.Errorf("unable to start gRPC listening server on port %d", rp.port)
		return err
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
	}
	opts = append(opts, grpc.MaxConcurrentStreams(grpcProxyMaxConcurrentStreams))
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(otelgrpc.UnaryServerInterceptor(), rp.metrics.UnaryServerInterceptor())))
	grpcServer := grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(grpcServer, rp)

	rp.logger.Infof("Setting grpc v2 client pool on port %d", rp.backendGRPCServerPort)

	conns, clients, err := createV2CRPCClients(rp.backendGRPCServerHost, int(rp.backendGRPCServerPort), maxConnsPerHostGRPC)
	if err != nil {
		return err
	}

	rp.v2GRPCClientPool = clients

	rp.logger.Infof("Starting gRPC listening server on port %d", rp.port)
	rp.grpcServer = grpcServer
	go func() {
		defer closeV2CRPCConnections(conns)
		rp.mu.Lock()
		rp.serverReady = true
		rp.mu.Unlock()
		err := rp.grpcServer.Serve(l)
		rp.logger.WithError(err).Infof("Reverse GRPC proxy stopped with error")
		rp.mu.Lock()
		rp.serverReady = false
		rp.mu.Unlock()
	}()
	return nil
}

func (rp *reverseGRPCProxy) Stop() error {
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.grpcServer.GracefulStop()
	rp.serverReady = false
	return nil
}

func (rp *reverseGRPCProxy) Ready() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.serverReady
}

func (rp *reverseGRPCProxy) Name() string {
	return "Reverse GRPC Proxy"
}

func (rp *reverseGRPCProxy) extractModelNamesFromContext(ctx context.Context) (string, string, error) {
	var internalModelName, externalModelName string
	var inHeader bool
	if internalModelName, externalModelName, inHeader = extractModelNamesFromHeaders(ctx); inHeader {
		rp.logger.Debugf("Extracted model name %s:%s %s:%s", resources.SeldonInternalModelHeader, internalModelName, resources.SeldonModelHeader, externalModelName)
		return internalModelName, externalModelName, nil
	} else {
		msg := fmt.Sprintf("Failed to extract model name %s:[%s] %s:[%s]", resources.SeldonInternalModelHeader, internalModelName, resources.SeldonModelHeader, externalModelName)
		rp.logger.Error(msg)
		return "", "", status.Error(codes.FailedPrecondition, msg)
	}
}

func (rp *reverseGRPCProxy) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	logger := rp.logger.WithField("func", "ModelInfer")
	internalModelName, externalModelName, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.ModelName = internalModelName
	r.ModelVersion = ""

	// to sync between Inc and Dec call running in go routines
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := rp.modelLagStats.IncDefault(internalModelName)
		if err != nil {
			logger.WithError(err).Warnf("Cannot increment metrics for %s", internalModelName)
		}
		wg.Done()
		err = rp.modelLastUsedStats.IncDefault(internalModelName)
		if err != nil {
			logger.WithError(err).Warnf("Cannot increment metrics for %s", internalModelName)
		}
	}()
	defer func() {
		wg.Wait() // make sure that Inc is called first
		err := rp.modelLagStats.DecDefault(internalModelName)
		if err != nil {
			logger.WithError(err).Warnf("Cannot decrement metrics for %s", internalModelName)
		}
	}()

	startTime := time.Now()
	err = rp.ensureLoadModel(r.ModelName)
	if err != nil {
		elapsedTime := time.Since(startTime).Seconds()
		go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, codes.NotFound.String())
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.ModelName, err))
	}
	var trailer metadata.MD
	opts := append(rp.callOptions, grpc.Trailer(&trailer))
	resp, err := rp.getV2GRPCClient().ModelInfer(ctx, r, opts...)
	if retryForLazyReload(err) {
		rp.stateManager.v2Client.LoadModel(internalModelName)
		resp, err = rp.getV2GRPCClient().ModelInfer(ctx, r, opts...)
	}

	grpcStatus, _ := status.FromError(err)
	elapsedTime := time.Since(startTime).Seconds()
	go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeGrpc, elapsedTime, grpcStatus.Code().String())
	errTrailer := grpc.SetTrailer(ctx, trailer) // pass on any trailers set by inference server such as MLServer
	if errTrailer != nil {
		logger.WithError(errTrailer).Error("Failed to set trailers")
	}
	return resp, err
}

func (rp *reverseGRPCProxy) ModelMetadata(ctx context.Context, r *v2.ModelMetadataRequest) (*v2.ModelMetadataResponse, error) {
	internalModelName, _, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.Name = internalModelName
	r.Version = ""

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	resp, err := rp.getV2GRPCClient().ModelMetadata(ctx, r)
	if retryForLazyReload(err) {
		rp.stateManager.v2Client.LoadModel(internalModelName)
		resp, err = rp.getV2GRPCClient().ModelMetadata(ctx, r)
	}
	return resp, err
}

func (rp *reverseGRPCProxy) ModelReady(ctx context.Context, r *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {
	internalModelName, _, err := rp.extractModelNamesFromContext(ctx)
	if err != nil {
		return nil, err
	}
	r.Name = internalModelName
	r.Version = ""

	if err := rp.ensureLoadModel(r.Name); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found (err: %s)", r.Name, err))
	}

	resp, err := rp.getV2GRPCClient().ModelReady(ctx, r)
	if retryForLazyReload(err) {
		rp.stateManager.v2Client.LoadModel(internalModelName)
		resp, err = rp.getV2GRPCClient().ModelReady(ctx, r)
	}
	return resp, err
}

func (rp *reverseGRPCProxy) ensureLoadModel(modelId string) error {
	return rp.stateManager.EnsureLoadModel(modelId)
}

func (rp *reverseGRPCProxy) getV2GRPCClient() v2.GRPCInferenceServiceClient {
	i := rand.Intn(len(rp.v2GRPCClientPool))
	return rp.v2GRPCClientPool[i]
}

func getV2GrpcConnection(host string, plainTxtPort int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_retry.UnaryClientInterceptor(retryOpts...), otelgrpc.UnaryClientInterceptor())),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, plainTxtPort), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createV2CRPCClients(backendGRPCServerHost string, backendGRPCServerPort int, size int) ([]*grpc.ClientConn, []v2.GRPCInferenceServiceClient, error) {
	conns := make([]*grpc.ClientConn, size)
	clients := make([]v2.GRPCInferenceServiceClient, size)
	for i := 0; i < size; i++ {
		conn, err := getV2GrpcConnection(backendGRPCServerHost, backendGRPCServerPort)

		if err != nil {
			// TODO: this could fail in later iterations, so close earlier connections
			conn.Close()
			return nil, nil, err
		}

		conns[i] = conn
		clients[i] = v2.NewGRPCInferenceServiceClient(conn)
	}
	return conns, clients, nil
}

func closeV2CRPCConnections(conns []*grpc.ClientConn) {
	for i := 0; i < len(conns); i++ {
		// TODO: handle errors in closing connections?
		_ = conns[i].Close()
	}
}

func extractHeader(key string, md metadata.MD) string {
	values, ok := md[key]
	if ok {
		if len(values) > 0 {
			// note if there are more than one elements we just return the first one
			return values[0]
		}
	}
	return ""
}

func extractModelNamesFromHeaders(ctx context.Context) (string, string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		internalModelName := extractHeader(resources.SeldonInternalModelHeader, md)
		externalModelName := extractHeader(resources.SeldonModelHeader, md)
		return internalModelName, externalModelName, internalModelName != "" && externalModelName != ""
	}
	return "", "", false
}

func getGrpcErrCode(err error) codes.Code {
	if err != nil {
		if e, ok := status.FromError(err); ok {
			return e.Code()
		}
	}
	return codes.Unknown
}

func retryForLazyReload(err error) bool {
	// we do lazy load in case of 404 and unavailable (triton), the idea being that if ml server restarts, state with agent is inconsistent.
	return getGrpcErrCode(err) == codes.NotFound || getGrpcErrCode(err) == codes.Unavailable
}
