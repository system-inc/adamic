Built: rebased eighteen existing native rule ports onto the integrated lint branch; no unclaimed rule remains.
Commits: rebased implementation 5df8c53b9; integration d65a8f931; main 39638d9e2; published history c69d48e4f retained.
Commands: all owned Go agreement gates, both recent validation scripts, bridge, filtered Node oracle, registry and vet passed; setup 196s, nproc 5.
Mutants: eighteen rule mutations, three regression mutations, scope export, Unicode case, released registry, seven bridge mutations and Node one-byte caught.
Not covered: dynamic native RegExp options, whole-rule JavaScript/checker-driver integration, and three parked React analysis rules.

The shared parser now builds with the current compiler. Both owned validation
scripts copy the current integrated parser and scanner; they no longer override
those files with the previously published JSX commit. The latest listener
manifests declare typescript-go AST kind names in `kinds`; the visitors continue
to receive their cached node and use numeric kinds internally. No shared harness,
registry, parser, lowerer or emitter was edited.

All eighteen implementations were re-green on this integration snapshot. The
older twelve-rule gate passed in 717.748s, including controls, complete finding/
fix/suggestion bytes, per-rule mutants, released handles, and normal/sanitized
runs over the frozen 287 repository and 77 TypeScript compiler files. It ran as:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -v -timeout=30m
python3 stage1/cohere/typeaware/rules/wave-22-sixth/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/prove_flags.py
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(regexp\.a|inherited_static_field_read\.a)$|^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m
go run ./cmd/lint-registry
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
```

Every command's output was saved to a file. The full repository gate was not run.
The sixth batch accepted 176 sources and rejected 188 with Go's parser; default,
element-fragment and allowGlobals streams agree at 52,437, 47,114 and 51,412 bytes.
The seventh batch accepted 272 and rejected 251; default, ignoreCase and
allowAsProps streams agree at 73,462, 73,739 and 58,964 bytes. Both batches' normal
and ASan/UBSan runs match over both frozen corpora, including every proposed edit
and suggestion. The corpora have zero findings for these six rules; positive
controls and their comparison-only mutants establish that this silence can fail.

| Recent mutant | First differing byte |
| --- | ---: |
| fragment-id | 735 |
| undef-id | 358 |
| adjacent-judgment | 32497 |
| default-id | 772 |
| nested-id | 18310 |
| sort-id | 53364 |
| computed-trivia | 66117 |
| jsx-attribute-flag | 18244 |
| class-fallback-flag | 933 |

Each compiled and ran with exit 0 and empty stderr. The twelve older rule
mutants and scope-export mutation are listed with exact bytes in the saved Go
log and summary. Wrong Unicode simple case fails TestUnicodeNodeText on ßName.
Both owned checker questions reject a released program with exit 70 and the
exact invalid-or-released-handle panic under sanitizers. The shared bridge gate
passed its 1,600-position, 54,982-byte oracle under ASan/UBSan/LSan and detected
input/output length, retained handles, wrong positions, removed link guard,
missing C free and missing region cleanup mutations. The filtered external Node
oracle also caught its one-byte mutation.

| Batch / corpus | Native seconds | Go seconds |
| --- | ---: | ---: |
| Sixth / repository | 0.600 | 0.293 |
| Sixth / compiler | 4.847 | 0.866 |
| Seventh / repository | 8.757 | 0.902 |
| Seventh / compiler | 37.933 | 1.158 |

These are single process measurements while other landing checks ran, not an
isolated throughput benchmark. The native implementations remain slower. Exact
commands, timings, compiled source hashes and canonical streams are in
[landing-evidence](landing-evidence/summary.json).

Two blockers remain observable on this base. The owned dynamic-pattern probe
fails at no-unstable-nested-components/options.a:12:46 with `stage 0 can't lower
RegExp with a nonconstant pattern yet`. The required new RegExp(source, 'u')
translation remains implemented; there is no substitute matcher. Whole-rule
`adamic js` fails on the native checker call in common/unicode_node_text.a.
The shared lint RuleContext has no checker-program binding, so its new .a and
suggestion support does not make these custom type-aware drivers runnable there.
The filtered Node runtime checks do not certify the full checker-backed rules.
The three prior React HIR/SSA/capture claims remain parked under #dnv6f2c.

The fetched snapshot contains 601 origin refs and 33 distinct Markdown claim
blobs. Every one of the 197 ranked rules is either claimed or among the verified
initial bridge ports. Both previously remaining candidates are claimed by wave
18. The full inventory and current bridge-report provenance are saved in
[claim-inventory.json](landing-evidence/claim-inventory.json). No new reservation
was made. Only codex/typeaware-wave-22 is pushed; integration owns merges.
