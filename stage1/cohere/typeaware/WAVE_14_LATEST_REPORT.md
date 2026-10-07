Rebased all owned wave-14 work onto the requested area and current main; no new rules claimed.
Tested source a6e9ab6ba205f24a27f7f7df3be3052b7e55afea, area b84a9d9314b65d3d0261ee017e233287b4f071da, main c7991b900362796aefd111474e65eb5398e91953.
Owned byte oracle PASS 474.949s; bridge, filtered uncached Node, new proven fixtures, registry and vet PASS.
Twenty-one semantic mutations caught by byte comparisons; nine metadata and three released-registry mutants caught separately.
No unclaimed ranked rules remain; complete regex validation, undefined labels and the full-stage gate remain uncovered.

This is a landing refresh, with no owned implementation changes. Only the owned branch is pushed;
integration owns main and area branches. The previous published branch was d7be6e58118c27dbd9ce4bcb0a4b319ce5812be3.
All nine claimed rules and prior implementations remain on this branch; no claim was added.

Commands and observations:

- `bash cloud/setup.sh`: Go, clang, Node and submodules each 0s; cache 138s; total 138s; `nproc` 5.
- `go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m`: PASS 474.949s.
- `go test ./bridge/tsgo/... -count=1 -v -timeout=15m`: PASS bridge 154.280s, checker 0.752s.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run` with the exact Node filter retained in the prior area refresh report: PASS 2.321s. It covers one-byte oracle mutation, maps/text, sorting, indexing, lone surrogates, functions, closures, inherited static field reads and runtime lastIndexOf.
- Additional uncached `TestNativeAgreesWithNode` filter `proven_(assertions|class_guards|guards|satisfies|upcasts)`: PASS 33.137s, all five newly landed fixtures.
- `go vet ./...` and `go run ./cmd/lint-registry`: exit 0.

The owned run explicitly supplies `/workspace/wave-14-typescript` and the compiler/repository manifests
under `/workspace/wave-14-artifacts`. ASan/UBSan/LSan comparisons and released-handle assertions pass.
There are no skipped tests in these retained logs. This targeted gate does not certify the full repository
or the 17 newly required stage1 external comparisons; those were not run. No check was relaxed or removed.

The exact mutant observations, including first differing bytes, are retained below. Semantic variants
compile and exit zero with empty stderr: comparison against production Go cohere is what catches them.
The nine numeric metadata tests are legacy declaration assertions, not evidence of dispatch performance;
new registry declarations use named ast.Kind listeners and handed-node visits as instructed.

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
wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 93
wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4629
wave_14_next_test.go:159: released-registry mutant exits 0, required panic catches it
wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 52
wave_14_render_test.go:137: real JSX mutant exits 0 with empty stderr; byte comparison catches byte 58
wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 54
wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1531
wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4334
wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 56
wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8665
wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 457
wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10211
wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6660
wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12706
wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 21832
wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 51
wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5098
wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5692
wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 17957
wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 6839
wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1344
wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4331
```

Bridge additionally catches input/output length mutants with ASan, retained released handles with an
assertion, wrong type positions with byte equality, absent C linking with refusal, and missing C frees
and unowned heap regions with LSan. The Node oracle catches its deliberate one-byte mutation.

Quiet alternating three-round whole-process count medians after all checks finished:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.627371 | 0.319315 | 5.10x |
| repository | 0.237193 | 0.130254 | 1.82x |
| controls | 0.021968 | 0.029530 | 0.74x |
| constructors | 0.024094 | 0.028485 | 0.85x |
| upstream | 0.024286 | 0.033903 | 0.72x |
| patterns | 0.027215 | 0.042999 | 0.63x |

Compilation and sanitizer time are excluded. Constructor and upstream populations use `--class-only`;
upstream verifies count 89, not a fresh complete upstream byte comparison. No driver speedup is claimed.

`validation-wave-14-latest/selection.json` retains all 636 fetched origin refs and 33 distinct Markdown
claim blobs. The conservative claim audit leaves zero candidates before additional main port exclusion.
Original claim ordering and historical evidence remain in the earlier reports.

Real JSX leaked-number-render comparisons continue to pass, including thirteen actual JSX findings and
a separate real-JSX mutant. It does not need the parked React analysis modules. Remaining exact boundaries:
`no-invalid-regexp` broader patterns such as `(` produce native NotYet panic 70 while Go reports invalidRegexp;
shared lowering still rejects nonconstant `new RegExp(pattern, 'u')`, and native-compatible Go validation
errors are unavailable. No hand-rolled matcher was added or extended. The translated regex table has no
row for these owned rules. `no-label-var` on `undefined:` also produces explicit parser NotYet panic 70
while Go reports identifierClashWithLabel. Both reproductions are rerun in this gate. Nondefault options,
all upstream JSX projects, complete regex grammar, the standalone 741-constructor matrix, and new-rule
emitted-JavaScript comparison were not newly certified. Shared harness/compiler files were not edited.

Raw comparison, sanitizer, mutation and refusal streams are losslessly gzip-retained beside the gate
logs; benchmark commands, outputs and timings are retained under `benchmark/`.
