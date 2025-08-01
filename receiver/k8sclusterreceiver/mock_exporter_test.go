// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package k8sclusterreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"

	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
)

type mockExporter struct{}

func (mockExporter) Start(context.Context, component.Host) error {
	return nil
}

func (mockExporter) Shutdown(context.Context) error {
	return nil
}

var _ component.Component = (*mockExporterWithK8sMetadata)(nil)

type mockExporterWithK8sMetadata struct {
	*consumertest.MetricsSink
}

func (mockExporterWithK8sMetadata) Start(context.Context, component.Host) error {
	return nil
}

func (mockExporterWithK8sMetadata) Shutdown(context.Context) error {
	return nil
}

func (mockExporterWithK8sMetadata) ConsumeMetadata([]*metadata.MetadataUpdate) error {
	consumeMetadataInvocation()
	return nil
}
