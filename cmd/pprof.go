package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/gabrie30/ghorg/colorlog"
)

const (
	cpuProfileFileName       = "ghorg-cpu.pprof"
	heapProfileFileName      = "ghorg-heap.pprof"
	goroutineProfileFileName = "ghorg-goroutine.pprof"
)

var (
	cpuProfileFile  *os.File
	profilingActive bool
)

// startProfiling begins a CPU profile when GHORG_PPROF is set to true. Profile
// files are written to the current working directory.
func startProfiling() {
	if os.Getenv("GHORG_PPROF") != "true" {
		return
	}

	f, err := os.Create(cpuProfileFileName)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not create cpu profile file %s: %v", cpuProfileFileName, err))
		return
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not start cpu profile: %v", err))
		_ = f.Close()
		return
	}

	cpuProfileFile = f
	profilingActive = true
}

// stopProfiling stops the CPU profile and writes heap and goroutine profiles.
// The heap and goroutine profiles are snapshots taken at this point in time.
// It is safe to call multiple times; only the first call after startProfiling
// does any work. Because ghorg exits via os.Exit, which skips deferred calls,
// this must be invoked before those exit points to flush profiles to disk.
func stopProfiling() {
	if !profilingActive {
		return
	}
	profilingActive = false

	pprof.StopCPUProfile()
	_ = cpuProfileFile.Close()
	cpuProfileFile = nil

	// Get up to date allocation statistics before writing the heap profile
	runtime.GC()
	writeSnapshotProfile("heap", heapProfileFileName)
	writeSnapshotProfile("goroutine", goroutineProfileFileName)

	if os.Getenv("GHORG_QUIET") != "true" {
		wd, err := os.Getwd()
		if err != nil {
			wd = ""
		}
		colorlog.PrintInfo(fmt.Sprintf("\npprof profiles written to %s, %s and %s, analyze with e.g. 'go tool pprof %s'", filepath.Join(wd, cpuProfileFileName), filepath.Join(wd, heapProfileFileName), filepath.Join(wd, goroutineProfileFileName), cpuProfileFileName))
	}
}

// writeSnapshotProfile writes one of the runtime/pprof snapshot profiles, such
// as heap or goroutine, to the given file in the current working directory.
func writeSnapshotProfile(profileName, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not create %s profile file %s: %v", profileName, fileName, err))
		return
	}
	if err := pprof.Lookup(profileName).WriteTo(f, 0); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not write %s profile: %v", profileName, err))
	}

	if err := f.Close(); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not close %s profile file %s: %v", profileName, fileName, err))
	}
}
