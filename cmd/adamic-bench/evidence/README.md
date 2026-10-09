# Step 38 observed benchmark

Final invocation (exit 0, harness completed; native finish line pending):

```sh
GOWORK=off /workspace/adamic-tools/go/bin/go build -o /workspace/bench-tools/adamic-bench ./cmd/adamic-bench
/workspace/bench-tools/adamic-bench --manifest cmd/adamic-bench/evidence/manifest.json --json cmd/adamic-bench/evidence/final-run.jsonl --warmup 1 --runs 5 --max-load 0.5 > cmd/adamic-bench/evidence/final-table.txt
```

Started 2026-10-08T21:46:03.355656557Z. Instrument: Go monotonic time.Now + Linux wait4 ru_maxrss (KiB).
Machine: INTEL(R) XEON(R) PLATINUM 8573C, 5 available logical
cores, Linux, Go go1.27.1; governor unavailable in this VM.
One-minute load before/after each invocation never exceeded 0.24 (limit 0.50).

Versions: Node v24.19.0, TypeScript 6.0.3, typescript-go
7.0.0-dev.20260707.2. `tools-package-lock.json` pins the npm packages and their
integrities. Reinstall with `npm ci --prefix /workspace/bench-tools` after copying
that lockfile and a package.json declaring those exact compiler packages; update
manifest absolute command paths for another machine. tsgo is the package's actual
Linux binary, with no Node launcher in the timed command.

The two programs share explicit minimal global declarations with `noLib`, not
each compiler's independently bundled libraries. `emit` produces JavaScript and
declarations; `diagnostics` deliberately exits 1 with a type error and emits
nothing. All 36 invocations (one warmup plus five timed per contestant/program)
match the oracle's input fingerprints, exit codes, stdout, stderr, and emitted
file hashes. The JSONL stores that evidence and the table stores medians/p10/p90.
The Node stand-in is marked pending in every record and table row. Its small
win or loss is timing noise between identical Node invocations, not native
compiler progress. These small programs mostly measure startup.

`run.jsonl` and `table.txt` preserve the first observed run; `final-run.jsonl`
and `final-table.txt` come from the final build, whose table names the pending
stand-in's measured win/loss as well as its pending status.

Validation:

```text
GORACE=atexit_sleep_ms=0 GOWORK=off go test -race -count=1 ./cmd/adamic-bench
ok github.com/system-inc/adamic/cmd/adamic-bench 0.727s
GOWORK=off go vet ./cmd/adamic-bench
exit 0
```

Mutants caught: changed emitted bytes, changed diagnostics, a late changed output
with an earlier good time, and load above threshold before contestant execution.
Pending never becomes pass; a deliberately slower equivalent contestant says
loss; timeout kills the process group. Input fingerprints are compared too.
Tests set process-wide child-fixture environment and therefore are not parallel.

Repository-wide `go vet ./...` and the uncached `go test -count=1 -timeout 30m ./...`
were attempted, and both stopped before running:

```text
go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory
```

The checkout's compiler submodule is uninitialized. The new command imports only
the standard library and its targeted checks pass with `GOWORK=off`. No claim
is made that the full repository gate passed.
