Certified four original tagged-union array pairs with complete receiver and element declarations; no compiler change.
Commits: follows group 20 d03c7e15 on codex/views-arrays-callables-parser; the group commit is recorded in the push output.
Validation: 101 probes against Node, sanitized native, release native and JavaScript; 81 finishing leak checks; 101 measured count rows after restoring mutants; vet passes.
Mutants: native and JavaScript numeric-field bypasses caught for heritage, type parameters, JSX type arguments and statements, eight executed guard mismatches; restored witnesses pass.
Limits: lane 4c untagged unions and tuples are skipped; production execution is not measured; required global counts still fail in existing integration fixtures.

| Original pair | Static reads |
| --- | ---: |
| ClassLikeDeclaration.heritageClauses | 6 |
| DeclarationWithTypeParameterChildren.typeParameters | 6 |
| JsxOpeningLikeElement.typeArguments | 6 |
| CaseOrDefaultClause.statements | 5 |

The preparation adapter validates every exact original source span and the full declared member type at pristine microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. It emits and hashes all 78 complete original declaration files. Fixtures hold all 25 receiver arms, complete selected HeritageClause, TypeParameterDeclaration, TypeNode and Statement field sets, optional absent/undefined arrays, lazy array and descendant reads, invalid arrays and elements, numeric kind/readiness rejections and map consumers. Initial verification found a fixture error: JSDocTemplateTag has a required typeParameters array, but a read through the entire union is still optional. The fixture now narrows the full union's optional read before selecting its element. Corrected full oracle passes in 108.364s.

Commands, with output redirected into evidence logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original21/generate.cjs /workspace/lane2-original-pin
node stage3/interface-downcasts/lane2/original21/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations21
export ADAMIC_ARRAY21_ORIGINAL_DECLS=/workspace/lane2-original-declarations21
go test ./internal/oracle -run '^TestCheckedViewRanked21OriginalArrays$' -count=1 -v -timeout 15m
python3 stage3/interface-downcasts/lane2/original21/run-mutants.py
go test ./internal/oracle -run '^TestCheckedViewRanked21ArrayCounts$|^TestCheckedViewRanked21OriginalArrays$/^(heritage-264-wrong-pos|type-parameters-264-wrong-pos|jsx-types-286-wrong-pos|statements-297-wrong-pos)$' -count=1 -v -timeout 15m -args -update-counts
go vet ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

Eight mutants return unchecked numeric descendants in native or accept them in JavaScript. Each executed witness fails on stderr mismatch against the required exit-70 rejection, without compilation failures. Mutant files are restored in finally; all count rows are remeasured with restored sources, and four restored witnesses pass. Removing this group's rows leaves the previous table byte for byte. The required global refresh failed in 46.793s in existing node_fs, graph_regions, fresh_refused, process and census fixtures and changed no rows. Complete evidence is under original21/evidence; restored-counts.log holds the final measured refresh and restored witnesses. No full package or full gate was run.

Fixture obligations now hold 178 pairs / 2887 reads of 334 / 3189; remaining 156 / 302. Consumer and intrinsic production totals remain uncredited, and own array fields remain 3 / 26. IncrementalBuildInfo.fileNames remains a lane 4c admission skip; InterfaceType.resolvedBaseTypes is the previously recorded untagged element union skip. Tuples remain assigned elsewhere. The next SignatureDeclaration | JSDocSignature.parameters candidate passes an initial probe and is prepared for group 22; FileWatcherWithModifiedTime.callbacks remains an uncredited callable candidate pending a complete original private-declaration probe.

Toolchain is reused from setup at the merged tip (43.041s, nproc 5, CPU quota 4): Go 1.27.1, clang 20.1.8, Node 24.19.0, /workspace/adamic-tools/env.sh. Integration 432d4913 is already merged. No cohere source or reduced declarations are copied.
