# Commands and observations

All commands ran from /workspace/adamic unless stated otherwise. Test stdout and
stderr went to the named logs under /tmp/adamic-gate; logs were read after runs.
The cloud-environment-runtime skill was used to inspect runtime/network readiness.

## Setup and branch

```sh
cat CLAUDE.md
cat README.md docs/0.1.md docs/memory.md
cat /etc/codex/network-policy.json
git remote -v
git branch -a
git status --short
git fetch origin main codex/normalize-long
bash cloud/setup.sh
nproc
git fetch origin codex/normalize-long:refs/remotes/origin/codex/normalize-long
git switch -c coverage/normalize-long origin/codex/normalize-long
git log --oneline origin/main..origin/codex/normalize-long
git diff origin/main...origin/codex/normalize-long
git diff --name-only origin/main...origin/codex/normalize-long
```

The initial log/diff request failed because the target remote ref had not been
fetched. The explicit refspec fetch fixed it. Setup: Go ready 0s; clang ready 0s;
Node ready 0s; submodules ready 0s; build cache warm 75s; done 75s. nproc=5;
cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.
Every toolchain command below sourced `/workspace/adamic-tools/env.sh`.

All changed files were read, as were oracle_test.go, counts_test.go, existing
normalize*.a, optional_strings.a, string_limits.a, concat_too_long.a and the
string length guard. `rg -n '\.normalize\(' internal/oracle/testdata` found the
existing normalization calls. Coverage.md records the resulting audit.

## Oracle and counts

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/normalize_coverage_' -count=1 -timeout 30m > /tmp/adamic-gate/normalize-coverage-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/normalize_coverage_limit.a$' -count=1 -timeout 5m > /tmp/adamic-gate/normalize-coverage-limit.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-gate/normalize-coverage-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/normalize_coverage_' -count=1 -timeout 30m > /tmp/adamic-gate/normalize-coverage-final-oracle.log 2>&1
```

Initial trial: Node timed out on the 300,000-mark input, as described in coverage.md.
Reduced that input to 12,288 marks. Final oracle passed in 10.212s. Limit-only oracle
passed in 11.985s. Counts update passed in 22.494s and added only six rows.

## Standalone builds and runs

```sh
mkdir -p /tmp/adamic-gate/normalize-coverage
for file in internal/oracle/testdata/normalize_coverage_*.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/normalize-coverage/$name" > "/tmp/adamic-gate/normalize-coverage/$name.build.log" 2>&1 || exit 1
 "/tmp/adamic-gate/normalize-coverage/$name" > "/tmp/adamic-gate/normalize-coverage/$name.native.txt" 2> "/tmp/adamic-gate/normalize-coverage/$name.native.err"
 native_status=$?
 node --input-type=module-typescript < "$file" > "/tmp/adamic-gate/normalize-coverage/$name.node.txt" 2> "/tmp/adamic-gate/normalize-coverage/$name.node.err"
 node_status=$?
 cmp "/tmp/adamic-gate/normalize-coverage/$name.node.txt" "/tmp/adamic-gate/normalize-coverage/$name.native.txt" || exit 1
 echo "$name stdout agrees; native exit=$native_status node exit=$node_status"
done
```

All six stdout comparisons passed. cache/long/quick/repeat/stream: both exit 0,
empty stderr. limit: stdout `29826161` followed by newline on both; Node exits 1
with `RangeError: Invalid string length` and stack; native exits 70 with
`adamic: panic: RangeError: Invalid string length`. The oracle translates Node's
exception to Adamic's documented panic contract and passes this case.

An earlier standalone command used `node --experimental-strip-types
--input-type=module`, which rejected TypeScript annotations; corrected to
`--input-type=module-typescript`. The obsolete 300,000-mark standalone Node trial
was stopped with `kill -TERM 3229`, then all final files were rebuilt and rerun.

## Formatting and repository checks

From /workspace/adamic/cohere:

```sh
source /workspace/adamic-tools/env.sh
go run ./command/cohere --help
go run ./command/cohere --directory /workspace/adamic --format-only --no-cache /workspace/adamic/internal/oracle/testdata/normalize_coverage_*.a > /tmp/adamic-gate/normalize-coverage-format.log 2>&1
go run ./command/cohere --directory /workspace/adamic --no-fix --no-cache /workspace/adamic/internal/oracle/testdata/normalize_coverage_*.a > /tmp/adamic-gate/normalize-coverage-cohere.log 2>&1
```

Format-only passed. The type/lint command exited 1 because this checked-out cohere
rejects `.a` files as unsupported, despite the root sourceExtensions setting.
An earlier relative glob from cohere did not select the files; the absolute paths
above were the successful formatting run. Type-checking/lowering in the Adamic
oracle accepted every new program.

From /workspace/adamic:

```sh
source /workspace/adamic-tools/env.sh
gofmt -w internal/oracle/oracle_test.go
gofmt -l cmd internal
git diff --check
go vet ./... > /tmp/adamic-gate/normalize-coverage-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/adamic-gate/normalize-coverage-gate.log 2>&1
```

Gofmt and diff check had no findings; vet passed. One gofmt invocation before
sourcing env.sh failed because gofmt was not on PATH; the sourced rerun passed.

The output helper was subsequently changed from a mutable array to a string, so
flow tracing does not repeatedly snapshot a growing array. The million-unit
fixture is excluded from flow instruction tracing following the existing large
fixture exclusions; all six remain registered with the full oracle. The final
oracle and standalone loop above were repeated after these changes. Final oracle
passed in 42.699s while the broader gate was running; counts passed in 51.212s.
The repeat fixture was separately rerun after correcting the byte cutoff formula
from length-8 to length-7, with another standalone comparison and counts update.

## Mutation proof

Temporarily changed exactly line 391 of internal/native/runtime/normalize.c:
`finish_segment(state, true);` to `finish_segment(state, false);` in accept_point.
The test command was run with ADAMIC_GATE_UNCACHED=1; source was restored in a
Python finally block even if the run failed.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/normalize_coverage_stream.a$' -count=1 -timeout 30m > /tmp/adamic-gate/normalize-coverage-mutant.log 2>&1
# Restore the original line.
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/normalize_coverage_stream.a$' -count=1 -timeout 30m > /tmp/adamic-gate/normalize-coverage-restored.log 2>&1
```

The mutant compiled and both executions finished with exit 0, but stdout differed.
The mixed slice's NFKC result had length 315 on Node and 316 on native: the mutant
emitted AC00 followed by 11A8 where Node emitted AC01. This is a semantic failure,
not a compiler warning or sanitizer compile failure. The mutant oracle exited 1
in 50.123s. Restored oracle passed in 3.928s. git diff confirmed no runtime change.

The complete gate was rerun with bounded package/test parallelism after the final
fixture and tracing changes:

```sh
go vet ./... > /tmp/adamic-gate/normalize-coverage-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -p 2 -parallel 2 -count=1 -timeout 30m ./... > /tmp/adamic-gate/normalize-coverage-gate-final.log 2>&1
```

The superseded full gate was terminated after the final gate started. Its active
children were still tracing earlier fixture bodies with the costly array helper.
A Python process-tree walk of `ps -eo pid=,ppid=,args=` verified the original
`go test -count=1 -timeout 30m ./...` PID 7462, then sent SIGTERM to its descendants
and parent, leaving the final gate running. This first gate is an aborted trial,
not a passing repository check.

An environment restart subsequently killed the final gate's session. Its log
records passing results through internal/regexp, including flow (104.742s),
native (110.777s) and the full oracle (145.002s). The workspace and all edits
survived; the killed processes were defunct and no longer running. The packages
whose completion was not recorded were run separately, uncached:

```sh
go list ./...
ADAMIC_GATE_UNCACHED=1 go test -p 4 -parallel 2 -count=1 -timeout 30m ./internal/unicodeproperties ./stage1/cohere/... ./stage1/typescript/... > /tmp/adamic-gate/normalize-coverage-gate-remaining.log 2>&1
```

The all-package invocation itself was interrupted, not reported as passing.

The remaining gate exited 0. All fourteen packages passed: Unicode properties
503.929s; cssnumbers 60.983s; cssstrings 9.418s; formatfiles 36.833s; gitignore
39.937s; graphql 59.297s; json 360.243s; lint 114.091s; mediaquery 21.718s;
selector 31.535s; suppression 45.244s; values 51.856s; parser 175.867s;
scanner 38.081s. Together with the completed results from the interrupted final
all-package run, every repository package has a passing uncached run against the
final code. Vet, gofmt and git diff --check passed. No completed normalization
output disagreement was found.

## Commit and push

```sh
git add internal/flow/flow_test.go internal/oracle/oracle_test.go internal/oracle/counts.md internal/oracle/testdata/normalize_coverage_*.a notes/normalize-long/coverage.md notes/normalize-long/runs.md
git commit -m "Cover long normalization paths against Node"
git push -u origin coverage/normalize-long
```
