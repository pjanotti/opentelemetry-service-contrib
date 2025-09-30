// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows && !arm64

package iisreceiver

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/filter"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/scraperinttest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
)

func TestIntegration(t *testing.T) {
	if !isIISInstalled(t) {
		t.Skip("IIS is not installed, skipping integration test")
	}

	scraperinttest.NewIntegrationTest(
		NewFactory(),
		scraperinttest.WithCustomConfig(
			func(_ *testing.T, cfg component.Config, _ *scraperinttest.ContainerInfo) {
				rCfg := cfg.(*Config)
				rCfg.CollectionInterval = 100 * time.Millisecond
				rCfg.MetricsBuilderConfig.ResourceAttributes.IisSite.MetricsInclude = []filter.Config{{Strict: "Default Web Site"}}
				rCfg.ResourceAttributes.IisApplicationPool.MetricsInclude = []filter.Config{{Strict: "DefaultAppPool"}}
			},
		),
		scraperinttest.WithCompareOptions(
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreTimestamp()),
	).Run(t)
}

func isIISInstalled(t *testing.T) bool {
	// Use the Windows API directly so full administrative privileges are not required.
	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	require.NoError(t, err)
	// Ownership of the handle is transferred to the Mgr struct
	scm := &mgr.Mgr{Handle: handle}
	defer func() {
		require.NoError(t, scm.Disconnect())
	}()

	const iisService = "W3SVC" // World Wide Web Publishing Service
	service, err := scm.OpenService(iisService)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return false
	}

	require.NoError(t, service.Close())
	return true
}
