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
	"github.com/open-telemetry/opentelemetry-collector/config/configmodels"
	"runtime"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector/component"
	"github.com/open-telemetry/opentelemetry-collector/consumer"
	"github.com/open-telemetry/opentelemetry-collector/exporter/exportertest"
	"github.com/open-telemetry/opentelemetry-collector/oterr"
	"github.com/open-telemetry/opentelemetry-collector/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_carbonreceiver_New(t *testing.T) {
	defaultConfig := (&Factory{}).CreateDefaultConfig().(*Config)
	type args struct {
		config       Config
		nextConsumer consumer.MetricsConsumer
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "default_config",
			args: args{
				config: *defaultConfig,
				nextConsumer: new(exportertest.SinkMetricsExporter),
			},
		},
		{
			name: "nil_nextConsumer",
			args: args{
				config: *defaultConfig,
			},
			wantErr: errNilNextConsumer,
		},
		{
			name: "empty_endpoint",
			args: args{
				config: Config{
					ReceiverSettings: configmodels.ReceiverSettings{},
				},
				nextConsumer: new(exportertest.SinkMetricsExporter),
			},
			wantErr: errEmptyEndpoint,
		},
		// TODO: invalid transport.
		// TODO: invalid TCP idle timeout.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(zap.NewNop(), tt.args.config, tt.args.nextConsumer)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.NotNil(t, got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func Test_carbonreceiver_EndToEnd(t *testing.T) {
	addr := testutils.GetAvailableLocalAddress(t)
	cfg := (&Factory{}).CreateDefaultConfig().(*Config)
	cfg.Endpoint = addr
	sink := new(exportertest.SinkMetricsExporter)
	r, err := New(zap.NewNop(), *cfg, sink)
	require.NoError(t, err)

	mh := component.NewMockHost()
	err = r.Start(mh)
	require.NoError(t, err)
	runtime.Gosched()
	defer r.Shutdown()
	require.Equal(t, oterr.ErrAlreadyStarted, r.Start(mh))

	//unixSecs := int64(1574092046)
	//unixNSecs := int64(11 * time.Millisecond)
	//tsUnix := time.Unix(unixSecs, unixNSecs)
	//doubleVal := 1234.5678
	//doublePt := metricstestutils.Double(tsUnix, doubleVal)
	//int64Val := int64(123)
	//int64Pt := &metricspb.Point{
	//	Timestamp: metricstestutils.Timestamp(tsUnix),
	//	Value:     &metricspb.Point_Int64Value{Int64Value: int64Val},
	//}
	//want := consumerdata.MetricsData{
	//	Metrics: []*metricspb.Metric{
	//		metricstestutils.Gauge("gauge_double_with_dims", nil, metricstestutils.Timeseries(tsUnix, nil, doublePt)),
	//		metricstestutils.GaugeInt("gauge_int_with_dims", nil, metricstestutils.Timeseries(tsUnix, nil, int64Pt)),
	//		metricstestutils.Cumulative("cumulative_double_with_dims", nil, metricstestutils.Timeseries(tsUnix, nil, doublePt)),
	//		metricstestutils.CumulativeInt("cumulative_int_with_dims", nil, metricstestutils.Timeseries(tsUnix, nil, int64Pt)),
	//	},
	//}
	//
	//expCfg := &signalfxexporter.Config{
	//	URL: "http://" + addr + "/v2/datapoint",
	//}
	//exp, err := signalfxexporter.New(expCfg, zap.NewNop())
	//require.NoError(t, err)
	//require.NoError(t, exp.Start(mh))
	//defer exp.Shutdown()
	//require.NoError(t, exp.ConsumeMetricsData(context.Background(), want))
	//// Description, unit and start time are expected to be dropped during conversions.
	//for _, metric := range want.Metrics {
	//	metric.MetricDescriptor.Description = ""
	//	metric.MetricDescriptor.Unit = ""
	//	for _, ts := range metric.Timeseries {
	//		ts.StartTimestamp = nil
	//	}
	//}
	//
	//got := sink.AllMetrics()
	//require.Equal(t, 1, len(got))
	//assert.Equal(t, want, got[0])

	assert.NoError(t, r.Shutdown())
	assert.Equal(t, oterr.ErrAlreadyStopped, r.Shutdown())
}
