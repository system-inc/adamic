Rebased the existing wave-14 branch; the unclaimed ranked pool is exhausted.
Tested source 1c56e364 on current main b8fb957a, replacing pushed 293ca9388.
Owned Go byte suites PASS 461.469 s; bridge, filtered uncached Node and vet pass.
Twenty semantic byte mutants, nine declaration mutants and released-handle mutants caught.
No new rules claimed; existing regex, JSX and undefined-label boundaries remain incomplete.

Selection audit: 584 fetched origin refs, 33 distinct Markdown claim blobs and
197 checker-dependent ranking rows. All rows are either in the original base
ports or mentioned in a claim on an origin branch, so there are zero candidates
even before excluding any additional main ports. This conservative reservation
scan does not reclaim parked or incomplete claims. selection.json retains every
ref SHA, claim blob/path/ref, matching rule names, baseline ports and empty
remaining list. No new claim was pushed and no rule code was written.

Main advanced from c01907a7 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Its four-file change fixes inherited static-field reads and adds an oracle
fixture. Rebase was clean, preserving integration's changes. No shared files
were edited or reverted by this worker. The landing-first rebase and oracle
refresh are this unit. Before publishing, main remained b8fb957a and the unit
branch remained 293ca93888dd1896226a792625c53ac235cc7e1b. Publishing uses that
exact lease for the explicitly requested rebase, and targets only
codex/typeaware-wave-14. No main or area branch is pushed.

Setup printed Go/clang/Node/submodules ready at 0 s, cache warm at 89 s and
finished at 89 s. nproc 5, cpu.max 400000 100000, memory 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0. All tests write directly to log files.

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
    ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript go test ./bridge/tsgo/... -count=1 -v -timeout=15m
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read)\.a$' -count=1 -v -timeout=10m
    go vet ./...

Owned invocation sets ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript,
the original and NEXT compiler/repository manifest variables to
/workspace/wave-14-artifacts/{compiler,repository}.manifest, and the four
artifact variables to /workspace/wave-14-exhausted-{original,next,render,third}.
The frozen populations remain 77 compiler roots and 287 repository roots.
No corpus was filtered. Complete streams, logs and benchmark rounds are retained
under validation-wave-14-exhausted. Normal and ASan/UBSan/LeakSanitizer native
compare complete findings, fixes and suggestions against unchanged production Go.

| Suite | Result |
| --- | --- |
| Declarations | PASS 0.00 s |
| Continuation | PASS 77.47 s |
| Synthetic render judgments | PASS 60.62 s |
| Original three rules | PASS 71.40 s |
| Third batch | PASS 251.98 s |
| Complete owned invocation | PASS 461.469 s |
| Bridge | PASS 117.276 s |
| Checker package | PASS 0.573 s |
| Filtered uncached Node | PASS 1.883 s |

Supported controls retain 33 original findings, 15 continuation findings, 18
synthetic render findings, 60 third-batch findings, 58 constructor findings and
36 supported-pattern findings. Both frozen corpora agree in each suite under
sanitizers. The new inherited-static-field fixture agrees with Node and emitted
JavaScript. No cache hits occurred in the filtered Node gate.

Exact owned mutant observations:

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
    wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 99
    wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4659
    wave_14_next_test.go:167: released-registry mutant exits 0, required panic catches it
    wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 55
    wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 57
    wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1546
    wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4376
    wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
    wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 59
    wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8710
    wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 463
    wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10280
    wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6690
    wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12805
    wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 21991
    wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 54
    wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5164
    wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5767
    wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 18113
    wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 6923
    wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1362
    wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4544
```

Semantic mutants compile, exit 0 with empty stderr and fail only complete byte
comparison. Declaration mutants are metadata assertions; registry-retention
mutants are separately caught by the required released-handle panic-70 check.
The complete bridge gate repeats its seven length/type/link/ownership mutants:
ASan catches lengths, the byte oracle catches wrong type positions, stale-handle
and link refusal assertions catch their mutants, and LeakSanitizer catches
missing output frees and unowned region allocations. The filtered Node gate
also catches its one-byte oracle mutant. Vet and whitespace checks are clean.

Quiet alternating three-round whole-process count medians, after tests:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.796873 | 0.352519 | 5.10x |
| repository | 0.234519 | 0.129435 | 1.81x |
| controls | 0.021385 | 0.028811 | 0.74x |
| constructors | 0.025871 | 0.031117 | 0.83x |
| upstream | 0.027391 | 0.036499 | 0.75x |
| patterns | 0.029790 | 0.039835 | 0.75x |

Compilation and sanitizers are excluded. Constructors/upstream use --class-only;
the upstream row confirms count 89, not a new full-byte upstream comparison.
No driver speed improvement is claimed.

The clarification is accepted: new rule.json kinds use validated ast.Kind names,
not numeric values, and new rules use handed-node visits. The historical numeric
metadata verifier remains evidence for its old declaration contract; numeric
API availability is no longer a blocker for new claims. There are no new rules
to declare here. ab70f38d4 is still not on current main. No developer-tools
leak-helper migration appeared in this main advance; none was reverted.

Prior boundaries remain explicit, not completed ports: native runtime RegExp
construction and Go-compatible pattern errors, real JSX parsing and undefined
labels. No custom regex matcher was added or extended. Synthetic JSX comparisons
prove only the implemented judgments. Leaked-render is not falsely parked on
IR, single assignment or capture analysis. The full repository gate, nondefault
options, complete pattern grammar and end-to-end JSX parity are not covered.
The additional 741-constructor matrix was not rerun in this resume; its previous
full-byte evidence remains in WAVE_14_RESUME_REPORT.md. No work is claimed on
main or on any area branch. With the ranked pool exhausted, selection stops.
