// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package processesscraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/scraper/scrapertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper/internal/processesmetadata"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.IsType(t, &Config{}, cfg)
}

func TestCreateMetrics(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}

	scraper, err := factory.CreateMetrics(t.Context(), scrapertest.NewNopSettings(processesmetadata.Type), cfg)

	assert.NoError(t, err)
	assert.NotNil(t, scraper)
}
