Held group seventeen's nine array pairs against complete original declarations; no compiler change was needed.
Commits: follows group sixteen 9ce75693; this report and fixtures are in the next group commit on codex/views-arrays-callables-parser.
Validation: 59 probes pass Node, sanitized/release native and JavaScript; 30 finishing probes pass leaks; 59 new count rows measured; vet passes.
Mutants: native and JavaScript numeric-field bypasses caught by signature-type-parameters-wrong-pos; restored witness and both GenericType missing-field probes pass.
Limits: runtime reachability unmeasured; tuples and union-target casts skipped; integrator owns graph_regions_regression_06.a and process_exit.a baseline counts failures.

| Original pair | Type id | Static reads | Original declaration |
| --- | ---: | ---: | --- |
| JSDocTemplateTag.comment | 7637 | 4 | `string \| NodeArray<JSDocComment> \| undefined` |
| JSDocTypeLiteral.jsDocPropertyTags | 7643 | 4 | `readonly JSDocPropertyLikeTag[] \| undefined` |
| JSDocSignature.typeParameters | 7664 | 4 | `readonly JSDocTemplateTag[] \| undefined` |
| IntersectionTypeNode.types | 7759 | 4 | `NodeArray<TypeNode>` |
| GenericType.localTypeParameters | 8599 | 4 | `TypeParameter[] \| undefined` |
| GenericType.outerTypeParameters | 8599 | 4 | `TypeParameter[] \| undefined` |
| JSDocCallbackTag.comment | 8813 | 4 | `string \| NodeArray<JSDocComment> \| undefined` |
| NewExpression.typeArguments | 8828 | 4 | `NodeArray<TypeNode> \| undefined` |
| JSDocSeeTag.comment | 8875 | 4 | `string \| NodeArray<JSDocComment> \| undefined` |

Original revision is microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The declaration adapter emits 78 full original files and validates all original member types and 36 source spans. The manifest pins digests and complete receiver field lists; the oracle verifies those full lists in emitted view contracts, plus TypeParameter wherever read. JSDocComment and JSDocPropertyLikeTag retain the original tagged unions. No cohere code is copied.

Every pair covers present arrays, unread wrong arrays, reached wrong array kind, descendant scalar kind, descendant readiness, and an unread malformed second element. JSDoc comment pairs also cover strings, and optional properties cover missing and explicit undefined. GenericType.localTypeParameters and .outerTypeParameters are required nullable fields: explicit undefined passes, while missing fields reject at the read with expected TypeParameter[] | undefined. Separate missing-array fixtures pin both rejections. Complete declarations keep callable text members unread; these probes do not credit text reads.

The initial fixture run exposed two incorrect assumptions: length requires narrowing the original string-or-array union, and required nullable properties do not admit missing fields. The fixtures now narrow before length and keep missing GenericType properties as rejection probes. The compiler's behavior was correct and was not changed.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original17/prepare.cjs /workspace/lane2-original-archive /workspace/lane2-original-declarations17 > /tmp/lane2-group17-prepare.log 2>&1
ADAMIC_ARRAY17_ORIGINAL_DECLS=/workspace/lane2-original-declarations17 go test ./internal/oracle -run '^TestCheckedViewRanked17OriginalArrays$' -count=1 -v -timeout 10m > /tmp/lane2-group17-oracle-corrected.log 2>&1
ADAMIC_ARRAY17_ORIGINAL_DECLS=/workspace/lane2-original-declarations17 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lane2-group17-required-counts.log 2>&1
ADAMIC_ARRAY17_ORIGINAL_DECLS=/workspace/lane2-original-declarations17 python3 stage3/interface-downcasts/lane2/original17/run-mutants.py > /tmp/lane2-group17-mutants.log 2>&1
ADAMIC_ARRAY17_ORIGINAL_DECLS=/workspace/lane2-original-declarations17 go test ./internal/oracle -run '^TestCheckedViewRanked17ArrayCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane2-group17-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-group17-vet.log 2>&1
ADAMIC_ARRAY17_ORIGINAL_DECLS=/workspace/lane2-original-declarations17 go test ./internal/oracle -run '^TestCheckedViewRanked17OriginalArrays$/^(signature-type-parameters-wrong-pos|generic-local-missing-array|generic-outer-missing-array)$' -count=1 -v -timeout 10m > /tmp/lane2-group17-restored.log 2>&1
```

Corrected complete oracle: PASS in 121.934s. Group counts: PASS in 62.757s. Restored witnesses: PASS in 5.308s. Vet: PASS, no output. Tests and mutants write logs, retained under original17/evidence. No whole-package tests or full gate run.

Native numeric-field mutant returns a numeric slot without checking stored kind. Both native executions exit zero and print unchecked numeric values instead of rejection. JavaScript numeric-field mutant makes the predicate unconditional and prints bad, exiting zero. Both are executed semantic kills, not compile failures; temporary source edits are restored in finally and verified afterward.

The required global counts refresh failed in 76.477s before the lane registry and wrote no table changes. Its log names internal/oracle/testdata/graph_regions_regression_06.a (invalid pointer free) and internal/oracle/testdata/process_exit.a (unsupported process.exit value lowering), among other failures. Both named failures reproduce at baseline 0f47b23c, with evidence in original15/evidence/counts-baseline.log. Per the user, the integrator owns them and lane 2 continues. We wrote only the 59 measured rows for this group; removing them yields the prior counts table unchanged. Declaration-dependent tests skip without generated inputs, and the count registry explicitly retains historical rows without claiming remeasurement when inputs are absent.

| Family | Candidate pairs / reads | Cumulative fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 164 / 2777 | 170 / 412 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Cumulative array obligations include inherited representative probes and nineteen pairs using complete original declarations in groups fifteen through seventeen. All read counts are static candidates. Continue at JsonSourceFile.parseDiagnostics, TsConfigSourceFile and lookup-location arrays. Toolchain setup and integration provenance are in group fifteen's report.
