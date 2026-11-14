# STOP: diagnostic read-view amendment is not accepted

Candidate: 18ae2912bd4adca63ca8bbd6952736fcdc996b95; parent: 8e3a784aa5a601f57342c07b68f145fe56252c21. One commit on the requested base.

The candidate and evidence are published under the requested branch names for review only. No Loom submission. The 0-new/0-changed gate fails; no successful final apply or whole-lane result is claimed.

## Full censuses

Every run used LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE, CENSUS_MODULE_FORMAT=CommonJSBundle and a 6000-second bound. All eight main/candidate runs and the full restoring-mutant program run completed normally. Complete streams include the final relation_report record.

| File | Main unique | Candidate unique | New | Changed | Removed | Main seconds | Candidate seconds |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| program.ts | 528 | 526 | 0 | 0 | 2 | 2489.779 | 2603.708 |
| utilitiesPublic.ts | 217 | 217 | 0 | 0 | 0 | 561.300 | 655.084 |
| utilities.ts | 984 | 994 | 10 | 2 | 0 | 2186.773 | 2199.196 |
| types.ts | 0 | 0 | 0 | 0 | 0 | 539.232 | 387.005 |

Keys are phase, kind, mapped where and reason. Checkout roots and embedded source locations in reasons are mapped; no type names or other reason text are normalized. The outer-location-only report retains the two path/line-only program differences. Both original program refusal_scan sites, 2796:16 and 2829:16, disappear.

## Why this stops inside the requested shape

The optional-widening findings at mapped utilities.ts:8808:39 and :8808:44 change their reason target from Diagnostic to DiagnosticReadView. All allowed comparator parameter substitutions preserve DiagnosticRelatedInformation recursive arguments without reportsUnnecessary and the DiagnosticReadView target that has it. Removing or reclassifying these changed findings is outside the specified shape and strict pairing rule.

The related-information array field and recursive expression compareDiagnostics(d1i, d2i) do not change under the permitted parameter substitutions. A read-only stock-checker counterfactual with all eight comparator signatures confirms identical argument and target types and missing/present reportsUnnecessary properties. This explains the full-census result; it is not another census. See stop-type-facts.json.

Only compareDiagnostics and diagnosticsEqualityComparer were changed in the measured candidate. The ten new readonly-to-writable findings identify compareDiagnosticsSkipRelatedInformation, compareRelatedInformation, getDiagnosticCode, getDiagnosticMessage and getDiagnosticFilePath as further callees needing the view if work continued. No further signatures, members, call expressions, overloads or diagnostic policy were changed after the stop.

## Every new finding in full

### utilities.ts, new 1
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8784:53.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8784:53: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 2
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8784:57.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8784:57: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 3
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8785:35.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8785:35: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 4
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8785:39.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8785:39: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 5
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8924:37.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8924:37: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 6
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8925:37.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8925:37: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 7
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8926:39.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8926:39: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 8
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8927:39.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8927:39: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 9
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8928:62.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8928:62: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

### utilities.ts, new 10
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8928:89.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8928:89: Adamic 0.1 refuses a value of type DiagnosticReadView seen as Diagnostic, whose readonly field file becomes writable: a readonly field may hold something narrower than SourceFile | undefined, which a write of SourceFile | undefined would replace; keep file readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)

## Every changed finding in full

### utilities.ts, changed 1
Mapped key: ['refusal_scan', 'Refused', 'src/compiler/utilities.ts:8808:39'].

before:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/main-tree/src/compiler/utilities.ts:8808:39.

/workspace/scratch/diag-read-view-1/main-tree/src/compiler/utilities.ts:8808:39: Adamic 0.1 refuses optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

after:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8809:39.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8809:39: Adamic 0.1 refuses optional property reportsUnnecessary in DiagnosticReadView absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

### utilities.ts, changed 2
Mapped key: ['refusal_scan', 'Refused', 'src/compiler/utilities.ts:8808:44'].

before:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/main-tree/src/compiler/utilities.ts:8808:44.

/workspace/scratch/diag-read-view-1/main-tree/src/compiler/utilities.ts:8808:44: Adamic 0.1 refuses optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

after:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8809:44.

/workspace/scratch/diag-read-view-1/candidate-tree/src/compiler/utilities.ts:8809:44: Adamic 0.1 refuses optional property reportsUnnecessary in DiagnosticReadView absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

## Other proof results

- Restoring mutant: exact original bytes restored in all four diagnostic files, and both original refusal_scan findings returned in its full program census (2710.820 seconds). Historical resolution-file restoration is retained.
- Stock TypeScript 6.0.3: 0 diagnostics on both main and candidate over 79 roots. Adamic types: exit 0, 0 diagnostics over 79 roots.
- Upstream npm ci and node_modules/.bin/hereby dts completed in each adapted checkout. typescript.d.ts and tsserverlibrary.d.ts are byte-identical; hashes are in declarations.json.
- All 604 JavaScript-emitting sources are byte-identical under transpilation in both comment modes. Both upstream compiler builds passed; all 9 actual JavaScript outputs are byte-identical.
- Drift, duplicate, partial-tree, mixed-line-ending, LF/CRLF, exact-restoration and zero-write replay controls passed with final owner routing.
- The resolution engine, mutant and original pin entry remain byte-identical. Program owner amendment: 95-checker-axolotl-h. Whole utilities import amendment: 99-correctness-core-multimap, with exact input/output variants.
- All 29 sanctioned API proof inputs are unchanged. No sanctioned-api.json, shard-plan runtime hash or generated shard test regeneration is required. No ledger line was added.
- First full candidate apply stopped at the later whole import owner. Its amendment was then validated at that owner boundary (17 historical runtime files, 40 historical edits). Final fresh whole lane was cancelled during apply on the census stop, before oracle tests. Final normal apply and 106366 passing are NOT proven.
- Additional stage-specific final-tree replay checks preserve existing refusals: 95 rejects identical normalized baseline/candidate createProgram hashes after later adapters; MultiMap rejects an unchanged later recursive-watcher sys owner. The amended utilities import owner reports adapted=true and zero edits. See the replay JSON records. No unrelated owner gate was widened.

See STOP.json, pair-candidate.json, controls.json, declarations.json, source-manifests.json, every result.json and the complete jsonl.gz streams.
