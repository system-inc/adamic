Rebased wave 14 onto updated area and main typeof fixes; no new claims.
Tested source 9fb3da437ce7c26ec7b6dab6d9a89cd49f675984 includes area d3a37422c and main b6b1538b0.
Owned oracle PASS 490.484s; uncached typeof Node gate PASS 12.241s; registry and vet exit 0.
Twenty-one semantic byte mutants, nine metadata mutants and released-registry mutants caught.
No unclaimed ranked rules remain; regex validation, undefined labels and full-stage coverage remain incomplete.

No owned rule implementation or shared compiler/harness files changed. The previous owned
remote was ec8a9695e50d7ea3c7f37edec7f879d876b2f2d8. Only the owned branch is published.

Commands, with /workspace/adamic-tools/env.sh sourced:

- `go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m`: PASS 490.484s.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTypeof|^TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_.*\.a$' -count=1 -v -timeout=10m`: PASS 12.241s, native misses 21 and Node misses 14, no cache hits.
- `go vet ./...` and `go run ./cmd/lint-registry`: exit 0.

Owned tests explicitly provide ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript and
compiler/repository manifests under /workspace/wave-14-artifacts. Scratch paths reuse
/workspace/wave-14-registry-{original,next,render,third} to avoid duplicate compiled artifacts.
Historical evidence is committed separately and unchanged. Fresh raw outputs are gzip retained
without loss in validation-wave-14-typeof. Supported findings, fixes and suggestions match
production Go cohere byte for byte; ASan/UBSan/LSan and ownership checks pass. No test skipped
in either retained run. Actual JSX still reports thirteen identical findings and catches its
separate mutant. Default options only: nondefault options were not certified. No options
adapter guard was relaxed or bypassed.

Fresh mutation observations:

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
wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 97
wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4649
wave_14_next_test.go:159: released-registry mutant exits 0, required panic catches it
wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 54
wave_14_render_test.go:137: real JSX mutant exits 0 with empty stderr; byte comparison catches byte 60
wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 56
wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1541
wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4362
wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 58
wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8695
wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 461
wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10257
wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6680
wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12772
wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 21938
wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 53
wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5142
wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5742
wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 18061
wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 6895
wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1356
wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4473
```

Semantic mutants compile, exit zero and have empty stderr; byte comparisons alone catch them.
Legacy numeric declaration assertions are separate metadata tests, not a numeric registry
requirement. New rule manifests use named ast.Kind listeners and handed-node visits.

Quiet alternating three-round whole-process count medians after tests completed:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.742286 | 0.333580 | 5.22x |
| repository | 0.248154 | 0.144029 | 1.72x |
| controls | 0.025012 | 0.037576 | 0.67x |
| constructors | 0.026946 | 0.035413 | 0.76x |
| upstream | 0.036376 | 0.035853 | 1.01x |
| patterns | 0.033502 | 0.039769 | 0.84x |

These exclude compilation/sanitizer costs. Constructors/upstream use --class-only; upstream
checks count 89, not a fresh full-byte upstream comparison. Commands and outputs are retained.
The all-origin audit covers 658 refs and 33 distinct claim blobs; zero ranked candidates remain
before further main port exclusion. No additional rules were reserved.

Remaining Go-positive boundaries rerun here: `(` gives Go invalidRegexp but native NotYet
panic 70; `undefined:` gives Go identifierClashWithLabel but native parser NotYet panic 70.
Dynamic new RegExp(pattern, 'u') and compatible validation errors remain shared compiler
limitations described in WAVE_14_REGEX_CONTRACT_REPORT.md. No custom matcher was added or
extended. Real JSX is supported and needs no parked high-level React analysis.

The full repository gate and 17 required external stage1 comparisons were not run. The bridge
suite, broader Node fixture filter, standalone 741-constructor matrix, every upstream JSX
project, nondefault options and new-rule emitted-JavaScript gate were not repeated or newly
certified. Earlier bridge evidence remains historical in WAVE_14_LATEST_REPORT.md. No test,
skip requirement, or correctness guard was deleted or weakened. Setup was not repeated in
this refresh; the existing toolchain environment was sourced.
