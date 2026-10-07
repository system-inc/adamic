Rebased onto the requested stage1-lint area; real JSX comparisons now pass.
Tested source c88028f3 includes area 7481e032 and current main 39638d9e.
Owned Go byte gate PASS 478.251 s; bridge, filtered Node, registry and vet pass.
Twenty-one semantic byte mutants, nine metadata mutants and released-handle mutants caught.
No unclaimed ranked rules remain; regex validation and undefined labels stay incomplete.

The explicit area-base instruction was followed. At fetch,
origin/area/stage1-lint was 7481e0324e34a2537aafa9db7eeacda50405611b;
rebase onto it was clean. It includes the named harness 41eb6eab2 and current
main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Shared integration changes
were retained, including the JSX parser and finding-model/registry support.
No shared parser, registration generator, test harness or compiler files were
edited. During verification the remote area moved to d65a8f93; this gate describes
the fetched area base, not that later area tip. Main and the unit remote SHA
were unchanged at the final lease check. Only codex/typeaware-wave-14 is pushed.

The sole source changes are owned tests. The continuation's old expected JSX
panic is replaced by full production-Go/native/sanitized comparison of the real
JSX witness. The render suite additionally has 16 raw TSX controls, covering
number, nullable and mixed unions, boolean, bigint/zero bigint, number literals,
any/unknown, nested logical/conditional/coalescing expressions, generic constraints,
fragments, attributes and Unicode/CRLF spans. These are linted raw TSX inputs,
not new compiled Adamic .ts modules. Exact raw witnesses are retained as .tsx.txt.
All native rule implementations remain .a.

Real JSX has 13 findings and 5,657 identical bytes in normal and sanitized native.
The continuation witness has one finding and 426 identical bytes. The same native
judgment mutant is now run on actual JSX as well as synthetic contexts: it
compiles, exits 0 with empty stderr and differs only at byte comparison. This
removes real JSX parsing as the leaked-number-render claim's previous blocker;
it is not a high-level IR, single-assignment or capture-analysis parked claim.
No claim of agreement on every possible JSX project is made.

Setup succeeded: Go ready 0 s, clang/Node/submodules ready 1 s, cache warm and
done 55 s. nproc 5, cpu.max 400000 100000, memory 17.6 GB. Go 1.27.1,
clang 20.1.8, Node 24.19.0. One formatting invocation before sourcing the
environment reported gofmt not found; it was rerun successfully with the sourced
toolchain before test compilation. No check was skipped.

Workspace free space was 2.2 GB before sanitizer builds. Removed only 185 verified
obsolete ELF binaries/archive outputs from earlier wave-14 scratch directories,
freeing 7,680,104,794 bytes. Current run artifacts, source inputs, logs and tracked
evidence were preserved. Every removed absolute path is recorded in the cleanup
log. No repository source or other worker's artifacts were removed.

Commands write directly to retained logs and source /workspace/adamic-tools/env.sh:

    bash cloud/setup.sh
    go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
    ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript go test ./bridge/tsgo/... -count=1 -v -timeout=15m
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read)\.a$' -count=1 -v -timeout=10m
    go vet ./...
    go run ./cmd/lint-registry

The owned invocation sets ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript,
original and NEXT compiler/repository manifest variables to
/workspace/wave-14-artifacts/{compiler,repository}.manifest, and four artifact
variables to /workspace/wave-14-area-{original,next,render,third}. Frozen corpora
remain 77 compiler and 287 repository roots; none were filtered. Complete
stdout/stderr streams, logs, witnesses and benchmark rounds are retained under
validation-wave-14-area. Supported comparisons include findings, fixes and
suggestions in normal and ASan/UBSan/LeakSanitizer builds, with empty native stderr.

| Suite | Result |
| --- | --- |
| Continuation, including actual JSX | PASS 75.22 s |
| Render, including actual JSX and mutant | PASS 65.11 s |
| Original three rules | PASS 72.35 s |
| Third batch | PASS 265.56 s |
| Complete owned invocation | PASS 478.251 s |
| Bridge | PASS 91.238 s |
| Checker package | PASS 0.281 s |
| Filtered uncached Node | PASS 1.726 s |
| Registry and vet | PASS |

Earlier control counts remain 33 original, 15 continuation, 18 synthetic render,
60 third-batch, 58 constructor and 36 supported-pattern findings. Both frozen
corpora agree in every suite. Exact owned mutant observations:

```text
    wave_14_listeners_test.go:101: Go listener kinds [221] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [227, 214, 229] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [264, 232] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [295] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [257] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214, 215] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [307, 13] agree; wrong-kind declaration mutant caught
    wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 89
    wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4609
    wave_14_next_test.go:159: released-registry mutant exits 0, required panic catches it
    wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 50
    wave_14_render_test.go:137: real JSX mutant exits 0 with empty stderr; byte comparison catches byte 56
    wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 52
    wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1521
    wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4306
    wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
    wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 54
    wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8635
    wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 453
    wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10165
    wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6640
    wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12640
    wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 21726
    wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 49
    wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5054
    wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5642
    wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 17853
    wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 6783
    wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1332
    wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4189
```

Twenty-one semantic executions are caught only by complete byte comparison;
nine wrong-kind declaration mutants are metadata assertions. Released-handle
registry retention is separately caught by required panic 70. The bridge repeats
its seven ABI/type/link/ownership mutants: ASan catches lengths, the byte oracle
catches wrong type positions, stale/link assertions catch their mutations, and
LeakSanitizer catches missing frees and unowned region allocations. The filtered
uncached Node gate includes its one-byte mutant and inherited-static-field fixture.

Quiet alternating three-round whole-process count medians after verification:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.799003 | 0.336840 | 5.34x |
| repository | 0.269510 | 0.161784 | 1.67x |
| controls | 0.026189 | 0.035518 | 0.74x |
| constructors | 0.025795 | 0.036480 | 0.71x |
| upstream | 0.026222 | 0.035471 | 0.74x |
| patterns | 0.027264 | 0.040832 | 0.67x |

Compilation/sanitizers are excluded; constructor/upstream rows use --class-only.
The upstream row confirms count 89, not a freshly repeated full-byte comparison.
This resume does not claim a shared dispatch speed gain or new-rule emitted-JS
comparison. Registry validation confirms named kinds and optional handed-node
registration; no new rule is added to the registry in this unit.

The retained all-origin audit still has zero remaining ranked candidates before
additional main port exclusion. No new reservation is taken. Existing regex
runtime construction/Go-compatible error validation and undefined-label parsing
remain demonstrated Go-positive panic-70 boundaries. No custom matcher was
added or extended. The full repository gate, nondefault options, complete regex
grammar and every JSX project are not covered. The standalone 741-constructor
matrix was not repeated; its historical evidence remains available.
