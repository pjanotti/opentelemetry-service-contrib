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

package tests

import (
	"fmt"
	"math"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector/testbed/testbed"
	scenarios "github.com/open-telemetry/opentelemetry-collector/testbed/tests"
)

const agentSignalFxReceiverPort = 4343
const mockBackEndPort = 4344

var testTargets = []struct {
	name         string
	senderFactory func(t *testing.T) testbed.DataSender
	receiverFactory func(t *testing.T) testbed.DataReceiver
	resourceSpec testbed.ResourceSpec
	testCaseOpts []testbed.TestCaseOption
}{
	{
		"OpenCensus",
		func(t *testing.T) testbed.DataSender {
			return testbed.NewOCMetricDataSender(testbed.GetAvailablePort(t))
		},
		func (t *testing.T) testbed.DataReceiver {
			return testbed.NewOCDataReceiver(testbed.GetAvailablePort(t))
		},
		testbed.ResourceSpec{
			ExpectedMaxCPU: math.MaxUint32,
			ExpectedMaxRAM: math.MaxUint32,
		},
		nil,
	},
	{
		"SignalFx",
		func(t *testing.T) testbed.DataSender {
			return NewSFxMetricDataSender(testbed.GetAvailablePort(t))
		},
		func (t *testing.T) testbed.DataReceiver {
			return NewSFxMetricsDataReceiver(testbed.GetAvailablePort(t))
		},
		testbed.ResourceSpec{
			ExpectedMaxCPU: math.MaxUint32,
			ExpectedMaxRAM: math.MaxUint32,
		},
		nil,
	},
	{
		"SFx-Gateway",
		func(t *testing.T) testbed.DataSender {
			return NewSFxMetricDataSender(agentSignalFxReceiverPort)
		},
		func (t *testing.T) testbed.DataReceiver {
			return NewSFxMetricsDataReceiver(mockBackEndPort)
		},
		testbed.ResourceSpec{
			ExpectedMaxCPU: math.MaxUint32,
			ExpectedMaxRAM: math.MaxUint32,
		},
		[]testbed.TestCaseOption{
			testbed.WithCustomAgent(&testbed.StartProcessParams{
				Name:    "Gateway",
				Cmd:     "/Users/pj/go/src/github.com/signalfx/gateway/gateway",
				CmdArgs: []string{"-configfile", "/Users/pj/go/src/github.com/signalfx/gateway/local-etc/gateway.conf"},
			}),
		},
	},
}

func TestM(t *testing.T) {
	threeAttribs := map[string]string{
		"dim0": "value0", "dim1": "value1", "dim2": "value2",
	}
	tests := []testbed.LoadOptions{
		{
			DataItemsPerSecond: 500,
			ItemsPerBatch: 100,
			Attributes: threeAttribs,
		},
		{
			DataItemsPerSecond: 1000,
			ItemsPerBatch: 100,
			Attributes: threeAttribs,
		},
		{
			DataItemsPerSecond: 5000,
			ItemsPerBatch: 100,
		},
		{
			DataItemsPerSecond: 5000,
			ItemsPerBatch: 100,
			Attributes: threeAttribs,
		},
		{
			DataItemsPerSecond: 1e4,
			ItemsPerBatch: 100,
		},
		{
			DataItemsPerSecond: 1e4,
			ItemsPerBatch: 100,
			Attributes: threeAttribs,
		},
	}
	args := []string{"--mem-ballast-size-mib", "50"}
	for _, test := range tests {
		testName := fmt.Sprintf(
			"DPS(%v)_B(%v)_L(%v)",
			test.DataItemsPerSecond,
			test.ItemsPerBatch,
			len(test.Attributes))
		for _, target := range testTargets {
			t.Run(testName + "/" + target.name, func(t *testing.T) {
				scenarios.Scenario10kItemsPerSecond(
					t,
					args,
					target.senderFactory(t),
					target.receiverFactory(t),
					test,
					target.resourceSpec,
					target.testCaseOpts...,
				)
			})
		}
	}
}

func TestTm(t *testing.T) {
	fifteenAttribs := map[string]string{
		"dim0": "value0", "dim1": "value1", "dim2": "value2", "dim3": "value3",
		"dim4": "value4", "dim5": "value5", "dim6": "value6", "dim7": "value7",
		"dim8": "value8", "dim9": "value9", "dimA": "valueA", "dimB": "valueB",
		"dimC": "valueC", "dimD": "valueD", "dimE": "valueE", "dimF": "valueF",
	}
	tests := []testbed.LoadOptions{
		{
			DataItemsPerSecond: 1e4,
			ItemsPerBatch: 50,
			Attributes: fifteenAttribs,
		},
		{
			DataItemsPerSecond: 1e4,
			ItemsPerBatch: 5000,
			Attributes: fifteenAttribs,
		},
		{
			DataItemsPerSecond: 2e4,
			ItemsPerBatch: 5000,
			Attributes: fifteenAttribs,
		},
	}
	args := []string{"--mem-ballast-size-mib", "50"}
	for _, test := range tests {
		testName := fmt.Sprintf(
			"DPS(%v)_B(%v)_L(%v)",
			test.DataItemsPerSecond,
			test.ItemsPerBatch,
			len(test.Attributes))
		for _, target := range testTargets {
			t.Run(testName + "/" + target.name, func(t *testing.T) {
				scenarios.Scenario10kItemsPerSecond(
					t,
					args,
					target.senderFactory(t),
					target.receiverFactory(t),
					test,
					target.resourceSpec,
					target.testCaseOpts...,
				)
			})
		}
	}
}
