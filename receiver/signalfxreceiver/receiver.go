// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signalfxreceiver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/oterr"
	"github.com/open-telemetry/opentelemetry-collector/receiver"
	sfxpb "github.com/signalfx/com_signalfx_metrics_protobuf"
	"go.uber.org/zap"
)

var (
	errNilNextConsumer = errors.New("nil nextConsumer")
)

// sfxReceiver implements the receiver.TraceReceiver for Zipkin Scribe protocol.
type sfxReceiver struct {
	sync.Mutex
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.MetricsConsumer
	server       *http.Server

	startOnce sync.Once
	stopOnce  sync.Once
}

var _ receiver.MetricsReceiver = (*sfxReceiver)(nil)
var _ http.Handler = (*sfxReceiver)(nil)

func (r *sfxReceiver) MetricsSource() string {
	const metricsSource string = "SignalFx"
	return metricsSource
}

func (r *sfxReceiver) StartMetricsReception(host receiver.Host) error {
	r.Lock()
	defer r.Unlock()

	err := oterr.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil

		go func() {
			host.ReportFatalError(errors.New("todo: not implemeted yet"))
		}()
	})

	return err
}

func (r *sfxReceiver) StopMetricsReception() error {
	r.Lock()
	defer r.Unlock()

	var err = oterr.ErrAlreadyStopped
	r.stopOnce.Do(func() {
		err = errors.New("todo: not implemented yet")
	})
	return err
}

// New creates the SignalFx receiver with the given configuration.
func New(
	logger *zap.Logger,
	config *Config,
	nextConsumer consumer.MetricsConsumer) (receiver.MetricsReceiver, error) {

	if nextConsumer == nil {
		return nil, errNilNextConsumer
	}

	r := &sfxReceiver{
		logger:       logger,
		config:       config,
		nextConsumer: nextConsumer,
		server: &http.Server{
			Addr: config.Endpoint,
			// TODO: What other properties to configure?
		},
	}
	r.server.Handler = r
	return r, nil
}

func (r *sfxReceiver) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			"Only \"POST\" method is supported")
		return
	}

	if req.Header.Get("Content-Type") != "application/x-protobuf" {
		r.writeResponse(
			resp,
			http.StatusUnsupportedMediaType,
			"\"Content-Type\" must be \"application/x-protobuf\"")
		return
	}

	encoding := req.Header.Get("Content-Encoding")
	if encoding != "" && encoding != "gzip" {
		r.writeResponse(
			resp,
			http.StatusUnsupportedMediaType,
			"\"Content-Encoding\" must be \"gzip\" or empty")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			"Failed to read message body")
	}

	if encoding == "gzip" {
		// TODO: decompress before unmarshall
	}

	var msg sfxpb.DataPointUploadMessage
	if err := proto.Unmarshal(body, msg); err != nil {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			"Failed to unmarshal message body")
	}

	md, _, err := metricDataToSignalFxV2(r.logger, msg.Datapoints)
	// TODO: add observability metrics
	if err != nil {
		// Assume that any error is for the whole request.
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			fmt.Sprintf(
				"Failed to convert SignalFx metric to internal format: %s",
				err.Error()))
	}

	err = r.nextConsumer.ConsumeMetricsData(req.Context(), *md)
	if err != nil {
		r.writeResponse(
			resp,
			http.StatusInternalServerError,
			err.Error())
	}
}

func (r *sfxReceiver) writeResponse(
	resp http.ResponseWriter,
	httpStatusCode int,
	msg string,
) {
	resp.WriteHeader(httpStatusCode)
	if msg != "" {
		_, err := resp.Write([]byte(msg))
		if err != nil {
			r.logger.Warn(
				"Error writing HTTP response message",
				zap.Error(err),
				zap.String("receiver", r.config.Name()))
		}
	}
}
