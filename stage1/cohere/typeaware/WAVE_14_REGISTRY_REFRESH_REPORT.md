Rebased wave 14 onto the current area registry migration; no new rules claimed.
Tested source a89cd2eecdf5001e6deb02c9ef10f9c06f37751c contains area b46914832 and main c7991b900.
Owned byte oracle PASS 480.336s; registry validation and Go vet exit 0.
Twenty-one semantic byte mutants, nine metadata mutants and released-registry mutants caught.
No unclaimed ranked rules remain; regex validation, undefined labels and the full-stage gate remain uncovered.

The updated area migrates legacy lint rules to the shared registry. No owned implementation,
shared harness, registration generator or compiler files were edited. Only the owned branch
is published; integration owns main and area. The exact previous remote lease was 93919b169.

Ran with `/workspace/adamic-tools/env.sh` sourced:

- `go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m`: PASS 480.336s.
- `go run ./cmd/lint-registry`: exit 0.
- `go vet ./...`: exit 0.

The owned test provides ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript, original and
next compiler/repository manifests under /workspace/wave-14-artifacts, and separate artifact
roots /workspace/wave-14-registry-{original,next,render,third}. Comparison streams are gzip
retained without loss in validation-wave-14-registry. Findings, fixes and suggestions on
supported inputs compare byte for byte against production Go cohere. ASan/UBSan/LSan and
released-handle checks pass. Real JSX still reports thirteen matching findings; its separate
mutation is caught. No test in this retained owned run skipped.

Exact mutant observations:

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

Semantic mutants compile and exit zero with empty stderr; only byte comparison catches them.
The numeric metadata assertions are legacy tests, not a new numeric-kind requirement or a
driver performance claim. New declarations use the named ast.Kind registry contract.

Quiet alternating three-round count medians, excluding compilation and sanitizer costs:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.667424 | 0.339132 | 4.92x |
| repository | 0.242381 | 0.142953 | 1.70x |
| controls | 0.022177 | 0.031498 | 0.70x |
| constructors | 0.025379 | 0.031374 | 0.81x |
| upstream | 0.025408 | 0.033482 | 0.76x |
| patterns | 0.029835 | 0.043611 | 0.68x |

Constructor and upstream rows use --class-only. Upstream counts 89 findings; its full byte
comparison was not newly repeated. Benchmark commands and raw output are retained.
The all-origin audit includes 644 refs and 33 distinct claim blobs, leaving zero candidates
before additional main port exclusion; see selection.json. No new claim was made.

Remaining Go-positive boundaries were rerun: `(` gives Go invalidRegexp while native refuses
with NotYet panic 70; `undefined:` gives Go identifierClashWithLabel while native parser refuses
with panic 70. Nonconstant new RegExp(pattern, 'u') and Go-compatible validation errors remain
shared compiler limitations documented in WAVE_14_REGEX_CONTRACT_REPORT.md. No custom matcher
was added or extended. No React analysis parking is needed for the owned JSX rule.

The full repository gate and 17 required external stage1 comparisons were not run. Neither
were the bridge suite, filtered Node fixtures, standalone 741-constructor matrix, nondefault
options, all upstream JSX projects or new-rule emitted-JavaScript gate repeated in this refresh.
Prior bridge and Node evidence remains in WAVE_14_LATEST_REPORT.md; it is not a fresh run here.

Automatic approval review rejected broad cleanup of prior owned compiled scratch directories
as an unacceptable deletion risk. No files were removed; sufficient space remained and the
oracle completed successfully. No action remains blocked by that rejection.
