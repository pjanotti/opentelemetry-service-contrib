// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nfsscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/nfsscraper"

import (
	"context"
	"errors"
	"runtime"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/nfsscraper/internal/nfsmetadata"
)

var (
	supportedOS      = runtime.GOOS == "linux"
	errUnsupportedOS = errors.New("the nfs scraper is only available on Linux")
)

// NewFactory for NFS scraper.
func NewFactory() scraper.Factory {
	return scraper.NewFactory(nfsmetadata.Type, createDefaultConfig, scraper.WithMetrics(createMetricsScraper, nfsmetadata.MetricsStability))
}

// createDefaultConfig creates the default configuration for the Scraper.
func createDefaultConfig() component.Config {
	return &Config{
		MetricsBuilderConfig: nfsmetadata.DefaultMetricsBuilderConfig(),
	}
}

// createMetricsScraper creates a resource scraper based on provided config.
func createMetricsScraper(
	_ context.Context,
	settings scraper.Settings,
	cfg component.Config,
) (scraper.Metrics, error) {
	if !supportedOS {
		return nil, errUnsupportedOS
	}

	nfsScraper := newNfsScraper(settings, cfg.(*Config))

	return scraper.NewMetrics(
		nfsScraper.scrape,
		scraper.WithStart(nfsScraper.start),
	)
}
