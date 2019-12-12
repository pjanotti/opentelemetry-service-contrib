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
	"compress/gzip"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/oterr"
	"github.com/open-telemetry/opentelemetry-collector/receiver"
	sfxpb "github.com/signalfx/com_signalfx_metrics_protobuf"
	"go.uber.org/zap"
)

const defaultServerTimeout = 20 * time.Second

var (
	errNilNextConsumer = errors.New("nil nextConsumer")
	errEmptyEndpoint   = errors.New("empty endpoint")
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

func (r *sfxReceiver) MetricsSource() string {
	const metricsSource string = "SignalFx"
	return metricsSource
}

// New creates the SignalFx receiver with the given configuration.
func New(
	logger *zap.Logger,
	config Config,
	nextConsumer consumer.MetricsConsumer,
) (receiver.MetricsReceiver, error) {

	if nextConsumer == nil {
		return nil, errNilNextConsumer
	}

	if config.Endpoint == "" {
		return nil, errEmptyEndpoint
	}

	// Handle config zero values.
	if config.Name() == "" {
		config.SetType(typeStr)
		config.SetName(typeStr)
	}

	r := &sfxReceiver{
		logger:       logger,
		config:       &config,
		nextConsumer: nextConsumer,
		server: &http.Server{
			Addr: config.Endpoint,
			// TODO: Evaluate what properties should be configurable, for now
			//		set some hard-coded values.
			ReadHeaderTimeout: defaultServerTimeout,
			WriteTimeout:      defaultServerTimeout,
		},
	}

	mux := mux.NewRouter()
	mux.HandleFunc("/v2/datapoint", r.handleReq)
	r.server.Handler = mux

	return r, nil
}

func (r *sfxReceiver) StartMetricsReception(host receiver.Host) error {
	r.Lock()
	defer r.Unlock()

	err := oterr.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil

		go func() {
			if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				host.ReportFatalError(err)
			}
		}()
	})

	return err
}

func (r *sfxReceiver) StopMetricsReception() error {
	r.Lock()
	defer r.Unlock()

	err := oterr.ErrAlreadyStopped
	r.stopOnce.Do(func() {
		err = r.server.Close()
	})
	return err
}

const (
	responseInvalidMethod      = "Only \"POST\" method is supported"
	responseInvalidContentType = "\"Content-Type\" must be \"application/x-protobuf\""
	responseInvalidEncoding    = "\"Content-Encoding\" must be \"gzip\" or empty"
	responseErrGzipReader      = "Error on gzip body"
	responseErrReadBody        = "Failed to read message body"
	responseErrUnmarshalBody   = "Failed to unmarshal message body"
	responseErrNextConsumer    = "Internal Server Error"
)

func (r *sfxReceiver) handleReq(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			responseInvalidMethod,
			nil)
		return
	}

	if req.Header.Get("Content-Type") != "application/x-protobuf" {
		r.writeResponse(
			resp,
			http.StatusUnsupportedMediaType,
			responseInvalidContentType,
			nil)
		return
	}

	encoding := req.Header.Get("Content-Encoding")
	if encoding != "" && encoding != "gzip" {
		r.writeResponse(
			resp,
			http.StatusUnsupportedMediaType,
			responseInvalidEncoding,
			nil)
		return
	}

	var err error
	bodyReader := req.Body
	if encoding == "gzip" {
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			r.writeResponse(
				resp,
				http.StatusBadRequest,
				responseErrGzipReader,
				err)
			return
		}
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			responseErrReadBody,
			err)
		return
	}

	msg := &sfxpb.DataPointUploadMessage{}
	if err := proto.Unmarshal(body, msg); err != nil {
		r.writeResponse(
			resp,
			http.StatusBadRequest,
			responseErrUnmarshalBody,
			err)
		return
	}

	if len(msg.Datapoints) == 0 {
		// TODO: add observability, perhaps should be considered error without
		//		data loss.
		return
	}

	md, _ := SignalFxV2ToMetricsData(r.logger, msg.Datapoints)

	err = r.nextConsumer.ConsumeMetricsData(req.Context(), *md)
	if err != nil {
		r.writeResponse(
			resp,
			http.StatusInternalServerError,
			responseErrNextConsumer,
			err)
		return
	}

	r.writeResponse(
		resp,
		http.StatusAccepted,
		"",
		nil)
}

func (r *sfxReceiver) writeResponse(
	resp http.ResponseWriter,
	httpStatusCode int,
	msg string,
	err error,
) {
	resp.WriteHeader(httpStatusCode)
	if msg != "" {
		if err == nil {
			r.logger.Debug(
				"Incorrect HTTP request",
				zap.String("msg", msg),
				zap.String("receiver", r.config.Name()))
		}
		_, writeErr := resp.Write([]byte(msg))
		if writeErr != nil {
			r.logger.Warn(
				"Error writing HTTP response message",
				zap.Error(writeErr),
				zap.String("receiver", r.config.Name()))
		}
	}
	if err != nil {
		r.logger.Warn(
			"Error processing HTTP request",
			zap.Error(err),
			zap.String("receiver", r.config.Name()))
	}
}
