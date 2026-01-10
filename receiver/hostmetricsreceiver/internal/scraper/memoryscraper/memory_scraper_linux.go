// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package memoryscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/memoryscraper"

import (
	"github.com/shirou/gopsutil/v4/mem"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/memoryscraper/internal/memorymetadata"
)

var useMemAvailable = featuregate.GlobalRegistry().MustRegister(
	"receiver.hostmetricsreceiver.UseLinuxMemAvailable",
	featuregate.StageBeta,
	featuregate.WithRegisterFromVersion("v0.137.0"),
	featuregate.WithRegisterDescription("When enabled, the used value for the system.memory.usage and system.memory.utilization metrics will be based on the Linux kernel’s MemAvailable statistic instead of MemFree, Buffers, and Cached."),
	featuregate.WithRegisterReferenceURL("https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/42221"),
)

func (s *memoryScraper) recordMemoryUsageMetric(now pcommon.Timestamp, memInfo *mem.VirtualMemoryStat) {
	// TODO: rely on memInfo.Used value once https://github.com/shirou/gopsutil/pull/1882 is released
	// gopsutil formula: https://github.com/shirou/gopsutil/pull/1882/files#diff-5af8322731595fb792b48f3c38f31ddb24f596cf11a74a9c37b19734597baef6R321
	if useMemAvailable.IsEnabled() {
		s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Total-memInfo.Available), memorymetadata.AttributeStateUsed)
	} else {
		// gopsutil legacy "Used" memory formula = Total - Free - Buffers - Cache
		s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Total-memInfo.Free-memInfo.Buffers-memInfo.Cached), memorymetadata.AttributeStateUsed)
	}
	s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Free), memorymetadata.AttributeStateFree)
	s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Buffers), memorymetadata.AttributeStateBuffered)
	s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Cached), memorymetadata.AttributeStateCached)
	s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Sreclaimable), memorymetadata.AttributeStateSlabReclaimable)
	s.mb.RecordSystemMemoryUsageDataPoint(now, int64(memInfo.Sunreclaim), memorymetadata.AttributeStateSlabUnreclaimable)
}

func (s *memoryScraper) recordMemoryUtilizationMetric(now pcommon.Timestamp, memInfo *mem.VirtualMemoryStat) {
	// TODO: rely on memInfo.Used value once https://github.com/shirou/gopsutil/pull/1882 is released
	// gopsutil formula: https://github.com/shirou/gopsutil/pull/1882/files#diff-5af8322731595fb792b48f3c38f31ddb24f596cf11a74a9c37b19734597baef6R321
	if useMemAvailable.IsEnabled() {
		s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Total-memInfo.Available)/float64(memInfo.Total), memorymetadata.AttributeStateUsed)
	} else {
		// gopsutil legacy "Used" memory formula = Total - Free - Buffers - Cache
		s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Total-memInfo.Free-memInfo.Buffers-memInfo.Cached)/float64(memInfo.Total), memorymetadata.AttributeStateUsed)
	}
	s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Free)/float64(memInfo.Total), memorymetadata.AttributeStateFree)
	s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Buffers)/float64(memInfo.Total), memorymetadata.AttributeStateBuffered)
	s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Cached)/float64(memInfo.Total), memorymetadata.AttributeStateCached)
	s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Sreclaimable)/float64(memInfo.Total), memorymetadata.AttributeStateSlabReclaimable)
	s.mb.RecordSystemMemoryUtilizationDataPoint(now, float64(memInfo.Sunreclaim)/float64(memInfo.Total), memorymetadata.AttributeStateSlabUnreclaimable)
}

func (s *memoryScraper) recordLinuxMemoryAvailableMetric(now pcommon.Timestamp, memInfo *mem.VirtualMemoryStat) {
	s.mb.RecordSystemLinuxMemoryAvailableDataPoint(now, int64(memInfo.Available))
}

func (s *memoryScraper) recordLinuxMemoryDirtyMetric(now pcommon.Timestamp, memInfo *mem.VirtualMemoryStat) {
	// This value is collected from /proc/meminfo and converted from kB to bytes in gopsutil:
	// https://github.com/shirou/gopsutil/blob/d8750909ba41f2de9750c90a6d2074c68dfc677e/mem/mem_linux.go#L148
	s.mb.RecordSystemLinuxMemoryDirtyDataPoint(now, int64(memInfo.Dirty))
}

func (s *memoryScraper) recordSystemSpecificMetrics(now pcommon.Timestamp, memInfo *mem.VirtualMemoryStat) {
	s.recordLinuxMemoryAvailableMetric(now, memInfo)
	s.recordLinuxMemoryDirtyMetric(now, memInfo)
}
