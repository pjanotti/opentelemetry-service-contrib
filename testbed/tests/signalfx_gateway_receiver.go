// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
	"github.com/open-telemetry/opentelemetry-collector/receiver"
	"github.com/open-telemetry/opentelemetry-collector/testbed/testbed"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver"
)

const (
	// SFxGatewayMetricsPort is the port hard coded in the Gateway config that
	// is expected to receive the v2/datapoint metrics.
	SFxGatewayMetricsPort = 4343
)

// SFxGatewayDataReceiver implements a way to call the SignalFx Gateway and to
// use it as a receiver for performance tests.
type SFxGatewayDataReceiver struct {
	testbed.DataReceiverBase
	receiver receiver.MetricsReceiver
}

// Ensure SFxGatewayDataReceiver implements DataReceiver.
var _ testbed.DataReceiver = (*SFxGatewayDataReceiver)(nil)

// NewSFxGatewayDataReceiver creates a new SFxGatewayDataReceiver that will listen on the
// specified port after Start is called.
func NewSFxGatewayDataReceiver(port int) *SFxGatewayDataReceiver {
	// TODO: For now the port is hardcoded so panic if not using the expected one.
	if port != SFxGatewayMetricsPort {
		panic("The port is hard-coded on the gateway configuration, it must be 4343.")
	}
	return &SFxGatewayDataReceiver{DataReceiverBase: testbed.DataReceiverBase{Port: SFxGatewayMetricsPort}}
}

// Start the receiver.
func (or *SFxGatewayDataReceiver) Start(tc *testbed.MockTraceConsumer, mc *testbed.MockMetricConsumer) error {
	addr := fmt.Sprintf("localhost:%d", or.Port)
	config := signalfxreceiver.Config{
		ReceiverSettings: configmodels.ReceiverSettings{Endpoint: addr},
	}
	var err error
	or.receiver, err = signalfxreceiver.New(zap.L(), config, mc)
	if err != nil {
		return err
	}

	return or.receiver.StartMetricsReception(or)
}

// Stop the receiver.
func (or *SFxGatewayDataReceiver) Stop() {
	or.receiver.StopMetricsReception()
}

// GenConfigYAMLStr returns exporter config for the agent.
func (or *SFxGatewayDataReceiver) GenConfigYAMLStr() string {
	// Note that this generates an exporter config for agent.
	return fmt.Sprintf(`
  signalfx:
    url: "http://localhost:%d/v2/datapoint"`, or.Port)
}

// ProtocolName returns protocol name as it is specified in Collector config.
func (or *SFxGatewayDataReceiver) ProtocolName() string {
	return "signalfx"
}

type gatewayreceiver struct {}

