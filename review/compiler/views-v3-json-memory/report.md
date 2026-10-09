Measured V3 JSON memory against main; no production change is included.
Base: compiler/views-v3 8880e6bc; comparison: main beb1be1b.
The full test was killed on both revisions; all serial sanitized chunks matched Node.
No check was added or removed; no semantic mutant or counts regression is claimed.
The original pool kill is not attributed conclusively; clang memory growth is isolated separately.

## Observations

The same box has nproc 5, a four-core CPU quota and a 17,179,869,184-byte memory limit. No extra cap was needed. Setup completed in 118.149 seconds: Go 0.044, Node 0.049, markdown 0.146, submodules 0.177, clang 0.403, Go build 117.447 seconds. The setup log is retained. GNU time was unavailable; package installation failed with permission denied. The included wait4 recorder measures each process's maximum RSS and status without changing its arguments or compiler flags.

Both runs used all 3,469 cases, the same physical package working directory, GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1. The package sources are identical between the two revisions. No sampling switch was set. Temporary Go overlays instrumented childguard, and a PATH wrapper instrumented clang. Neither instrument changed repository source. Every completed child command and RSS record is in evidence/*-children.json.gz; /proc samples also retain descendant identities and high-water marks. An initial main run from a worktree failed corpus discovery because WalkDir does not follow the cohere symlink; it is excluded.

| Measurement | Main | V3 |
| --- | ---: | ---: |
| Full test seconds | 233.966354 | 259.147236 |
| Full test status | SIGKILL, 137 | SIGKILL, 137 |
| Test binary peak RSS, KiB | 3,298,780 | 3,299,024 |
| Release native child peak RSS, KiB | 2,226,132 | 2,226,176 |
| Original sanitizer clang peak RSS, KiB | 279,212 | 1,232,608 |
| Sampled cgroup peak, bytes | 17,179,869,184 | 17,179,869,184 |

The cgroup oom_kill counter rose by two in each run. Both kills happened during parallel sanitized chunks, after clang had completed. The box also had substantial inherited /tmp tmpfs usage: memory.stat recorded about 7.97 GB of shmem during the main run. This is aggregate memory pressure, and the measurements do not establish V3 native-port growth. Surviving children were terminated after each killed parent; generated artifacts were preserved on disk. Main's initial setup briefly overlapped those surviving V3 children, which were stopped before its Node/native phases. Package timings are observations on this box, not a speed comparison on an otherwise idle machine.

All five chunks then ran serially, with ASAN_OPTIONS=detect_leaks=1, on both existing sanitized binaries. Every run exited zero, had empty stderr, and matched the corresponding Node output byte for byte. Results.json records output hashes and Node peaks as well as native peaks.

| Chunk | Main peak KiB | V3 peak KiB | Main seconds | V3 seconds |
| --- | ---: | ---: | ---: | ---: |
| 0 | 571,128 | 564,204 | 10.880274 | 11.988407 |
| 1 | 3,551,420 | 3,541,228 | 120.415154 | 111.999655 |
| 2 | 599,024 | 606,788 | 15.023648 | 10.236422 |
| 3 | 2,014,800 | 2,054,052 | 91.006305 | 82.853516 |
| 4 | 2,079,792 | 1,987,312 | 74.699190 | 74.345338 |

No top-level test was added. The lane check required gofmt on the inherited internal/ir/call_targets_guard_test.go; its existing TestCallTargetReaders passed in 29.97 seconds with GOMAXPROCS=4. Only map alignment changed. The existing test and some existing native chunks exceed that budget. No fixture was added; counts.md has zero changed rows.

## Compiler memory finding

The JSON C output differs by exactly 2,885 element_kind stores introduced by ef762440 in internal/native/view_arrays.go, on the path where ir.HasArrayViews is false. Removing only those stores makes V3's emitted C byte-identical to main's. A diagnostic compile with the same V3 runtime headers and sanitizer flags peaked at 1,233,560 KiB with stores and 279,016 KiB without them: a difference of 954,544 KiB, about 932 MiB. Compilation took 22.881902 and 9.519110 seconds respectively.

The stripped diagnostic object was never executed. Removing metadata is not proposed as a fix: the certificates must remain correct. Equivalent runtime helper calls are a possible compiler-memory fix, but this unit does not establish that compiler peak caused the reported pool kill. Runtime allocation counts cannot detect these nonallocating stores, so a mutant reintroducing them would require a compiler RSS proof rather than the requested runtime counts proof. No unsafe removal or claimed counts mutant is included.

## Commands and reproduction

All test output was redirected to logs. The recorder and measurement scripts are included; their paths describe this box. The childguard patch is a temporary measurement overlay, not a repository code change.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/v3-json-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# At V3; equivalent command at main with main-overlay.json and main.test:
go test -c -overlay=/tmp/v3-json-measure/v3-overlay.json -o /tmp/v3-json-measure/v3.test ./stage1/cohere/json > /tmp/v3-json-measure/v3-build.log 2>&1
# Both binary runs use /workspace/adamic/stage1/cohere/json as cwd:
python3 /tmp/v3-json-measure/measure.py v3 /workspace/adamic > /tmp/v3-json-measure/v3-measure.log 2>&1
python3 /tmp/v3-json-measure/measure.py main /workspace/adamic > /tmp/v3-json-measure/main-measure-valid.log 2>&1
# Actual measured command, one per revision:
/tmp/v3-json-measure/rss TEST_RSS REVISION.test -test.run=^TestPortMatchesGoCohere$ -test.count=1 -test.v -test.timeout=30m
python3 /tmp/v3-json-measure/replay.py > /tmp/v3-json-measure/replay.log 2>&1
```

The main CLI required go build -buildvcs=false because the detached worktree reused the existing cohere submodule through a symlink; its first CLI build's VCS-status failure is retained in the build log. Both native CLI builds succeeded after that workaround. Diagnostic clang records include the complete command and flags. Large binaries and stdout files remain outside Git; their hashes and RSS records are committed.

No whole package or full gate was run. The full target test remains red on this loaded box; the serial controls are green. A fresh-box comparison would be needed to distinguish the original pool's background pressure from this box's inherited tmpfs pressure.

## Integration lane

The first committed-tip lane run refused inherited gofmt drift in internal/ir/call_targets_guard_test.go. Applied gofmt, preserving all entries and behavior; no protected admission test was edited. The focused command was `GOMAXPROCS=4 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v`, with output retained in evidence/call-target.log.gz. The required lane command is rerun on the amended committed tip; its output is retained in evidence/lane-checks.log.

Lane output:

```text
lane checks 2.9 s: gofmt and tools on 47 Go files, t.Parallel on 6 test packages; no t.Parallel analyzer on this tree; vet 6 packages
```
