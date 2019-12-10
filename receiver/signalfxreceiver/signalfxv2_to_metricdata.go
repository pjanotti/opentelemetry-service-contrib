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
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/open-telemetry/opentelemetry-collector/consumer/consumerdata"
	sfxpb "github.com/signalfx/com_signalfx_metrics_protobuf"
	"go.uber.org/zap"
)

func metricDataToSignalFxV2(
	logger *zap.Logger,
	sfxDataPoints []*sfxpb.DataPoint,
) (*consumerdata.MetricsData, int, error) {

	// TODO: not optimized at all, basically regenerating everything for each
	// 	data point.

	md := &consumerdata.MetricsData{
		Node:     nil, // TODO: Add this instance itself? Leave it to a future Processor?
		Resource: nil, // TODO: likely do nothing since Resources are injected via processors
	}
	metrics := make([]*metricspb.Metric, 0, len(sfxDataPoints))
	for _, sfxDataPoint := range sfxDataPoints {
		labelKeys, labelValues := buildLabelKeysAndValues(sfxDataPoint.Dimensions)
		descriptor := buildDescriptor(sfxDataPoint, labelKeys) // TODO: add way to account for dropped timeseries
		point := buildPoint(sfxDataPoint)
		ts := &metricspb.TimeSeries{
			StartTimestamp: nil, // TODO: for cumulative this is relevant but doesn't seem to have any info on data point
			LabelValues:    labelValues,
			Points:         []*metricspb.Point{point},
		}
		metric := &metricspb.Metric{
			MetricDescriptor: descriptor,
			Timeseries:       []*metricspb.TimeSeries{ts},
		}
		metrics = append(metrics, metric)
	}

	md.Metrics = metrics
	return md, 0, nil
}

func buildPoint(sfxDataPoint *sfxpb.DataPoint) *metricspb.Point {
	if sfxDataPoint.Value == nil {
		// TODO: handle nil Datum case
		panic("TODO: handle nil Datum case")
	}

	p := &metricspb.Point{
		Timestamp: convertTimestamp(sfxDataPoint.GetTimestamp()),
	}

	switch {
	case sfxDataPoint.Value.IntValue != nil:
		p.Value = &metricspb.Point_Int64Value{Int64Value: *sfxDataPoint.Value.IntValue}
	case sfxDataPoint.Value.DoubleValue != nil:
		p.Value = &metricspb.Point_DoubleValue{DoubleValue: *sfxDataPoint.Value.DoubleValue}
	case sfxDataPoint.Value.StrValue != nil:
		// TODO: Ensure that this is properly handled.
		panic("TODO: sfxDataPoint.Value.StrValue != nil")
	default:
		// TODO: handle unexpected case
		panic("TODO: unknown datum type")
	}

	return p
}

func convertTimestamp(msec int64) *timestamp.Timestamp {
	if msec == 0 {
		return nil
	}

	ts := &timestamp.Timestamp{
		Seconds: msec / 1e3,
		Nanos:   (msec % 1e3) * 1e3,
	}
	return ts
}

func buildDescriptor(
	sfxDataPoint *sfxpb.DataPoint,
	labelKeys []*metricspb.LabelKey,
) *metricspb.MetricDescriptor {

	// TODO: Initially create this every single one and do not worry about
	// caching.
	descriptor := &metricspb.MetricDescriptor{
		Name:        *sfxDataPoint.Metric,
		Description: "", // TODO: Anything to go here?
		Unit:        "", // TODO: Anything to go here?
		Type:        convertType(sfxDataPoint),
		LabelKeys:   labelKeys,
	}

	return descriptor
}

func convertType(sfxDataPoint *sfxpb.DataPoint) metricspb.MetricDescriptor_Type {
	var descType metricspb.MetricDescriptor_Type

	// Combine metric type with the actual data point type
	sfxMetricType := *sfxDataPoint.MetricType
	sfxDatum := sfxDataPoint.Value
	if sfxDatum.StrValue != nil {
		// TODO: count as dropped metric?
		panic("TODO: not supported")
	}

	switch sfxMetricType {
	case sfxpb.MetricType_GAUGE:
		// Numerical: Periodic, instantaneous measurement of some state.
		descType = metricspb.MetricDescriptor_GAUGE_DOUBLE
		if sfxDatum.IntValue != nil {
			descType = metricspb.MetricDescriptor_GAUGE_INT64
		}

	case sfxpb.MetricType_COUNTER:
		// Numerical: Count of occurrences. Generally non-negative integers.
		fallthrough

	case sfxpb.MetricType_CUMULATIVE_COUNTER:
		// Tracks a value that increases over time, where only the difference is important.
		descType = metricspb.MetricDescriptor_CUMULATIVE_DOUBLE
		if sfxDatum.IntValue != nil {
			descType = metricspb.MetricDescriptor_CUMULATIVE_INT64
		}

	case sfxpb.MetricType_ENUM:
		// String: Used for non-continuous quantities (that is, measurements where there is a fixed
		// set of meaningful values). This is essentially a special case of gauge.
		// TODO: any way in OC to support this? Likely log and count as dropped timeseries
		panic("TODO: any way to support this")

	default:
		// TODO: handle any metric type
		panic("TODO: handle unknown sfxMetricType")
	}

	return descType
}

func buildLabelKeysAndValues(
	dimensions []*sfxpb.Dimension,
) ([]*metricspb.LabelKey, []*metricspb.LabelValue) {
	keys := make([]*metricspb.LabelKey, 0, len(dimensions))
	values := make([]*metricspb.LabelValue, 0, len(dimensions))
	for _, dim := range dimensions {
		lk := &metricspb.LabelKey{Key: *dim.Key}
		keys = append(keys, lk)

		lv := &metricspb.LabelValue{}
		if dim.Value != nil {
			lv.Value = *dim.Value
			lv.HasValue = true
		}
		values = append(values, lv)
	}
	return keys, values
}
