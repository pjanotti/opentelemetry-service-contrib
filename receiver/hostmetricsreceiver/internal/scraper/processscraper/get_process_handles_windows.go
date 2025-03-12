// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package processscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper"

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
	"go.opentelemetry.io/collector/scraper/scrapererror"
	"golang.org/x/sys/windows"
)

var (
	ErrNotImplementedError = errors.New("not implemented")

	procGetProcessIoCounters       = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetProcessIoCounters")
	procGetProcessMemoryInfo       = windows.NewLazySystemDLL("psapi.dll").NewProc("GetProcessMemoryInfo")
	procQueryFullProcessImageNameW = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryFullProcessImageNameW")
	procGetProcessHandleCount      = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetProcessHandleCount")
)

func getProcessHandlesInternal(ctx context.Context) (processHandles, error) {
	var processes []*windowsProcess

	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("could not create snapshot: %w", err)
	}
	defer windows.CloseHandle(snap)

	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))
	if err = windows.Process32First(snap, &pe32); err != nil {
		return nil, fmt.Errorf("could not get first process: %w", err)
	}

	machineMemory, errMachineMemory := mem.VirtualMemoryWithContext(ctx)
	if errMachineMemory != nil {
		errMachineMemory = fmt.Errorf("failed to get machine memory info: %w", errMachineMemory)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			process := windowsProcess{ProcessEntry32: pe32}
			processes = append(processes, &process)

			process.getProcessInfoViaHandle(ctx, machineMemory)
		}

		if err = windows.Process32Next(snap, &pe32); err != nil {
			break
		}
	}

	return &windowsProcesses{processes: processes}, nil
}

type windowsProcesses struct {
	processes []*windowsProcess
}

var _ processHandle = (*windowsProcess)(nil)
var _ processHandles = (*windowsProcesses)(nil)

func (p *windowsProcesses) Pid(index int) int32 {
	return int32(p.processes[index].ProcessID)
}

func (p *windowsProcesses) At(index int) processHandle {
	return p.processes[index]
}

func (p *windowsProcesses) Len() int {
	return len(p.processes)
}

// ioCounters is an equivalent representation of IO_COUNTERS in the Windows API.
// https://docs.microsoft.com/windows/win32/api/winnt/ns-winnt-io_counters
type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type windowsProcess struct {
	windows.ProcessEntry32
	errs           scrapererror.ScrapeErrors
	exe            string
	createTime     int64
	ioCountersStat process.IOCountersStat
	memoryInfoStat process.MemoryInfoStat
	memoryPercent  float32
	name           string
	numFDs         int32
	cpuTimesStat   cpu.TimesStat
	username       string

	process *process.Process
}

func (w *windowsProcess) getProcessInfoViaHandle(ctx context.Context, machineMemory *mem.VirtualMemoryStat) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, w.ProcessID)
	if err != nil {
		w.errs.Add(fmt.Errorf("failed to open process with pid %d: %w", w.ProcessID, err))
		return
	}
	defer windows.CloseHandle(h)

	// Gets the info retrieved by the CreateTimeWithContext method.
	var times windows.Rusage
	err = windows.GetProcessTimes(h, &times.CreationTime, &times.ExitTime, &times.KernelTime, &times.UserTime)
	if err != nil {
		w.errs.Add(fmt.Errorf("failed to get process times for pid %d: %w", w.ProcessID, err))
	}
	w.createTime = times.CreationTime.Nanoseconds() / 1_000_000

	// Gets the info retrieved by the IOCountersWithContext method.
	var ioCounters ioCounters
	ret, _, err := procGetProcessIoCounters.Call(uintptr(h), uintptr(unsafe.Pointer(&ioCounters)))
	if ret == 0 {
		w.errs.Add(fmt.Errorf("failed to get IO counters for pid %d: %w", w.ProcessID, err))
	} else {
		w.ioCountersStat = process.IOCountersStat{
			ReadCount:  ioCounters.ReadOperationCount,
			ReadBytes:  ioCounters.ReadTransferCount,
			WriteCount: ioCounters.WriteOperationCount,
			WriteBytes: ioCounters.WriteTransferCount,
		}
	}

	// Gets the info retrieved by the MemoryInfoWithContext method.
	var memCounters process.PROCESS_MEMORY_COUNTERS
	ret, _, err = procGetProcessMemoryInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&memCounters)), uintptr(unsafe.Sizeof(memCounters)))
	if ret == 0 {
		w.errs.Add(fmt.Errorf("failed to get memory info for pid %d: %w", w.ProcessID, err))
	} else {
		w.memoryInfoStat = process.MemoryInfoStat{
			RSS: memCounters.WorkingSetSize,
			VMS: memCounters.PagefileUsage,
		}
	}

	// Gets the info retrieved by the MemoryPercentWithContext method.
	if machineMemory.Total > 0 {
		w.memoryPercent = (100 * float32(memCounters.WorkingSetSize) / float32(machineMemory.Total))
	}

	// Gets the info retrieved by the ExeNameWith method.
	buf := make([]uint16, syscall.MAX_LONG_PATH)
	size := uint32(syscall.MAX_LONG_PATH)
	ret, _, err = procQueryFullProcessImageNameW.Call(
		uintptr(h),
		uintptr(0),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		w.errs.Add(fmt.Errorf("failed to get process name for pid %d: %w", w.ProcessID, err))
	} else {
		w.exe = windows.UTF16ToString(buf[:])
	}

	// Gets the info retrieved by the NameWithContext method.
	switch w.ProcessID {
	case 0:
		w.name = "System Idle Process"
	case 4:
		w.name = "System"
	default:
		w.name = filepath.Base(w.exe)
	}

	// Gets the info retrieved by the NumFDsWithContext method.
	var handleCount uint32
	ret, _, err = procGetProcessHandleCount.Call(uintptr(h), uintptr(unsafe.Pointer(&handleCount)))
	if ret == 0 {
		w.errs.Add(fmt.Errorf("failed to get handle count for pid %d: %w", w.ProcessID, err))
	} else {
		w.numFDs = int32(handleCount)
	}

	// Gets the info retrieved by the TimesWithContext method.
	var sysTimes process.SYSTEM_TIMES
	err = syscall.GetProcessTimes(
		syscall.Handle(h),
		&sysTimes.CreateTime,
		&sysTimes.ExitTime,
		&sysTimes.KernelTime,
		&sysTimes.UserTime)
	if err != nil {
		w.errs.Add(fmt.Errorf("failed to get CPU times for pid %d: %w", w.ProcessID, err))
	} else {
		// User and kernel times are represented as a FILETIME structure
		// which contains a 64-bit value representing the number of
		// 100-nanosecond intervals since January 1, 1601 (UTC):
		// http://msdn.microsoft.com/en-us/library/ms724284(VS.85).aspx
		// To convert it into a float representing the seconds that the
		// process has executed in user/kernel mode I borrowed the code
		// below from psutil's _psutil_windows.c, and in turn from Python's
		// Modules/posixmodule.c
		user := float64(sysTimes.UserTime.HighDateTime)*429.4967296 + float64(sysTimes.UserTime.LowDateTime)*1e-7
		kernel := float64(sysTimes.KernelTime.HighDateTime)*429.4967296 + float64(sysTimes.KernelTime.LowDateTime)*1e-7
		w.cpuTimesStat = cpu.TimesStat{
			User:   user,
			System: kernel,
		}
	}

	// Gets the info retrieved by the UsernameWithContext method.
	var token syscall.Token
	err = syscall.OpenProcessToken(syscall.Handle(h), syscall.TOKEN_QUERY, &token)
	if err != nil {
		w.errs.Add(fmt.Errorf("failed to open process token for pid %d: %w", w.ProcessID, err))
	} else {
		defer token.Close()
		tokenUser, err := token.GetTokenUser()
		if err != nil {
			w.errs.Add(fmt.Errorf("failed to get token user for pid %d: %w", w.ProcessID, err))
		} else {
			user, domain, _, err := tokenUser.User.Sid.LookupAccount("")
			if err != nil {
				w.errs.Add(fmt.Errorf("failed to lookup account for pid %d: %w", w.ProcessID, err))
			} else {
				w.username = domain + "\\" + user
			}
		}
	}

	// Temporary workaround until implementation of all methods directly
	w.process, err = process.NewProcess(int32(w.ProcessID))
	if err != nil {
		w.errs.Add(fmt.Errorf("failed to get gopsutil process handle for pid %d: %w", w.ProcessID, err))
	}
}

// CgroupWithContext implements processHandle.
func (w *windowsProcess) CgroupWithContext(ctx context.Context) (string, error) {
	// Not supported on Windows.
	return "", nil
}

// CmdlineSliceWithContext implements processHandle.
func (w *windowsProcess) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	return w.process.CmdlineSliceWithContext(ctx)
}

// CmdlineWithContext implements processHandle.
func (w *windowsProcess) CmdlineWithContext(ctx context.Context) (string, error) {
	return w.process.CmdlineWithContext(ctx)
}

// CreateTimeWithContext implements processHandle.
func (w *windowsProcess) CreateTimeWithContext(context.Context) (int64, error) {
	return w.createTime, nil
}

// ExeWithContext implements processHandle.
func (w *windowsProcess) ExeWithContext(context.Context) (string, error) {
	return w.exe, nil
}

// IOCountersWithContext implements processHandle.
func (w *windowsProcess) IOCountersWithContext(context.Context) (*process.IOCountersStat, error) {
	return &w.ioCountersStat, nil
}

// MemoryInfoWithContext implements processHandle.
func (w *windowsProcess) MemoryInfoWithContext(context.Context) (*process.MemoryInfoStat, error) {
	return &w.memoryInfoStat, nil
}

// MemoryPercentWithContext implements processHandle.
func (w *windowsProcess) MemoryPercentWithContext(context.Context) (float32, error) {
	return w.memoryPercent, nil
}

// NameWithContext implements processHandle.
func (w *windowsProcess) NameWithContext(context.Context) (string, error) {
	return w.name, nil
}

// NumCtxSwitchesWithContext implements processHandle.
func (w *windowsProcess) NumCtxSwitchesWithContext(context.Context) (*process.NumCtxSwitchesStat, error) {
	// Same as gopsutil, not supported on Windows.
	return nil, ErrNotImplementedError
}

// NumFDsWithContext implements processHandle.
func (w *windowsProcess) NumFDsWithContext(context.Context) (int32, error) {
	return w.numFDs, nil
}

// NumThreadsWithContext implements processHandle.
func (w *windowsProcess) NumThreadsWithContext(context.Context) (int32, error) {
	return int32(w.Threads), nil
}

// PageFaultsWithContext implements processHandle.
func (w *windowsProcess) PageFaultsWithContext(context.Context) (*process.PageFaultsStat, error) {
	// Same as gopsutil, not supported on Windows.
	return nil, ErrNotImplementedError
}

// PpidWithContext implements processHandle.
func (w *windowsProcess) PpidWithContext(context.Context) (int32, error) {
	return int32(w.ParentProcessID), nil
}

// RlimitUsageWithContext implements processHandle.
func (w *windowsProcess) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]process.RlimitStat, error) {
	// Same as gopsutil, not supported on Windows.
	return nil, ErrNotImplementedError
}

// TimesWithContext implements processHandle.
func (w *windowsProcess) TimesWithContext(context.Context) (*cpu.TimesStat, error) {
	return &w.cpuTimesStat, nil
}

// UsernameWithContext implements processHandle.
func (w *windowsProcess) UsernameWithContext(context.Context) (string, error) {
	return w.username, nil
}
