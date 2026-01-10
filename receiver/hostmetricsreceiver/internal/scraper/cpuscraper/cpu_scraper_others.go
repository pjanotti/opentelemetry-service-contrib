// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/cpuucal"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/cpumetadata"
)

func (s *cpuScraper) recordCPUTimeStateDataPoints(now pcommon.Timestamp, cpuTime cpu.TimesStat) {
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.User, cpuTime.CPU, cpumetadata.AttributeStateUser)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.System, cpuTime.CPU, cpumetadata.AttributeStateSystem)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Idle, cpuTime.CPU, cpumetadata.AttributeStateIdle)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Irq, cpuTime.CPU, cpumetadata.AttributeStateInterrupt)
}

func (s *cpuScraper) recordCPUUtilization(now pcommon.Timestamp, cpuUtilization cpuucal.CPUUtilization) {
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.User, cpuUtilization.CPU, cpumetadata.AttributeStateUser)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.System, cpuUtilization.CPU, cpumetadata.AttributeStateSystem)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Idle, cpuUtilization.CPU, cpumetadata.AttributeStateIdle)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Irq, cpuUtilization.CPU, cpumetadata.AttributeStateInterrupt)
}

func (*cpuScraper) getCPUInfo() ([]cpuInfo, error) {
	var cpuInfos []cpuInfo
	return cpuInfos, nil
}
