u085 starts at origin/main 12e77e8972a2e606cab6db05d84f428246a85339; all 12 names remain in the named files.
All 36 isolated clean runs pass with both original-library rows enabled; nproc 5.
The two original-library wrappers form one family, giving 11 judged rows.
Four fixed production mutants and one empty-entry probe; matrix bounded to reached port rows.
Evidence: test-audit/stage1-cohere-estree-deep, review/test-audit/stage1-cohere-estree-deep/.

```json
[
  {
    "test": "TestDeepGrammar",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/deep_test.go",
    "seconds": 69.442,
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: deep_test.go:38: Node: line 25: Go \"2 .operator string +\", port \"2 .left node\"",
    "subsumed_by": [
      "TestGeneratedAgreement"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 63.504,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDeepGrammar",
      "TestGeneratedAgreement",
      "TestDecoratedExports",
      "TestRecoveredExpressions",
      "TestUnattachedDecorator"
    ],
    "oracle_kind": "external-run",
    "oracle": "Unmodified Go cohere ESTree API through testdata/oracle.go. Exact canonical bytes are compared with the TypeScript port on source Node, sanitized native and emitted JavaScript.",
    "mutants_run": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "verdict": "subsumed",
    "subsumption_kills": 2,
    "probe_status": "fail",
    "evidence": "selector file /tmp/u085/selector=M2; ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestDeepGrammar$ > review/test-audit/stage1-cohere-estree-deep/M2-TestDeepGrammar.log 2>&1; deep_test.go:38: Node: line 25: Go \"2 .operator string +\", port \"2 .left node\""
  },
  {
    "test": "TestGeneratedAgreement",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/estree_test.go",
    "seconds": 63.504,
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: estree_test.go:186: source Node: line 2432: Go \"1 ExportDefaultDeclaration 32 62 0 0 0 32 62 0\", port \"1 ExportNamedDeclaration 32 62 0 0 0 32 62 0\"",
    "subsumed_by": [
      "TestDecoratedExports",
      "TestDeepGrammar",
      "TestRecoveredExpressions"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDeepGrammar",
      "TestGeneratedAgreement",
      "TestDecoratedExports",
      "TestRecoveredExpressions",
      "TestUnattachedDecorator"
    ],
    "oracle_kind": "external-run",
    "oracle": "Unmodified Go cohere ESTree API through testdata/oracle.go. Exact canonical bytes are compared with the TypeScript port on source Node, sanitized native and emitted JavaScript.",
    "mutants_run": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "verdict": "overlapping",
    "subsumption_kills": 3,
    "probe_status": "fail",
    "evidence": "selector file /tmp/u085/selector=M3; ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestGeneratedAgreement$ > review/test-audit/stage1-cohere-estree-deep/M3-TestGeneratedAgreement.log 2>&1; estree_test.go:186: source Node: line 2432: Go \"1 ExportDefaultDeclaration 32 62 0 0 0 32 62 0\", port \"1 ExportNamedDeclaration 32 62 0 0 0 32 62 0\""
  },
  {
    "test": "TestOriginalLibraries family",
    "package": "stage1/cohere/estree",
    "file": [
      "stage1/cohere/estree/estree_test.go",
      "stage1/cohere/estree/exports_test.go"
    ],
    "seconds": 2.072,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "members": [
      "TestOriginalLibraries",
      "TestDecoratedExportLibraries"
    ],
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "oracle": "Go cohere compared with pinned @typescript-eslint/typescript-estree 8.65.0 and Prettier 3.9.6 using TypeScript 6.0.3; exact handwritten deltas permit three documented whitespace gaps in the generated corpus.",
    "verdict": "cannot-judge",
    "reason": "No Adamic port code executes. A meaningful mutation would change one of the external oracles, which the brief prohibits.",
    "evidence": "ADAMIC_ESTREE_LIBRARY=/tmp/u085/library go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^(TestOriginalLibraries|TestDecoratedExportLibraries)$; three family runs pass, no failing line produced"
  },
  {
    "test": "TestDecoratedExports",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/exports_test.go",
    "seconds": 2.64,
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: exports_test.go:34: stage1/cohere/estree/exports_test.go:export:source-76710bcbab44ce4b799cae210ad896f1086115ab253caea159c7c662325372fb Node: line 3: Go \"1 ExportDefaultDeclaration 3 28 0 0 0 0 28 0\", port \"1 ExportNamedDeclaration 3 28 0 0 0 0 28 0\"",
    "subsumed_by": [
      "TestGeneratedAgreement"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 63.504,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDeepGrammar",
      "TestGeneratedAgreement",
      "TestDecoratedExports",
      "TestRecoveredExpressions",
      "TestUnattachedDecorator"
    ],
    "oracle_kind": "external-run",
    "oracle": "Unmodified Go cohere ESTree API through testdata/oracle.go. Exact canonical bytes are compared with the TypeScript port on source Node, sanitized native and emitted JavaScript.",
    "mutants_run": [
      "M1",
      "M3",
      "M4"
    ],
    "verdict": "subsumed",
    "subsumption_kills": 2,
    "probe_status": "fail",
    "evidence": "selector file /tmp/u085/selector=M3; ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestDecoratedExports$ > review/test-audit/stage1-cohere-estree-deep/M3-TestDecoratedExports.log 2>&1; exports_test.go:34: stage1/cohere/estree/exports_test.go:export:source-76710bcbab44ce4b799cae210ad896f1086115ab253caea159c7c662325372fb Node: line 3: Go \"1 ExportDefaultDeclaration 3 28 0 0 0 0 28 0\", port \"1 ExportNamedDeclaration 3 28 0 0 0 0 28 0\""
  },
  {
    "test": "TestDecoratedExportsPlantedDisagreement",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/exports_test.go",
    "seconds": 0.03,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: exports_test.go:42: planted failure must be caught only by TestDecoratedExportsPlantedDisagreement/shard-009: exit=<nil>",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "oracle_kind": "self",
    "oracle": "Planted agree/disagree answers and exact owning-shard failure are handwritten.",
    "verdict": "witness",
    "witness_checks": [
      "W1"
    ],
    "evidence": "ADAMIC_ESTREE_LIBRARY=/tmp/u085/library timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^(TestDecoratedExportsPlantedDisagreement|TestDecoratedExportMutantPlantedSurvivor)$ > review/test-audit/stage1-cohere-estree-deep/W1-planted.log 2>&1; exports_test.go:42: planted failure must be caught only by TestDecoratedExportsPlantedDisagreement/shard-009: exit=<nil>"
  },
  {
    "test": "TestDecoratedExportMutant",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/exports_test.go",
    "seconds": 2.191,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: exports_test.go:73: stage1/cohere/estree/exports_test.go:export-mutant:source-077658c3c7c53cf94a025164758bd4b02ac60dca27b566c0ba2ca2a7d0db9189 native: mutant survived",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "oracle": "Go cohere bytes and expected disagreement from a built-in source mutant; production mutation failures do not establish the witness.",
    "verdict": "witness",
    "witness_checks": [
      "W1"
    ],
    "evidence": "ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestDecoratedExportMutant$ > review/test-audit/stage1-cohere-estree-deep/W1-TestDecoratedExportMutant.log 2>&1; exports_test.go:73: stage1/cohere/estree/exports_test.go:export-mutant:source-077658c3c7c53cf94a025164758bd4b02ac60dca27b566c0ba2ca2a7d0db9189 native: mutant survived"
  },
  {
    "test": "TestDecoratedExportMutantPlantedSurvivor",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/exports_test.go",
    "seconds": 0.026,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: exports_test.go:81: planted failure must be caught only by TestDecoratedExportMutantPlantedSurvivor/shard-023: exit=exit status 1",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "oracle_kind": "self",
    "oracle": "Planted agree/disagree answers and exact owning-shard failure are handwritten.",
    "verdict": "witness",
    "witness_checks": [
      "W1"
    ],
    "evidence": "ADAMIC_ESTREE_LIBRARY=/tmp/u085/library timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^(TestDecoratedExportsPlantedDisagreement|TestDecoratedExportMutantPlantedSurvivor)$ > review/test-audit/stage1-cohere-estree-deep/W1-planted.log 2>&1; exports_test.go:81: planted failure must be caught only by TestDecoratedExportMutantPlantedSurvivor/shard-023: exit=exit status 1"
  },
  {
    "test": "TestRecoveredExpressions",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/expressions_test.go",
    "seconds": 1.926,
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1: expressions_test.go:22: Node: line 1: Go \"0 Program 0 22 0 0 0 0 22 0\", port \"0 Program 1 22 0 0 0 1 22 0\"",
    "subsumed_by": [
      "TestDecoratedExports"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 2.64,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDeepGrammar",
      "TestGeneratedAgreement",
      "TestDecoratedExports",
      "TestRecoveredExpressions",
      "TestUnattachedDecorator"
    ],
    "oracle_kind": "external-run",
    "oracle": "Unmodified Go cohere ESTree API through testdata/oracle.go. Exact canonical bytes are compared with the TypeScript port on source Node, sanitized native and emitted JavaScript.",
    "mutants_run": [
      "M1",
      "M3",
      "M4"
    ],
    "verdict": "subsumed",
    "subsumption_kills": 1,
    "probe_status": "fail",
    "evidence": "selector file /tmp/u085/selector=M1; ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestRecoveredExpressions$ > review/test-audit/stage1-cohere-estree-deep/M1-TestRecoveredExpressions.log 2>&1; expressions_test.go:22: Node: line 1: Go \"0 Program 0 22 0 0 0 0 22 0\", port \"0 Program 1 22 0 0 0 1 22 0\""
  },
  {
    "test": "TestRecoveredExpressionMutant",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/expressions_test.go",
    "seconds": 1.674,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: expressions_test.go:36: Node mutant survived",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "oracle_kind": "external-run",
    "oracle": "Go cohere bytes and expected disagreement from a built-in source mutant; production mutation failures do not establish the witness.",
    "verdict": "witness",
    "witness_checks": [
      "W1"
    ],
    "evidence": "ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -overlay=/tmp/u085/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestRecoveredExpressionMutant$ > review/test-audit/stage1-cohere-estree-deep/W1-TestRecoveredExpressionMutant.log 2>&1; expressions_test.go:36: Node mutant survived"
  },
  {
    "test": "TestUnattachedDecorator",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/expressions_test.go",
    "seconds": 1.1,
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: expressions_test.go:51: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestUnattachedDecorator2012039104/001/input.ts]: timeout=false exit=<nil> stdout=343 stderr=",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDeepGrammar",
      "TestGeneratedAgreement",
      "TestDecoratedExports",
      "TestRecoveredExpressions",
      "TestUnattachedDecorator"
    ],
    "oracle_kind": "self",
    "oracle": "Handwritten ESTree unattached decorator diagnostic substring, failed exit, empty stdout, and 2-second CPU deadline; no external authority checked.",
    "mutants_run": [
      "M4"
    ],
    "verdict": "sacred",
    "probe_status": "fail",
    "evidence": "selector file /tmp/u085/selector=M4; ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestUnattachedDecorator$ > review/test-audit/stage1-cohere-estree-deep/M4-TestUnattachedDecorator.log 2>&1; expressions_test.go:51: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestUnattachedDecorator2012039104/001/input.ts]: timeout=false exit=<nil> stdout=343 stderr="
  },
  {
    "test": "TestUnattachedDecoratorControl",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/expressions_test.go",
    "seconds": 1.42,
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "oracle": "Go cohere audit must contain two errors; guard-disabled source Node and native must contain two Program records. Counts do not establish diagnostic identity or correct AST bytes. This proves the fixture precondition but does not execute the refusal comparison it is meant to support.",
    "verdict": "untrue",
    "witness_checks": [
      "W2"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u085/weak/W2.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestUnattachedDecoratorControl$; refusedBeforeDeadline disabled, --- PASS: TestUnattachedDecoratorControl"
  }
]
```

| id | origin/main file:line | change | failed bounded rows |
|---|---|---|---|
| M1 | stage1/cohere/estree/postprocess.ts:193 | change constant | TestDeepGrammar, TestGeneratedAgreement, TestDecoratedExports, TestRecoveredExpressions |
| M2 | stage1/cohere/estree/binaryConvert.ts:67 | drop operator-field statement | TestDeepGrammar, TestGeneratedAgreement |
| M3 | stage1/cohere/estree/convert.ts:229 | change node-kind constant | TestGeneratedAgreement, TestDecoratedExports |
| M4 | stage1/cohere/estree/pipeline.ts:49 | drop whole orphan-decorator guard loop | TestUnattachedDecorator |

Survivors: none in the observed bounded matrix. Kills outside the reached rows are unknown. No repo-wide replay was run.

Code and oracle were named before production mutation in code-and-oracle.md. The named-function index and full reached ESTree function inventory, including anonymous callbacks, per-row reach sets and compressed V8 evidence are saved. The menu was finalized before any production mutant outcome. M2 was changed from child-index substitution to dropping the operator statement before execution, to avoid quadratic recursive conversion. The menu is four mutants because this is a native port; no supplemental production mutant was used.

P1 returns an empty string from pipeline.answer. Only production rows calling that entry are probed. Witnesses and the library family have vacuous=null, because port probes do not judge their jobs. Production probes are separate from kills.

Witness W1 makes firstDifference always return an empty difference. The two planted proofs and both actual wrong-output witnesses must fail under this weakened comparison. W2 disables refusedBeforeDeadline entirely. TestUnattachedDecoratorControl still passes; it checks that a built-in guard mutant accepts two inputs, but does not apply the acceptance/refusal comparison to prove it can catch those acceptances. Its untrue verdict is specifically the brief's witness-strength result, not a claim that its fixture precondition is useless. Neither experiment changes Go cohere or the pinned libraries.

Brief ambiguities, costs and constraints:

- The historical commit 8de93800f4 differs from fetched origin/main 12e77e8972. All requested names still exist; no name moved or vanished. Every mutant/probe line and standalone diff uses the starting commit.
- The whole package cooks at 90 seconds after TestAcceptanceGrammar passes in 77.64 seconds. The requested combined slice also cooks at 90 seconds after deep grammar passes in 66.44 seconds. These are accumulated-budget timeouts, not red assertion baselines.
- Deep grammar and generated agreement each take more than 60 seconds alone. Combining reached rows in one 90-second binary would hide later rows. Each bounded matrix cell therefore runs its row alone under the same timeout. All clean rows were observed three times before production mutation.
- The older build(t) helper does not cache native products. A file-driven selector cannot prevent it rebuilding on every deep/generated invocation without a harness change. The cache-aware misc/recovery helpers share the stable switched source; no build harness was changed. The shared selector cache is correct for port-source mutants whose choice is read at runtime. Standalone validation gives each fixed source its own build cache.
- Restored-source witness products were seeded into the selector cache by copying four exact content-addressed products built in this session; keys and paths are in witness-cache-seed.json. These do not match switched-source keys and cannot supply a production mutant answer.
- The 30-minute port target competes with 36 required isolated runs, native matrix runs and standalone build checks. Native rebuilds account for most of the elapsed time; no individual observation was allowed beyond its 90-second budget.
- Two library wrappers have the same checker and differ only by input recipe and expected gap count, so they are one family. Their native-port probes are inapplicable because they execute no Adamic source. Mutating Go cohere to force them to fail would mutate the oracle prohibited by the brief.
- A built-in mutant witness is not a production coverage row. Weakening its comparison gives its verdict; other production failures or successful fixture preconditions do not. The unattached control demonstrates the distinction.
- Go cohere decides exact agreement, while the orphan-refusal diagnostic is a handwritten port string. The library family also includes handwritten whitespace exceptions; it therefore has a mixed external-run/self oracle. No outside spec diagnostic was independently checked.
- Pinned library dependencies were absent despite the warm toolchain. Installing them took 3.989 seconds. npm ci also ran in stage3/api even though these tests use plain Node without those modules.
- Fixture extraction initially treated go run's -- separator as a source filename. It was corrected before coverage or production mutation; only the successful four corpus coverage files support reach claims.
- Dropping M2's whole operator-field statement leaves operator used for choosing BinaryExpression versus LogicalExpression. Dropping M4's whole guard loop leaves no unused loop local. Standalone probe P1 replaces the complete answer body.
- Repository log ignore rules require explicit force-adding the recorded logs; replay files are stored under review and scratch Go drivers use .go.txt so normal package discovery does not compile them.

Measured phases:

```json
{
  "nproc": 5,
  "setup_seconds": 0,
  "library_install_seconds": 3.989,
  "whole_package_binary_seconds": 90.022,
  "scoped_package_binary_seconds": 90.023,
  "isolated_baseline_command_wall_seconds": 905.7241439489999,
  "isolated_baseline_binary_seconds": 794.01,
  "matrix_probe_witness_command_wall_seconds": 717.2305782750018,
  "standalone_native_builds": [
    {
      "id": "M1",
      "exit": 0,
      "wall": 68.55736592700032,
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/M1 timeout 90 go run ./cmd/adamic build stage1/cohere/estree/main.ts -o /tmp/u085/standalone-M1 --sanitize"
    },
    {
      "id": "M2",
      "exit": 0,
      "wall": 59.2336069160001,
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/M2 timeout 90 go run ./cmd/adamic build stage1/cohere/estree/main.ts -o /tmp/u085/standalone-M2 --sanitize"
    },
    {
      "id": "M3",
      "exit": 0,
      "wall": 58.5212518190001,
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/M3 timeout 90 go run ./cmd/adamic build stage1/cohere/estree/main.ts -o /tmp/u085/standalone-M3 --sanitize"
    },
    {
      "id": "M4",
      "exit": 0,
      "wall": 57.839299956000104,
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/M4 timeout 90 go run ./cmd/adamic build stage1/cohere/estree/main.ts -o /tmp/u085/standalone-M4 --sanitize"
    },
    {
      "id": "P1",
      "exit": 0,
      "wall": 27.573368993998884,
      "command": "ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/P1 timeout 90 go run ./cmd/adamic build stage1/cohere/estree/main.ts -o /tmp/u085/standalone-P1 --sanitize"
    }
  ],
  "library_family_binary_seconds": 7.004
}
```

Cold lowering and native build observations are in cold-build-observations.json; matrix-runs.json distinguishes command wall time from test-binary timing. No setup script ran because env.sh worked. Whole/scoped baselines were narrowed after 90 seconds, not allowed to continue.

Limits: bounded uniqueness only, no other package run, no repo-wide uniqueness, no external-oracle mutation, no strengthening or rewriting of production tests, no PR or main push. The original-library family cannot receive a meaningful Adamic production-mutant verdict. Production source restoration and native-diff check results are part of the evidence.

Final verification: restored clean subset passed in 6.911 s, all seven standalone diffs apply to the starting source, both Go witness overlays pass vet, and all five TypeScript mutant/probe variants compile with sanitizers. Production diff is empty and the selector helper is removed. Approximate full session: 39 minutes, beyond the 30-minute target; measured phase totals are recorded above.
