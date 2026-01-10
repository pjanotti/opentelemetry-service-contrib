// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v4/cpu"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/scraper/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/cpumetadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/ucal"
)

func (s *cpuScraper) recordCPUTimeStateDataPoints(now pcommon.Timestamp, cpuTime cpu.TimesStat) {
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.User, cpuTime.CPU, cpumetadata.AttributeStateUser)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.System, cpuTime.CPU, cpumetadata.AttributeStateSystem)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Idle, cpuTime.CPU, cpumetadata.AttributeStateIdle)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Irq, cpuTime.CPU, cpumetadata.AttributeStateInterrupt)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Nice, cpuTime.CPU, cpumetadata.AttributeStateNice)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Softirq, cpuTime.CPU, cpumetadata.AttributeStateSoftirq)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Steal, cpuTime.CPU, cpumetadata.AttributeStateSteal)
	s.mb.RecordSystemCPUTimeDataPoint(now, cpuTime.Iowait, cpuTime.CPU, cpumetadata.AttributeStateWait)
}

func (s *cpuScraper) recordCPUUtilization(now pcommon.Timestamp, cpuUtilization ucal.CPUUtilization) {
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.User, cpuUtilization.CPU, cpumetadata.AttributeStateUser)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.System, cpuUtilization.CPU, cpumetadata.AttributeStateSystem)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Idle, cpuUtilization.CPU, cpumetadata.AttributeStateIdle)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Irq, cpuUtilization.CPU, cpumetadata.AttributeStateInterrupt)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Nice, cpuUtilization.CPU, cpumetadata.AttributeStateNice)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Softirq, cpuUtilization.CPU, cpumetadata.AttributeStateSoftirq)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Steal, cpuUtilization.CPU, cpumetadata.AttributeStateSteal)
	s.mb.RecordSystemCPUUtilizationDataPoint(now, cpuUtilization.Iowait, cpuUtilization.CPU, cpumetadata.AttributeStateWait)
}

func (*cpuScraper) getCPUInfo() ([]cpuInfo, error) {
	var cpuInfos []cpuInfo
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, scrapererror.NewPartialScrapeError(err, metricsLen)
	}
	cInf, err := fs.CPUInfo()
	if err != nil {
		return nil, scrapererror.NewPartialScrapeError(err, metricsLen)
	}
	for i := range cInf {
		cInfo := &cInf[i]
		c := cpuInfo{
			frequency: cInfo.CPUMHz,
			processor: cInfo.Processor,
		}
		cpuInfos = append(cpuInfos, c)
	}
	return cpuInfos, nil
}
