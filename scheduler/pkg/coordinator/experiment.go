/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package coordinator

import (
	"context"
	"reflect"

	busV3 "github.com/mustafaturan/bus/v3"
	log "github.com/sirupsen/logrus"
)

func (h *EventHub) RegisterExperimentEventHandler(
	name string,
	queueSize int,
	logger log.FieldLogger,
	handle func(event ExperimentEventMsg),
) {
	events := make(chan ExperimentEventMsg, queueSize)
	h.addExperimentEventHandlerChannel(events)

	go func() {
		for e := range events {
			handle(e)
		}
	}()

	handler := h.newExperimentEventHandler(logger, events)
	h.bus.RegisterHandler(name, handler)
}

func (h *EventHub) newExperimentEventHandler(
	logger log.FieldLogger,
	events chan ExperimentEventMsg,
) busV3.Handler {
	handleExperimentEventMessage := func(_ context.Context, e busV3.Event) {
		l := logger.WithField("func", "handleExperimentEventMessage")
		l.Debugf("Received event on %s from %s (ID: %s, TxID: %s)", e.Topic, e.Source, e.ID, e.TxID)

		me, ok := e.Data.(ExperimentEventMsg)
		if !ok {
			l.Warnf(
				"Event (ID %s, TxID %s) on topic %s from %s is not a ExperimentEventMsg: %s",
				e.ID,
				e.TxID,
				e.Topic,
				e.Source,
				reflect.TypeOf(e.Data).String(),
			)
			return
		}

		h.lock.RLock()
		if h.closed {
			return
		}
		events <- me
		h.lock.RUnlock()
	}

	return busV3.Handler{
		Matcher: topicExperimentEvents,
		Handle:  handleExperimentEventMessage,
	}
}

func (h *EventHub) addExperimentEventHandlerChannel(c chan ExperimentEventMsg) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.experimentEventHandlerChannels = append(h.experimentEventHandlerChannels, c)
}

func (h *EventHub) PublishExperimentEvent(source string, event ExperimentEventMsg) {
	err := h.bus.EmitWithOpts(
		context.Background(),
		topicExperimentEvents,
		event,
		busV3.WithSource(source),
	)
	if err != nil {
		h.logger.WithError(err).Errorf(
			"unable to publish experiment event message from %s to %s",
			source,
			topicExperimentEvents,
		)
	}
}
