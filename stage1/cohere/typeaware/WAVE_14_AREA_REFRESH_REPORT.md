Rebased onto the latest requested stage1-lint area and re-greened all owned rules.
Tested source 7a600bed3 includes area d65a8f93 and current main 39638d9e.
Owned Go byte gate PASS 511.172 s; bridge, uncached filtered Node, registry and vet pass.
Twenty-one semantic byte mutants, nine metadata mutants and released-handle mutants caught.
No unclaimed rules remain; regex validation and undefined labels are incomplete.

This refresh takes the runtime allocation and string changes that landed after
area 7481e032. Rebase onto origin/area/stage1-lint
(d65a8f931c98655936ae04c6899f38f14862b73e) was clean. Current main
39638d9e278d38bb5aeae887f46d55a70e47aaad is an ancestor, as is the requested
shared harness. No owned semantic source changes were required. Integration's
runtime, oracle and evidence changes were preserved; no shared file was edited
or reverted. All source changes from the prior real-JSX unit remain.

Main, area and remote unit branch stayed unchanged during verification.
The exact lease for publishing this explicitly requested rebase is remote
unit d85dd3dee795ea91725b14c4da11415036afbac0. The only push target is
codex/typeaware-wave-14. Neither main nor any area branch is pushed.

Setup completed in 105 s: Go/clang/Node ready at 0 s, submodules ready at 1 s,
cache warm at 105 s. nproc 5, cpu.max 400000 100000, memory 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0. Tests write directly to retained logs.

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
    ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript go test ./bridge/tsgo/... -count=1 -v -timeout=15m
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read|runtime_last_index_of)\.a$|^TestRuntimeLastIndex' -count=1 -v -timeout=10m
    go vet ./...
    go run ./cmd/lint-registry

Owned tests use ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript,
the original and NEXT compiler/repository manifest variables pointing to
/workspace/wave-14-artifacts/{compiler,repository}.manifest, and the four
artifact variables pointing to /workspace/wave-14-area-refresh-{original,next,
render,third}. The frozen corpora remain 77 compiler and 287 repository roots;
no inputs were filtered. Full streams/logs and benchmark rounds are retained
in validation-wave-14-area-refresh. The exact real-JSX witness sources remain
in validation-wave-14-area/inputs, unchanged apart from scratch path prefixes.

| Gate | Result |
| --- | --- |
| Continuation | PASS 81.69 s |
| Render judgments and real JSX | PASS 74.64 s |
| Original rules | PASS 81.34 s |
| Third batch | PASS 273.49 s |
| Complete owned invocation | PASS 511.172 s |
| Bridge | PASS 134.932 s |
| Checker package | PASS 0.311 s |
| Filtered uncached Node | PASS 13.184 s |
| Runtime last-index-of Node comparison | PASS 11.75 s |
| Registry and vet | PASS |

All supported findings, fixes and suggestions agree with unchanged production
Go under normal and ASan/UBSan/LeakSanitizer builds, with empty native stderr.
Control populations retain 33 original, 15 continuation, 18 synthetic render,
13 real JSX, 60 third-batch, 58 constructor and 36 supported-pattern findings.
Both frozen corpora agree in every suite; the continuation's real JSX witness
also retains one finding under sanitizers. The uncached Node gate includes
inherited static reads, the newly landed runtime string fixture and its direct
Node comparison. Registry validation is green with the integrated finding model.

Every owned mutant's exact observation and catcher:

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
    wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 105
    wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4689
    wave_14_next_test.go:159: released-registry mutant exits 0, required panic catches it
    wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 58
    wave_14_render_test.go:137: real JSX mutant exits 0 with empty stderr; byte comparison catches byte 64
    wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 60
    wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1561
    wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4418
    wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
    wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 62
    wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8755
    wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 469
    wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10349
    wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6720
    wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12904
    wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 22150
    wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 57
    wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5230
    wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5842
    wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 18269
    wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 7007
    wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1380
    wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4757
```

Twenty-one semantic runs compile and exit 0 with empty stderr and are caught
only by full byte comparison. Nine declaration mutants are metadata assertions.
Released-handle retention mutants are caught separately by required panic-70
assertions. The bridge repeats its seven ABI/type/link/ownership mutants: ASan
catches lengths, comparison catches wrong type positions, stale/link assertions
catch their mutants, and LeakSanitizer catches missing frees and unowned regions.
The uncached Node gate repeats its one-byte oracle mutant.

Quiet alternating three-round whole-process count medians after verification:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.841067 | 0.354245 | 5.20x |
| repository | 0.257722 | 0.140896 | 1.83x |
| controls | 0.025685 | 0.030103 | 0.85x |
| constructors | 0.031693 | 0.033993 | 0.93x |
| upstream | 0.025752 | 0.039613 | 0.65x |
| patterns | 0.031424 | 0.046258 | 0.68x |

Compilation/sanitizers are excluded. Constructor and upstream rows use
--class-only; the upstream row checks count 89, not a freshly repeated full-byte
upstream comparison. No new kind-indexed dispatch speedup is claimed.

The refreshed all-origin audit still leaves zero ranked candidates before any
additional main port exclusion. selection.json retains ref SHAs, distinct claim
blobs and matching names. No new rules or reservations were added. New rules
would use named ast.Kind declarations and handed-node visits; no numeric API
requirement is imposed. Real JSX remains supported by the landed parser and is
no longer the leaked-render blocker. Broader native regex validation/errors and
undefined labels remain independent Go-positive panic-70 boundaries. No custom
matcher was added or extended. Nondefault options, the full repository gate,
complete regex grammar and every JSX project remain outside covered scope.
No standalone 741-constructor matrix or new-rule emitted-JS gate was repeated.
