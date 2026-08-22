package cmd

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestStartProfilingDisabledByDefault(t *testing.T) {
	t.Setenv("GHORG_PPROF", "false")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	if profilingActive {
		t.Fatal("expected profiling to be inactive when GHORG_PPROF is not true")
	}
	stopProfiling()

	if _, err := os.Stat(filepath.Join(dir, cpuProfileFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no cpu profile file to be created, got err: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, heapProfileFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no heap profile file to be created, got err: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, goroutineProfileFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no goroutine profile file to be created, got err: %v", err)
	}
}

func TestStartAndStopProfilingWritesProfiles(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	if !profilingActive {
		t.Fatal("expected profiling to be active when GHORG_PPROF is true")
	}

	stopProfiling()
	if profilingActive {
		t.Fatal("expected profiling to be inactive after stopProfiling")
	}

	cpuInfo, err := os.Stat(filepath.Join(dir, cpuProfileFileName))
	if err != nil {
		t.Fatalf("expected cpu profile file to exist, got err: %v", err)
	}
	if cpuInfo.Size() == 0 {
		t.Fatal("expected cpu profile file to be non-empty")
	}

	heapInfo, err := os.Stat(filepath.Join(dir, heapProfileFileName))
	if err != nil {
		t.Fatalf("expected heap profile file to exist, got err: %v", err)
	}
	if heapInfo.Size() == 0 {
		t.Fatal("expected heap profile file to be non-empty")
	}

	goroutineInfo, err := os.Stat(filepath.Join(dir, goroutineProfileFileName))
	if err != nil {
		t.Fatalf("expected goroutine profile file to exist, got err: %v", err)
	}
	if goroutineInfo.Size() == 0 {
		t.Fatal("expected goroutine profile file to be non-empty")
	}
}

func TestStartAndStopProfilingWritesBlockAndMutexProfiles(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	stopProfiling()

	blockInfo, err := os.Stat(filepath.Join(dir, blockProfileFileName))
	if err != nil {
		t.Fatalf("expected block profile file to exist, got err: %v", err)
	}
	if blockInfo.Size() == 0 {
		t.Fatal("expected block profile file to be non-empty")
	}

	mutexInfo, err := os.Stat(filepath.Join(dir, mutexProfileFileName))
	if err != nil {
		t.Fatalf("expected mutex profile file to exist, got err: %v", err)
	}
	if mutexInfo.Size() == 0 {
		t.Fatal("expected mutex profile file to be non-empty")
	}
}

func TestTracingDisabledByDefault(t *testing.T) {
	t.Setenv("GHORG_TRACE", "false")
	t.Setenv("GHORG_PPROF", "false")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	if traceActive {
		t.Fatal("expected tracing to be inactive when GHORG_TRACE is not true")
	}
	stopProfiling()

	if _, err := os.Stat(filepath.Join(dir, traceFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no trace file to be created, got err: %v", err)
	}
}

func TestStartAndStopTracingWritesTraceFile(t *testing.T) {
	t.Setenv("GHORG_TRACE", "true")
	t.Setenv("GHORG_PPROF", "false")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	if !traceActive {
		t.Fatal("expected tracing to be active when GHORG_TRACE is true")
	}
	if profilingActive {
		t.Fatal("expected pprof profiling to remain inactive when only GHORG_TRACE is true")
	}

	stopProfiling()
	if traceActive {
		t.Fatal("expected tracing to be inactive after stopProfiling")
	}

	traceInfo, err := os.Stat(filepath.Join(dir, traceFileName))
	if err != nil {
		t.Fatalf("expected trace file to exist, got err: %v", err)
	}
	if traceInfo.Size() == 0 {
		t.Fatal("expected trace file to be non-empty")
	}

	if _, err := os.Stat(filepath.Join(dir, cpuProfileFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no cpu profile file to be created when only tracing, got err: %v", err)
	}
}

func TestStopProfilingIsIdempotent(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()
	stopProfiling()
	// A second call must be a no-op and must not panic or error
	stopProfiling()

	if profilingActive {
		t.Fatal("expected profiling to remain inactive after repeated stopProfiling calls")
	}
}

func TestExitWithProfilesFlushesProfilesBeforeExiting(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	exitCode := -1
	originalExitFunc := exitFunc
	exitFunc = func(code int) { exitCode = code }
	defer func() { exitFunc = originalExitFunc }()

	startProfiling()
	exitWithProfiles(7)

	if exitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", exitCode)
	}
	if profilingActive {
		t.Fatal("expected profiling to be flushed before exiting")
	}

	cpuInfo, err := os.Stat(filepath.Join(dir, cpuProfileFileName))
	if err != nil {
		t.Fatalf("expected cpu profile file to exist after exitWithProfiles, got err: %v", err)
	}
	if cpuInfo.Size() == 0 {
		t.Fatal("expected cpu profile file to be non-empty after exitWithProfiles")
	}
}

func TestProfilingInterruptFlushesProfiles(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	exitCode := -1
	done := make(chan struct{})
	originalExitFunc := exitFunc
	exitFunc = func(code int) {
		exitCode = code
		close(done)
	}
	defer func() { exitFunc = originalExitFunc }()

	startProfiling()

	c := make(chan os.Signal, 1)
	go handleProfilingInterrupt(c)
	c <- os.Interrupt

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for interrupt handler to exit")
	}

	if exitCode != 130 {
		t.Fatalf("expected exit code 130 after interrupt, got %d", exitCode)
	}
	if profilingActive {
		t.Fatal("expected profiling to be flushed after interrupt")
	}

	cpuInfo, err := os.Stat(filepath.Join(dir, cpuProfileFileName))
	if err != nil {
		t.Fatalf("expected cpu profile file to exist after interrupt, got err: %v", err)
	}
	if cpuInfo.Size() == 0 {
		t.Fatal("expected cpu profile file to be non-empty after interrupt")
	}
}

func TestProfilingSigtermExitCode(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	exitCode := -1
	done := make(chan struct{})
	originalExitFunc := exitFunc
	exitFunc = func(code int) {
		exitCode = code
		close(done)
	}
	defer func() { exitFunc = originalExitFunc }()

	startProfiling()

	c := make(chan os.Signal, 1)
	go handleProfilingInterrupt(c)
	c <- syscall.SIGTERM

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for interrupt handler to exit")
	}

	if exitCode != 143 {
		t.Fatalf("expected exit code 143 after SIGTERM, got %d", exitCode)
	}
}

func TestStopProfilingIsSafeForConcurrentUse(t *testing.T) {
	t.Setenv("GHORG_PPROF", "true")
	t.Setenv("GHORG_TRACE", "true")
	t.Setenv("GHORG_QUIET", "true")

	dir := t.TempDir()
	restoreWd := chdirForTest(t, dir)
	defer restoreWd()

	startProfiling()

	// The interrupt handler may call stopProfiling concurrently with the main
	// goroutine, run with -race to catch unsynchronized access
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stopProfiling()
		}()
	}
	wg.Wait()

	if profilingActive || traceActive {
		t.Fatal("expected profiling and tracing to be inactive after concurrent stopProfiling calls")
	}
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("could not change working directory: %v", err)
	}
	return func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("could not restore working directory: %v", err)
		}
	}
}
