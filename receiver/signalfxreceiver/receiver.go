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
	"sync"

	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/oterr"
	"github.com/open-telemetry/opentelemetry-collector/receiver"
)

var (
	errNilNextConsumer = errors.New("nil nextConsumer")
)

var _ receiver.MetricsReceiver = (*sfxReceiver)(nil)

// sfxReceiver implements the receiver.TraceReceiver for Zipkin Scribe protocol.
type sfxReceiver struct {
	sync.Mutex

	startOnce sync.Once
	stopOnce  sync.Once
}

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
	config *Config,
	consumer consumer.TraceConsumer) (*sfxReceiver, error) {

	if consumer == nil {
		return nil, errNilNextConsumer
	}

	r := &sfxReceiver{}
	return r, nil
}
