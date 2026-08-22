package cmd

import (
	"os"
	"path/filepath"
	"testing"
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
