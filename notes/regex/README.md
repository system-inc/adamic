# Regex feature observations

One differing program; six new executable oracle programs; twenty-one additional
isolated compile-time refusal programs. The six executable programs all agree
on stdout, stderr and exit code across source Node, native release, native
sanitized, and emitted JavaScript. The oracle also checks leaks. Their six new
rows are in internal/oracle/counts.md; no existing row moved.

Base: d91e8f9afdce2d428584a62227c9a2fcac44e419. The separate matcher tip
50a1dc8217f6d715fc27038e6ce12ab429416a30 was inspected, including its literal
freshness commit. coverage.md maps existing and new source cases.

## Disagreement

Program: split_surrogate_assertions.a. Exact output files are alongside it.
Every execution exits 0 with empty stderr. Node's eight lines are:

```text
a�|�b
🌍
a�|�b
🌍
a�||�b
🌍
a�|�|b
�|�
```

Native's eight lines are:

```text
a🌍b
🌍
a🌍b
🌍
a🌍b
🌍
a🌍|b
🌍
```

The replacement glyphs are literal U+FFFD as printed by console.log of lone
surrogate halves. Node splits the pair at an assertion-only interior match.
The native split loop skips that candidate with regex_advance at
internal/native/runtime/regexp.c:789. It advances over the whole code point,
although the branch's regex_execute can now accept assertions inside a pair.
This is the suspected responsibility, not a patch or a proven root cause.

The emitted JavaScript also differs, printing:

```text
a🌍b
🌍
a�|�b
🌍
a🌍b
🌍
a�|�|b
�|�
```

Lowering inserts a limit of 4294967295 for an omitted split limit at
internal/lower/regexp.go:178. check-split-limits.py independently observes
that split(regex) and split(regex, 4294967295) differ in Node for the two u
patterns, with precisely the same changes seen in emitted JavaScript.
The explicit limit therefore accounts for this JavaScript discrepancy;
constructor spelling alone did not reproduce it in an isolated Node check.
check-difference.py temporarily
registers the note with the existing oracle, requires the stdout failure,
and removes the registration. The note is absent from permanent oracle fixtures.
No production code was changed.

## Mutation proof

check-mutant.py edits exactly one branch line at
internal/native/runtime/regexp.c:519: start = (size_t)at + 1 becomes + 2.
The new regex_coverage_methods.a fails at runtime with stdout differs (exit 1
from go test). A first witness is Node printing all:2:undefined:99 followed by
count:1:99, whereas the mutant starts with count:0:99. Both programs exit 0;
clang succeeds. The script restores the original source in finally and reruns
the oracle successfully. The production runtime diff is empty after restoration.

## Limits

Replacement callbacks are not supported. unsupported_replace.a and
unsupported_replaceAll.a print [a]b[a] on Node; build and js both refuse with
stage 0 can't lower regex replacement other than a string yet.

Scoped modifier literals fail TS18062 because this compiler targets before
es2025. The constant constructors in controls and refusal fixtures test those
patterns. new RegExp() fails TS2554, Expected 1-2 arguments, but got 0;
new RegExp('') is covered instead. Both rejected source attempts are preserved
as .a notes, with the build diagnostics in refusals.txt. Go provider injection,
mutable provider storage, and execution-budget controls are not accessible
from Adamic source; coverage.md explains these internal seams.

## Commands and evidence

All commands were run from /workspace/adamic. Branch inspection included
git fetch origin main, git diff origin/main...origin/codex/regex-v8-divergences,
and git log origin/main..origin/codex/regex-v8-divergences. The refreshed main
comparison contains just the V8 commit, rather than the earlier integrations
shown by the initial stale remote reference. Tests wrote to logs; no test
process was piped to head or tail. The repeatable commands for the final sources:

```sh
bash cloud/setup.sh > /tmp/regex-coverage-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
python3 notes/regex/check-probes.py > /tmp/regex-coverage/direct.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regex_coverage|TestRegexCoverageRefusals' -count=1 -timeout 15m -v > /tmp/regex-coverage/oracle-final.log 2>&1
python3 notes/regex/check-split-limits.py > /tmp/regex-coverage/split-limit-driver.log 2>&1
python3 notes/regex/check-difference.py > /tmp/regex-coverage/difference-driver.log 2>&1
python3 notes/regex/check-mutant.py > /tmp/regex-coverage/mutant-driver.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/regex-coverage/counts-update.log 2>&1
gofmt -l cmd internal > /tmp/regex-coverage/gofmt.log
go vet ./... > /tmp/regex-coverage/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/regex-coverage/full-gate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle ./internal/flow > /tmp/regex-coverage/final-changed-packages.log 2>&1
git diff --check
```

check-probes.py enumerates every new .a file and runs exactly these forms:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs <file>
go run ./cmd/adamic build <file> -o /tmp/regex-coverage/<stem>
/tmp/regex-coverage/<stem>
go run ./cmd/adamic js <file>
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/regex-coverage/<stem>.mjs
```

The binary and generated JavaScript are run only if their respective compilation
succeeds. direct.json records all thirty-two source attempts (six agreeing,
twenty-one expected refusals, one disagreement, four unsupported attempts).
Build and js were both attempted for every program. Refusal diagnostics and
exact difference output are retained; large agreeing stdout lives in scratch.

Setup: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc=5; cgroup cpu.max is
400000 100000. Go/clang/Node/submodules ready at 0s, cache warm at 65s, done
in 65s. setup.log contains the timing lines. The first setup attempt failed
cache warming because I switched branches while it was running; the stable
rerun above succeeded. Initial refusal probes used boolean console.log rather
than this subset's required string; they were corrected and all twenty-one then
passed exact-message checks. The initial direct JavaScript attempt bypassed
oracle/node.mjs and could not resolve adamic; all final observations use the
proper loader. These exploratory failures were not counted as differences.

Filtered final oracle: PASS in 3.239s. Full counts update: PASS in 17.949s.
The expected difference oracle fails in 0.394s; the real mutant fails in
8.760s. The restored targeted oracle passes in 0.429s. Final affected-package rerun passes: oracle 149.014s, flow 108.486s.
This rerun includes the additional final controls/refusals and their counts.
The repository-wide uncached gate passes (exit 0), including native in
230.508s, oracle in 183.095s, Unicode properties in 782.432s, and JSON in
549.993s. full-gate.log retains every package result. That run started with
the initial 19-refusal suite; after the final three safe controls and two
additional refusals, both affected packages were rerun uncached as above.
Formatting and vet have empty logs and exit 0. Staged whitespace checks pass.

Publish commands after validation:

```sh
git commit -m "Add regex oracle coverage and record split differences"
git push -u origin coverage/regex
```
