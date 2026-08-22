package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"syscall"

	"github.com/gabrie30/ghorg/colorlog"
)

const (
	cpuProfileFileName       = "ghorg-cpu.pprof"
	heapProfileFileName      = "ghorg-heap.pprof"
	goroutineProfileFileName = "ghorg-goroutine.pprof"
	blockProfileFileName     = "ghorg-block.pprof"
	mutexProfileFileName     = "ghorg-mutex.pprof"
	traceFileName            = "ghorg-trace.out"
)

var (
	// profilingMu guards the state below, stopProfiling can be called from
	// the interrupt handler goroutine concurrently with the main goroutine
	profilingMu               sync.Mutex
	cpuProfileFile            *os.File
	profilingActive           bool
	traceFile                 *os.File
	traceActive               bool
	interruptHandlerInstalled bool

	// exitFunc allows tests to observe exit codes without terminating the
	// test binary
	exitFunc = os.Exit
)

// startProfiling begins a CPU profile when GHORG_PPROF is set to true and an
// execution trace when GHORG_TRACE is set to true. Profile and trace files are
// written to the current working directory.
func startProfiling() {
	startCPUProfiling()
	startTracing()

	if profilingActive || traceActive {
		setupProfilingInterruptHandler()
	}
}

func startCPUProfiling() {
	if os.Getenv("GHORG_PPROF") != "true" {
		return
	}

	profilingMu.Lock()
	defer profilingMu.Unlock()

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

	// Record every blocking and mutex contention event. ghorg is mostly
	// waiting on git subprocesses and the network, so contention profiles are
	// more informative than the cpu profile and the overhead is acceptable for
	// an opt-in profiling run.
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	cpuProfileFile = f
	profilingActive = true
}

// startTracing begins a runtime execution trace when GHORG_TRACE is set to
// true. The trace shows scheduling and worker utilization over time and is
// analyzed with 'go tool trace'.
func startTracing() {
	if os.Getenv("GHORG_TRACE") != "true" {
		return
	}

	profilingMu.Lock()
	defer profilingMu.Unlock()

	f, err := os.Create(traceFileName)
	if err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not create trace file %s: %v", traceFileName, err))
		return
	}

	if err := trace.Start(f); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not start trace: %v", err))
		_ = f.Close()
		return
	}

	traceFile = f
	traceActive = true
}

// exitWithProfiles flushes any active profiles and trace before exiting.
// It must be used instead of os.Exit on paths where profiling may be active,
// because os.Exit skips deferred calls and would drop the profiles.
func exitWithProfiles(code int) {
	stopProfiling()
	exitFunc(code)
}

// setupProfilingInterruptHandler flushes profiles when the clone is
// interrupted with SIGINT or SIGTERM. Long clones are exactly when users hit
// Ctrl+C, and without this the CPU profile and trace would be lost. It is only
// installed when profiling or tracing is active so default signal behavior is
// otherwise unchanged.
func setupProfilingInterruptHandler() {
	if interruptHandlerInstalled {
		return
	}
	interruptHandlerInstalled = true

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleProfilingInterrupt(c)
}

func handleProfilingInterrupt(c chan os.Signal) {
	sig := <-c
	stopProfiling()

	// Mirror the exit codes a shell reports for death by these signals
	code := 130
	if sig == syscall.SIGTERM {
		code = 143
	}
	exitFunc(code)
}

// stopProfiling stops the CPU profile, writes heap, goroutine, block and mutex
// profile snapshots, and stops the execution trace. It is safe to call
// multiple times; only the first call after startProfiling does any work.
// Because ghorg exits via os.Exit, which skips deferred calls, this must be
// invoked before those exit points to flush profiles to disk.
func stopProfiling() {
	stopCPUProfiling()
	stopTracing()
}

func stopCPUProfiling() {
	profilingMu.Lock()
	defer profilingMu.Unlock()

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
	writeSnapshotProfile("block", blockProfileFileName)
	writeSnapshotProfile("mutex", mutexProfileFileName)

	// Stop collecting block and mutex events now that the snapshots are written
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)

	if os.Getenv("GHORG_QUIET") != "true" {
		wd, err := os.Getwd()
		if err != nil {
			wd = ""
		}
		profilePaths := []string{
			filepath.Join(wd, cpuProfileFileName),
			filepath.Join(wd, heapProfileFileName),
			filepath.Join(wd, goroutineProfileFileName),
			filepath.Join(wd, blockProfileFileName),
			filepath.Join(wd, mutexProfileFileName),
		}
		colorlog.PrintInfo(fmt.Sprintf("\npprof profiles written to %s, analyze with e.g. 'go tool pprof %s'", strings.Join(profilePaths, ", "), cpuProfileFileName))
	}
}

func stopTracing() {
	profilingMu.Lock()
	defer profilingMu.Unlock()

	if !traceActive {
		return
	}
	traceActive = false

	trace.Stop()
	if err := traceFile.Close(); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not close trace file %s: %v", traceFileName, err))
	}
	traceFile = nil

	if os.Getenv("GHORG_QUIET") != "true" {
		wd, err := os.Getwd()
		if err != nil {
			wd = ""
		}
		colorlog.PrintInfo(fmt.Sprintf("\nexecution trace written to %s, analyze with 'go tool trace %s'", filepath.Join(wd, traceFileName), traceFileName))
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

	defer func() { _ = f.Close() }()

	if err := pprof.Lookup(profileName).WriteTo(f, 0); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not write %s profile: %v", profileName, err))
	}

	if err := f.Close(); err != nil {
		colorlog.PrintError(fmt.Sprintf("Could not close %s profile file %s: %v", profileName, fileName, err))
	}
}
