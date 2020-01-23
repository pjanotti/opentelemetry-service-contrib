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
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_plaintextParser_Parse(t *testing.T) {
	p := plaintextParser{}
	tests := []struct{
		line string
		want *metricspb.Metric
		wantErr error
	}{
		{
			line: "tst_int 1 9072000",
			want: buildMetric(
				metricspb.MetricDescriptor_GAUGE_INT64,
				"tst_int",
				nil,
				&metricspb.TimeSeries{
					Points: []*metricspb.Point{
						{
							Timestamp: &timestamp.Timestamp{Seconds: 9072000},
							Value: &metricspb.Point_Int64Value{1},
						},
					},
				},
			),
		},
		{
			line: "tst_dbl 3.14 9072050",
			want: buildMetric(
				metricspb.MetricDescriptor_GAUGE_DOUBLE,
				"tst_dbl",
				nil,
				&metricspb.TimeSeries{
					Points: []*metricspb.Point{
						{
							Timestamp: &timestamp.Timestamp{Seconds: 9072050},
							Value: &metricspb.Point_DoubleValue{3.14},
						},
					},
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func (t *testing.T){
			got, err := p.Parse(tt.line)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func buildMetric(ty metricspb.MetricDescriptor_Type, name string, keys []string, timeseries *metricspb.TimeSeries) *metricspb.Metric {
	return &metricspb.Metric{
		MetricDescriptor: &metricspb.MetricDescriptor{
			Name:        name,
			Type:        ty,
			LabelKeys:   nil,
		},
		Timeseries: []*metricspb.TimeSeries{timeseries},
	}
}

