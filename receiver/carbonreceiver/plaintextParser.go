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
	"fmt"
	"strconv"
	"strings"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Converts a line of https://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol,
// treating tags per spec at https://graphite.readthedocs.io/en/latest/tags.html#carbon.
type plaintextParser struct {}

var _ (parser) = (*plaintextParser)(nil)

func (p plaintextParser) Parse(line string) (*metricspb.Metric, error) {
	parts := strings.SplitN(line," ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid carbon metric [%s]", line)
	}

	path := parts[0]
	valueStr := parts[1]
	timestampStr := parts[2]

	metricName, tagMap, err := p.extractNameAndTags(path)
	if err != nil {
		return nil, fmt.Errorf("invalid carbon metric [%s]: %v", line, err)
	}

	unixTime, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid carbon metric time [%s]: %v", line, err)
	}

	var metricType metricspb.MetricDescriptor_Type
	point := metricspb.Point{
		Timestamp: convertUnixSec(unixTime),
	}
	intVal, err := strconv.ParseInt(valueStr, 10, 64)
	if err == nil {
		metricType = metricspb.MetricDescriptor_GAUGE_INT64
		point.Value = &metricspb.Point_Int64Value{Int64Value: intVal}
	} else {
		dblVal, err := strconv.ParseFloat(valueStr, 64);
		if  err != nil {
			return nil, fmt.Errorf("invalid carbon metric value [%s]: %v", line, err)
		}
		metricType = metricspb.MetricDescriptor_GAUGE_DOUBLE
		point.Value = &metricspb.Point_DoubleValue{DoubleValue: dblVal}
	}

	labelKeys, labelValues := p.buildLabelKeysAndValues(tagMap)
	descriptor := buildDescriptor(metricName, metricType, labelKeys)
	ts := metricspb.TimeSeries{
		// TODO: StartTimestamp can be set if each cumulative time series are
		//  	tracked but right now it is not clear if it brings benefits.
		//		Perhaps as an option so cost is "pay for play".
		LabelValues: labelValues,
		Points:      []*metricspb.Point{&point},
	}
	metric := metricspb.Metric{
		MetricDescriptor: descriptor,
		Timeseries:       []*metricspb.TimeSeries{&ts},
	}

	return &metric, nil
}

func (p *plaintextParser) extractNameAndTags(path string) (string, map[string]string, error) {
	// TODO: just skeleton for now
	return path, nil, nil
}

func (p *plaintextParser) buildLabelKeysAndValues(tagMap map[string]string) ([]*metricspb.LabelKey, []*metricspb.LabelValue) {
	// TODO: not generating anything yet.
	return nil, nil
}

func buildDescriptor(
	name string,
	metricType metricspb.MetricDescriptor_Type,
	labelKeys []*metricspb.LabelKey,
) *metricspb.MetricDescriptor {

	descriptor := &metricspb.MetricDescriptor{
		Name: name,
		// Description: no value to go here
		// Unit:        no value to go here
		Type:      metricType,
		LabelKeys: labelKeys,
	}

	return descriptor
}

func convertUnixSec(sec int64) *timestamp.Timestamp {
	if sec == 0 {
		return nil
	}

	ts := &timestamp.Timestamp{
		Seconds: sec,
	}
	return ts
}
