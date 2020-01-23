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

package carbonreceiver

import (
	"errors"
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"sync"

	"github.com/open-telemetry/opentelemetry-collector/component"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/oterr"
	"github.com/open-telemetry/opentelemetry-collector/receiver"
	"go.uber.org/zap"
)

var (
	errNilNextConsumer = errors.New("nil nextConsumer")
	errEmptyEndpoint   = errors.New("empty endpoint")
)

// carbonreceiver implements a receiver.MetricsReceiver for Carbon plaintext, aka "line", protocol.
// see https://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol.
type carbonReceiver struct {
	sync.Mutex
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.MetricsConsumer

	parser parser

	startOnce sync.Once
	stopOnce  sync.Once
}

var _ receiver.MetricsReceiver = (*carbonReceiver)(nil)

type parser interface {
	Parse(line string) (*metricspb.Metric, error)
}

func (r *carbonReceiver) MetricsSource() string {
	return "Carbon"
}

// New creates the Carbon receiver with the given configuration.
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

	r := &carbonReceiver{
		logger:       logger,
		config:       &config,
		nextConsumer: nextConsumer,
	}

	return r, nil
}

// Start tells the receiver to start its processing.
// By convention the consumer of the received data is set when the receiver
// instance is created.
func (r *carbonReceiver) Start(host component.Host) error {
	r.Lock()
	defer r.Unlock()

	err := oterr.ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil

		go func() {
			// TODO: Start either TCP or UDP listener.
		}()
	})

	return err
}

// Shutdown tells the receiver that should stop reception,
// giving it a chance to perform any necessary clean-up.
func (r *carbonReceiver) Shutdown() error {
	r.Lock()
	defer r.Unlock()

	err := oterr.ErrAlreadyStopped
	r.stopOnce.Do(func() {
		// TODO: Stop listener.
		err = nil
	})
	return err
}
