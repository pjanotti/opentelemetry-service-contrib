// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux || darwin || freebsd || openbsd

package processesscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper"

import (
	"context"
	"runtime"

	"github.com/shirou/gopsutil/v4/process"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper/internal/processesmetadata"
)

const (
	enableProcessesCount   = true
	enableProcessesCreated = runtime.GOOS == "openbsd" || runtime.GOOS == "linux"
)

func (s *processesScraper) getProcessesMetadata(ctx context.Context) (processesMetadata, error) {
	processes, err := s.getProcesses(ctx)
	if err != nil {
		return processesMetadata{}, err
	}

	countByStatus := map[processesmetadata.AttributeStatus]int64{}
	for _, process := range processes {
		var status []string
		status, err = process.Status()
		if err != nil {
			// We expect an error in the case that a process has
			// been terminated as we run this code.
			continue
		}
		state, ok := toAttributeStatus(status)
		if !ok {
			countByStatus[processesmetadata.AttributeStatusUnknown]++
			continue
		}
		countByStatus[state]++
	}

	// Processes are actively changing as we run this code, so this reason
	// the above loop will tend to underestimate process counts.
	// getMiscStats is a single read/syscall so it should be more accurate.
	miscStat, err := s.getMiscStats(ctx)
	if err != nil {
		return processesMetadata{}, err
	}

	var procsCreated *int64
	if enableProcessesCreated {
		v := int64(miscStat.ProcsCreated)
		procsCreated = &v
	}

	countByStatus[processesmetadata.AttributeStatusBlocked] = int64(miscStat.ProcsBlocked)
	countByStatus[processesmetadata.AttributeStatusRunning] = int64(miscStat.ProcsRunning)

	totalKnown := int64(0)
	for _, count := range countByStatus {
		totalKnown += count
	}
	if int64(miscStat.ProcsTotal) > totalKnown {
		countByStatus[processesmetadata.AttributeStatusUnknown] = int64(miscStat.ProcsTotal) - totalKnown
	}

	return processesMetadata{
		countByStatus:    countByStatus,
		processesCreated: procsCreated,
	}, nil
}

func toAttributeStatus(status []string) (processesmetadata.AttributeStatus, bool) {
	if len(status) == 0 || status[0] == "" {
		return processesmetadata.AttributeStatus(0), false
	}
	state, ok := charToState[status[0]]
	return state, ok
}

var charToState = map[string]processesmetadata.AttributeStatus{
	process.Blocked:  processesmetadata.AttributeStatusBlocked,
	process.Daemon:   processesmetadata.AttributeStatusDaemon,
	process.Detached: processesmetadata.AttributeStatusDetached,
	process.Idle:     processesmetadata.AttributeStatusIdle,
	process.Lock:     processesmetadata.AttributeStatusLocked,
	process.Orphan:   processesmetadata.AttributeStatusOrphan,
	process.Running:  processesmetadata.AttributeStatusRunning,
	process.Sleep:    processesmetadata.AttributeStatusSleeping,
	process.Stop:     processesmetadata.AttributeStatusStopped,
	process.System:   processesmetadata.AttributeStatusSystem,
	process.Wait:     processesmetadata.AttributeStatusPaging,
	process.Zombie:   processesmetadata.AttributeStatusZombies,
}
