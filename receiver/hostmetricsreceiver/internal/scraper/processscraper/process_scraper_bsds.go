// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build darwin || freebsd

package processscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper"

import (
	"context"
	"regexp"

	"github.com/shirou/gopsutil/v4/cpu"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper/internal/processmetadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper/processucal"
)

func (s *processScraper) recordCPUTimeMetric(now pcommon.Timestamp, cpuTime *cpu.TimesStat) {
	s.mb.RecordProcessCPUTimeDataPoint(now, cpuTime.User, processmetadata.AttributeStateUser)
	s.mb.RecordProcessCPUTimeDataPoint(now, cpuTime.System, processmetadata.AttributeStateSystem)
	s.mb.RecordProcessCPUTimeDataPoint(now, cpuTime.Iowait, processmetadata.AttributeStateWait)
}

func (s *processScraper) recordCPUUtilization(now pcommon.Timestamp, cpuUtilization processucal.CPUUtilization) {
	s.mb.RecordProcessCPUUtilizationDataPoint(now, cpuUtilization.User, processmetadata.AttributeStateUser)
	s.mb.RecordProcessCPUUtilizationDataPoint(now, cpuUtilization.System, processmetadata.AttributeStateSystem)
	s.mb.RecordProcessCPUUtilizationDataPoint(now, cpuUtilization.Iowait, processmetadata.AttributeStateWait)
}

func getProcessName(ctx context.Context, proc processHandle, _ string) (string, error) {
	name, err := proc.NameWithContext(ctx)
	if err != nil {
		return "", err
	}

	return name, nil
}

func getProcessCgroup(_ context.Context, _ processHandle) (string, error) {
	return "", nil
}

func getProcessExecutable(ctx context.Context, proc processHandle) (string, error) {
	cmdline, err := proc.CmdlineWithContext(ctx)
	if err != nil {
		return "", err
	}
	regex := regexp.MustCompile(`^\S+`)
	exe := regex.FindString(cmdline)

	return exe, nil
}

func getProcessCommand(ctx context.Context, proc processHandle) (*commandMetadata, error) {
	cmdline, err := proc.CmdlineWithContext(ctx)
	if err != nil {
		return nil, err
	}

	command := &commandMetadata{command: cmdline, commandLine: cmdline}
	return command, nil
}
