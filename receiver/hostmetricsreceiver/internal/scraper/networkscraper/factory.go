// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package networkscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper/internal/networkmetadata"
)

// NewFactory for Network scraper.
func NewFactory() scraper.Factory {
	return scraper.NewFactory(networkmetadata.Type, createDefaultConfig, scraper.WithMetrics(createMetricsScraper, networkmetadata.MetricsStability))
}

// createDefaultConfig creates the default configuration for the Scraper.
func createDefaultConfig() component.Config {
	return &Config{
		MetricsBuilderConfig: networkmetadata.DefaultMetricsBuilderConfig(),
	}
}

// createMetricsScraper creates a scraper based on provided config.
func createMetricsScraper(
	ctx context.Context,
	settings scraper.Settings,
	config component.Config,
) (scraper.Metrics, error) {
	cfg := config.(*Config)
	s, err := newNetworkScraper(ctx, settings, cfg)
	if err != nil {
		return nil, err
	}

	return scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
	)
}
