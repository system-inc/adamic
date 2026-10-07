Built: all three continuation rules are native and production-oracle tested in the owned profile.
Commits: continuation claim 126c0358; race/core groundwork 7330620a; completion follows in this commit.
Commands and outputs: 100 upstream programs, compiler 77, repository 287 and race controls 22 agree normally and under sanitizers.
Mutants: process callee 6 programs, blocking order 19 programs, race byte 50; raw facts/state and bridge ownership mutants also caught.
Not covered: shared registration/integration and the full repository Go gate; no additional rules claimed before this completion.

The native ports are `no_process_exit_after_output.a`, `no_uncleared_race_timeout.a`
and `require_blocking_standard_streams.a`. Earlier partial refusal behavior from
7330620a is replaced by complete graph construction and module/function ordering.
`native_graph.a` ports the pinned Go control-flow builder, including branches,
loops, short-circuiting, optional chains, labels, destructuring, exception forks,
and duplicated finally paths. Rule decisions remain native. Raw checker facts
provide resolved overload declarations/return flags, program membership/module
resolutions/import syntax, and symbol declarations/ancestors. They return no lint
verdicts. External-module metadata gives the owned parser top-level await context.

The user prohibited shared dispatcher, harness and generator edits. Therefore
`overlay_bridge.py` writes a scratch Go build overlay that connects the three raw
questions; it does not modify any shared source file. `suite.a` is the owned
three-rule profile (`--all`) and can run process rules individually. The unified
shared harness still needs central registration/profile integration. A build
without this owned overlay explicitly rejects unsupported questions. All new
Adamic implementation files are `.a`; upstream TypeScript input fixtures are
materialized in scratch only.

Production evidence is in `validation-process/`. `capture_process.py` wraps the
unmodified upstream Go test calls, retaining their assertions and original
configs, and captures all 100 fixture programs, including 14 one-level callee
programs. `testdata/oracle_process.go` uses Go cohere's unmodified production
rules with an independent loader and walk. The byte protocol serializes ranges,
rule/message ids, descriptions, every fix, and every suggestion. These process
rules have no fixes or suggestions. Findings over captured programs: {'nexus/correctness-no-process-exit-after-output': 129, 'nexus/correctness-require-blocking-standard-streams': 21}.
Normal and ASan/UBSan output is identical across all 100 programs. Both required
corpora have zero findings for this continuation: compiler 77 files / 5,780 bytes;
repository 287 files / 18,485 bytes. The race controls produce 13 findings across
22 files / 8,262 bytes with their fresh scratch paths, normally and sanitized.
The initial three rules' nonzero corpus evidence remains in ../WAVE08_REPORT.md.

Each compiling rule mutant exits 0 with empty stderr and is rejected solely by
production output comparison: disabling one-level process callees differs in six
programs (first byte 164); disabling blocking order differs in nineteen (first
byte 1690); inverting lost-race-timeout recognition differs at byte 50. Raw
resolved return flags, resolved module targets and ancestor names have compiling
mutants caught by pinned checker witnesses. Native write-state and blocking-state
mutants compile and exit 0, but disagree with the independent Go event cases.
Every new question rejects a released handle with exit 70 before output.

The full bridge gate passed in 101.76s with the owned overlay: C ABI ownership,
100 queries, outputs surviving release, stale/zero handles, 162-position native
oracle, ASan/UBSan/LSan, and all existing input/output length, stale handle,
wrong-position, link guard, C-output leak and region ownership mutants. Commands:
`GOFLAGS=-overlay=/workspace/wave08-resume/overlay.json go test ./bridge/tsgo
./bridge/tsgo/checker -count=1 -v`, `go test
./stage1/cohere/typeaware/wave08-next -count=1 -v`, `go vet
./stage1/cohere/typeaware/wave08-next`, and gofmt verification. Test output is
written to log files, never piped. The filtered Node oracle and setup evidence
from the original three ports still apply; protected compiler source is untouched.
Setup completed in 132s; nproc was 5 with a four-core CPU quota.

Quiet end-to-end timing includes fresh program load, rules and serialization.
Three alternating rounds after concurrent verification completed: compiler
native 4.131370s versus Go 0.433010s
(9.541x); repository native
0.501672s versus Go 0.166121s
(3.020x). These are observations;
no claim is made that this native profile is faster. Concurrent validation timing
logs are retained but the quiet timing results above are the reported comparison.

Reproduce using the installed toolchain environment and the stage 0 compiler:

```sh
python3 stage1/cohere/typeaware/wave08-next/overlay_bridge.py /tmp/wave08-bridge
CC=clang go build -overlay /tmp/wave08-bridge/overlay.json -buildmode=c-archive -o /tmp/wave08-bridge/checker.a ./bridge/tsgo/archive
CC=clang CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all' go build -overlay /tmp/wave08-bridge/overlay.json -buildmode=c-archive -o /tmp/wave08-bridge/checker-asan.a ./bridge/tsgo/archive
python3 stage1/cohere/typeaware/wave08-next/capture_process.py /tmp/wave08-fixtures
python3 stage1/cohere/typeaware/wave08-next/validate_process.py --artifacts /tmp/wave08-process --fixtures /tmp/wave08-fixtures --stage0 /path/adamic --archive /tmp/wave08-bridge/checker.a --asan-archive /tmp/wave08-bridge/checker-asan.a --compiler-root /path/TypeScript
python3 stage1/cohere/typeaware/wave08-next/validate.py --artifacts /tmp/wave08-race --stage0 /path/adamic --archive /tmp/wave08-bridge/checker.a --asan-archive /tmp/wave08-bridge/checker-asan.a --compiler-root /path/TypeScript
python3 stage1/cohere/typeaware/wave08-next/validate_pending.py --artifacts /tmp/wave08-raw --stage0 /path/adamic
python3 stage1/cohere/typeaware/wave08-next/time_process.py --artifacts /tmp/wave08-timing --native /tmp/wave08-process/native --go /tmp/wave08-process/oracle --compiler-root /path/TypeScript
```

`validation-process/streams.json.gz` holds full validation output; inputs and
source hashes are saved beside it. Older `validation/` records the 7330620a
partial stage and is retained for provenance, not asserted as the current status.
The next claim search happens only after this completed work is committed/pushed.
