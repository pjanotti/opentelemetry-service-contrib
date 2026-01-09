// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package interfacesscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver/internal/scraper/interfacesscraper"

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver/internal/connection"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver/internal/scraper/interfacesscraper/internal/interfacesmetadata"
)

// Config holds configuration for the interfaces scraper
type Config struct {
	interfacesmetadata.MetricsBuilderConfig `mapstructure:",squash"`
	Device                                  connection.DeviceConfig `mapstructure:"-"` // Passed from receiver config
}
