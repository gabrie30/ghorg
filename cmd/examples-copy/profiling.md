# Profiling Clones

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples profiling`.

ghorg can profile its own clone runs with the standard Go profiling tools. This is useful for understanding where a slow clone spends its time, tuning `GHORG_CONCURRENCY`, or attaching data to a performance issue.

There are two ways to profile, they can be used independently or together.

## pprof Profiles

Set `--pprof` on the clone command or `GHORG_PPROF=true`. When enabled, ghorg writes the following files to the directory the command was run from

| File | Profile | Shows |
| --- | --- | --- |
| `ghorg-cpu.pprof` | cpu | Where CPU time was spent over the whole clone |
| `ghorg-heap.pprof` | heap | Memory allocations, snapshot taken when the clone finishes |
| `ghorg-goroutine.pprof` | goroutine | What every goroutine was doing when the clone finished |
| `ghorg-block.pprof` | block | Where goroutines spent time blocked waiting |
| `ghorg-mutex.pprof` | mutex | Lock contention between goroutines |

Because clones spend most of their time waiting on git subprocesses and the network, the block and goroutine profiles are usually the most informative, the CPU profile is often mostly idle.

```bash
ghorg clone my-org --pprof

# analyze the profiles
go tool pprof ghorg-cpu.pprof
go tool pprof ghorg-heap.pprof
go tool pprof ghorg-goroutine.pprof
go tool pprof ghorg-block.pprof
go tool pprof ghorg-mutex.pprof

# or open an interactive flame graph in the browser
go tool pprof -http=:8080 ghorg-block.pprof
```

## Execution Trace

Set `--trace` on the clone command or `GHORG_TRACE=true`. This writes a runtime execution trace (`ghorg-trace.out`) to the directory the command was run from. The trace shows scheduling and worker utilization over time, which is the best tool for tuning `GHORG_CONCURRENCY`, you can see whether the worker pool is actually saturated or throttled.

```bash
ghorg clone my-org --trace

# analyze the trace, opens in the browser
go tool trace ghorg-trace.out
```

## Things to know

1. Profiles and the trace are flushed even when the clone exits early with an error or is interrupted with Ctrl+C, so it is safe to profile a long clone and stop it partway through.

1. Trace files can get large on big clones, prefer `--pprof` unless you specifically need the timeline view.

1. `--pprof` and `--trace` can be combined on the same run.

1. Successive profiled runs overwrite the previous profile files, move them elsewhere if you want to compare runs.
