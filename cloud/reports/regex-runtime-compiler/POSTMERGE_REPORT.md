Built and pushed dynamic native RegExp construction with opt-in compiler/tables and existing matcher ownership.
SHAs: lowering 6e47d677, integration merge da32fdeb, audit followup a469dd03, inherited test spelling 85b5016e.
Checks: no-overlay compiler contract 39.183s, Node backends, WASI, reference mutants, lint parity and repository vet pass.
Mutants: all 18 rerun and caught at their intended identity, refusal, ownership or symbol assertions; listed below.
Not covered: full repository/native concurrency gate; WASI sanitizer instrumentation; full native ports of two absent lint rules.

The explicitly approved Build hook is **internal/native/native.go:143**:
`library, err := runtimeLibraryForSource(source, options)`. It selects matcher
ownership support only when generated C contains the runtime compiler directive.
Compiler definitions and immutable property tables are included only by dynamic
RegExp lowering. There are no compiler caches, mutable compiler statics,
constructors or startup registration. Dynamic flags also require runtime parsing.

The branch merged the requested `origin/library/merge-p2b` at
`047e857207bee9aa62dbd22204445a0ab59d4319` with merge commit `da32fdeb`.
Its message records conflict assumptions: retain async lowering and concurrent
string-index publication, combine module declarations, retain both fixture sets,
use the incoming counted leak API, adapt the fuzzer's generic API and isolate its
feature-specific tests from parallel/move families. Earlier throw discovery is
preserved across later nonthrowing host operations. The test262 runner now uses
the same opt-in cached archive selector as Build.

A shell failed to stop on the merge's statics-marker failure before the merge
push. This was reported immediately and corrected in `a469dd03`; subsequent
commit/push commands stop on the first failed check. The final statics guard and
all regex gates pass. The inherited `slabs` test field was corrected to `Slabs`
in `85b5016e`, with the integration assumption in its message. Final gates need
no temporary Go overlay.

| Linux built-ins/RegExp | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| Integration 047e857 before runtime construction | 53 | 0 | 424 | 73 | 1329 |
| Merged runtime compiler branch | 56 | 0 | 421 | 73 | 1329 |

The runner is cmd/adamic-test262, test262 commit
7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd, adaptation on, five jobs, Linux
sanitizers. Baseline VCS stamping was disabled because its shared dependency
symlink prevents Git status; compiler semantics and runtime sources were
unchanged. Both JSON reports are under postmerge/. Every newly passing case
agrees with Node:

- built-ins/RegExp/S15.10.2.10_A2.1_T1.js
- built-ins/RegExp/S15.10.2.10_A2.1_T2.js
- built-ins/RegExp/S15.10.4.1_A8_T2.js

The compiler identity corpus is **5,746 test262 patterns**, including the
required 2,776, **875 pinned cohere patterns**, **4,000 seeded patterns** with
seed 0xEC2025, and nine reference probes. Linux and WASI compare **7,433
byte-identical programs and 3,197 syntax/refusal cases**. Compatibility adds
400 shape probes and one refusal-precedence probe, producing 7,677 programs
and 3,354 syntax/refusal cases. Parser checks compare complete counted Node
SyntaxError messages, including embedded NUL, and Go's exact divergence reasons.
Property identity checks 1,722 aliases, 448 unique properties, 23,045 ranges and
7,906 strings entry by entry. Tables remain static const.

Seven fixtures cover dynamic_gap, id-length exceptions, inline-comment ignores,
warning terms/decorations, flags/SyntaxErrors, ownership and argument order.
The uncached oracle passes against source Node and emitted JavaScript with
ASan/UBSan, release, slabs and leak checks. All seven also run under WASI with
ADAMIC_ORACLE_WASI=1. Four V8 divergences refuse with their exact Go reasons;
a valid counter wider than the native representation refuses as Error rather
than SyntaxError. The ownership mutant reaches a real ASan heap-use-after-free.
Matching still uses the existing native VM.

| Mutant | Intended check that caught it |
|---|---|
| Change ASCII property endpoint | C/Go entry identity |
| Accept lowercase ascii alias | rejected-alias identity |
| Remove compiler feature guard | ordinary-program symbol proof |
| Return NULL for properties only on WASI | WASI entry comparison/trap |
| Change generated provenance | regeneration identity |
| Stop top-level pattern at NUL | Go versus Node validity |
| Drop Other_ID_Start | raw/escaped capture name versus Node |
| Drop Other_ID_Continue | raw/escaped capture name versus Node |
| Accept Pattern_Syntax U+2E2F name | Node rejection |
| Wrap oversized Unicode escape | Node rejection |
| Accept clamped reversed bounds silently | exact V8DivergenceError |
| Classify reversed-bound divergence as SyntaxError | exact V8DivergenceError |
| Emit ASSERT instead of SET | bytecode comparison, case 0 byte 288 |
| Accept ? in C | rejection comparison, case 4700 |
| Bypass runtime V8 check | refusal status/reason, case 10630 |
| Drop compiled storage owner | ASan heap-use-after-free |
| Always select ownership archive | ordinary-program ownership symbols |
| Corrupt error message after NUL | complete counted-byte error comparison |

All mutants use real inputs and reach the named checks, not compiler diagnostics.
The portable table runner is internal/regexp/testdata/run-runtime-table-mutants.py;
its copies and overlays never modify repository sources. Other mutants are
ordinary Go tests. Their logs are under postmerge/.

| Fixture | Target | Raw before | Raw after | Brotli before | Brotli after |
|---|---|---:|---:|---:|---:|
| hello | wasm32-wasi | 288644 | 288644 | 84071 | 84071 |
| request | wasm32-wasi | 325240 | 325240 | 97552 | 97552 |
| hello | native release | 450752 | 450984 | 159682 | 161783 |
| request | native release | 451480 | 451720 | 162642 | 162561 |
| dynamic_gap | wasm32-wasi | - | 819371 | - | 158900 |
| dynamic_gap | native release | - | 1892096 | - | 291059 |

The post-merge common baseline uses the exact same merged front/runtime and
fixture/output basenames, toggling only the five new compiler translation units
through the recorded Go build overlay. This isolates their unused-program cost
from the requested integration's other changes. The property table unit remains
in both and is empty without the directive. Both hello/request WASI artifacts
are byte-identical, not merely equal sizes. Earlier actual git baseline a7f6f64
before/after measurements are in LOWERING_PROPOSAL.md.

Native symbol checks find no compiler, property table or ownership functions in
an ordinary program. Native text sizes are unchanged; request has identical
allocated section sizes, while hello rodata differs by 16 bytes. The assertion
cache path changed from a nine-digit to a ten-digit temporary directory suffix,
with padding and jump-table relocation changes. Native raw totals also include
five empty translation-unit filename symbols and alignment. Native Brotli cache
path noise is reported as allowed by the ruling; raw native bytes are not claimed
unchanged. The prior 14,236-byte hello was a different build.

WASI uses SDK 27 with 32-bit pointers, ADAMIC_TARGET_WASI=1 and no atomics.
Linux is the ASan/UBSan gate of record; this WASI SDK lacks those sanitizer
runtimes, so WASI execution itself is not instrumented.

The pinned cohere revision 7945d102a6c18dd36adf9114a758ce646e8b2359 passes
TestIdLength*, TestNoInlineComments*, TestNoWarningComments* and the warning
option decoder tests. The registered stage1 lint driver passes TestRulesAgree:
**13,053,452 identical output bytes on Go, Node, emitted JavaScript and sanitized
native**, in 154.534s. This includes no-warning-comments. The reachable lint
area and integration have no registered id-length or no-inline-comments ports;
for those two, evidence is their upstream tests plus the dynamic constructor
fixtures, not complete native rule-port certification. The upstream pinned
rules still use their existing Go regexp implementation; passing those tests
is not a claim of ECMAScript option equivalence beyond the held fixtures.

Commands (all output is saved directly to logs):

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go test ./internal/native -run '^TestRegExpRuntime|^TestRuntimeStaticsAreListed$|^TestRuntimeFieldLayoutsAreIncluded$' -count=1 -v
go test ./internal/regexp -run '^TestRuntimeReference' -count=1 -v
python3 internal/regexp/testdata/run-runtime-table-mutants.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestWASIAgreesWithNode)/internal/oracle/testdata/regexp_dynamic/|^TestDynamicRegExp' -count=1 -v
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode/internal/oracle/testdata/regexp_dynamic/' -count=1 -v
go test ./internal/lower ./internal/fresh ./cmd/adamic-test262 ./internal/fuzz -count=1
go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v
go vet ./...
```

Final no-overlay compiler gate: 39.183s. Seven reference mutants: 2.372s.
Uncached dynamic oracle: 197.883s. Explicit WASI oracle: 14.917s.
Full fuzzer package after merge fixes: 69.277s. Lower, fresh and runner packages
passed during merge validation; their initial slow runs included cold checker
compilation. Setup retry with --wasi-sdk succeeded in 88.271s: Go 0.145s,
Node 0.147s, submodules 0.448s, markdown dependencies 0.504s, clang 1.111s,
WASI SDK 1.271s, build cache 88.022s; nproc 5, CPU quota four cores.
The first setup failed on the fuzzer API conflict, which was fixed and rerun.

The full repository oracle and full native concurrency package were not rerun.
Earlier baseline TSan/signal-mutant failures are recorded in the prior report,
and are not recertified after this merge. Stress-scale nesting/allocation
exhaustion is not covered. Per-fixture leak checks pass; the full allocation
count table was not regenerated after this integration.

Relay to #adamic_runtime_platforms: codex/regex-runtime-compiler, lowering
6e47d677, integration da32fdeb, audit a469dd03, test spelling 85b5016e.
Overlaps: regexp.c/regexp.h ownership hooks; native.go:143 and library archive
selection; native expression dispatch and fields; lower regexp and exception
hooks; test262 archive selection. Merge resolutions also touch async declaration
dispatch, fuzzer API/tests and oracle leak plumbing. All new regex C units build
and execute on WASI. Hello/request WASI sizes are unchanged on the merged common
baseline. docs/runtime-statics.md now lists 40 newly merged host declarations
and 14 file markers; process exit status, lazy argument/cwd/performance caches
and performance marks remain explicitly unsafe for concurrent host calls and
need their owners' work. The regex compiler introduces none of that state.
