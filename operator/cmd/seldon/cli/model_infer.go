package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	stickySessionFlag = "sticky-session"
)

func createModelInfer() *cobra.Command {
	cmdModelInfer := &cobra.Command{
		Use:   "infer <modelName> (data)",
		Short: "run inference on a model",
		Long:  `call a model with a given input and get a prediction`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inferenceHost, err := cmd.Flags().GetString(inferenceHostFlag)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(fileFlag)
			if err != nil {
				return err
			}
			showRequest, err := cmd.Flags().GetBool(showRequestFlag)
			if err != nil {
				return err
			}
			showResponse, err := cmd.Flags().GetBool(showResponseFlag)
			if err != nil {
				return err
			}
			stickySesion, err := cmd.Flags().GetBool(stickySessionFlag)
			if err != nil {
				return err
			}
			showHeaders, err := cmd.Flags().GetBool(showHeadersFlag)
			if err != nil {
				return err
			}
			inferMode, err := cmd.Flags().GetString(inferenceModeFlag)
			if err != nil {
				return err
			}
			inferenceClient, err := cli.NewInferenceClient(inferenceHost)
			if err != nil {
				return err
			}
			iterations, err := cmd.Flags().GetInt(inferenceIterationsFlag)
			if err != nil {
				return err
			}
			headers, err := cmd.Flags().GetStringArray(addHeaderFlag)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(authorityFlag)
			if err != nil {
				return err
			}
			modelName := args[0]
			// Get inference data
			var data []byte
			if len(args) > 1 {
				data = []byte(args[1])
			} else if filename != "" {
				data = loadFile(filename)
			} else {
				return fmt.Errorf("required inline data or from file with -f <file-path>")
			}
			err = inferenceClient.Infer(modelName, inferMode, data, showRequest, showResponse, iterations, cli.InferModel, showHeaders, headers, authority, stickySesion)
			return err
		},
	}

	cmdModelInfer.Flags().StringP(fileFlag, "f", "", "inference payload file")
	cmdModelInfer.Flags().BoolP(stickySessionFlag, "s", false, "use sticky session from last infer (only works with inference to experiments)")
	cmdModelInfer.Flags().String(inferenceHostFlag, env.GetString(EnvInfer, DefaultInferHost), "seldon inference host")
	cmdModelInfer.Flags().String(inferenceModeFlag, "rest", "inference mode rest or grpc")
	cmdModelInfer.Flags().IntP(inferenceIterationsFlag, "i", 1, "inference iterations")
	cmdModelInfer.Flags().Bool(showHeadersFlag, false, "show headers")
	cmdModelInfer.Flags().StringArray(addHeaderFlag, []string{}, fmt.Sprintf("add header key%svalue", cli.HeaderSeparator))
	cmdModelInfer.Flags().String(authorityFlag, "", "authority (HTTP/2) or virtual host (HTTP/1)")

	return cmdModelInfer
}
