# Diagnostic read-view v2: final proof

Implementation: d7a6aa49976828aa9627b64c137234b718c9e503. Parent: 8e3a784aa5a601f57342c07b68f145fe56252c21. One commit on the requested base. No Loom submission.

All eight reader helpers use DiagnosticReadView: compareDiagnostics, compareDiagnosticsSkipRelatedInformation, compareRelatedInformation, compareMessageText, diagnosticsEqualityComparer, getDiagnosticFilePath, getDiagnosticCode, getDiagnosticMessage. Bodies and public sorter signature are unchanged. The resolution half remains unchanged.

## Full censuses

All eight main/candidate censuses and the restoring-mutant program census completed with exit 0 within the individual 6000-second bounds. Every run used LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE and CENSUS_MODULE_FORMAT=CommonJSBundle. Complete raw streams end in relation_report.

| File | Main unique | Candidate unique | New | Changed | Removed | Main seconds | Candidate seconds |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| program.ts | 528 | 526 | 0 | 0 | 2 | 1116.986 | 1108.932 |
| utilitiesPublic.ts | 217 | 217 | 0 | 0 | 0 | 214.526 | 215.363 |
| utilities.ts | 984 | 984 | 0 | 2 | 0 | 571.296 | 572.251 |
| types.ts | 0 | 0 | 0 | 0 | 0 | 109.646 | 112.199 |

Pair keys are phase, kind, mapped where and reason. Checkout roots and source references embedded in reasons are mapped; type names are not normalized. Both program refusal_scan targets at 2796:16 and 2829:16 are removed. There are no new findings in any phase.

## Accepted changed pair: both texts

changed: target type renamed, pre-existing refusal.
Mapped key: ['refusal_scan', 'Refused', 'src/compiler/utilities.ts:8808:39'].

before:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-2/main-tree/src/compiler/utilities.ts:8808:39.

/workspace/scratch/diag-read-view-2/main-tree/src/compiler/utilities.ts:8808:39: Adamic 0.1 refuses optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

after:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilities.ts:8809:39.

/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilities.ts:8809:39: Adamic 0.1 refuses optional property reportsUnnecessary in DiagnosticReadView absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

changed: target type renamed, pre-existing refusal.
Mapped key: ['refusal_scan', 'Refused', 'src/compiler/utilities.ts:8808:44'].

before:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-2/main-tree/src/compiler/utilities.ts:8808:44.

/workspace/scratch/diag-read-view-2/main-tree/src/compiler/utilities.ts:8808:44: Adamic 0.1 refuses optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

after:
Phase: refusal_scan. Kind: Refused. Where: /workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilities.ts:8809:44.

/workspace/scratch/diag-read-view-2/candidate-tree/src/compiler/utilities.ts:8809:44: Adamic 0.1 refuses optional property reportsUnnecessary in DiagnosticReadView absent from structural source DiagnosticRelatedInformation, which can hide fields; declare reportsUnnecessary on the source type, or build a fresh object with known fields (adamic/no-optional-widening)

## Remaining proofs

- Restoring mutant: exact original diagnostic files restored; full program census returns both original refusal_scan findings. See mutant-targets.json.
- Stock TypeScript 6.0.3: 0 diagnostics on both main and candidate over 79 roots. Adamic types: 0 diagnostics over the final 79 roots.
- Both upstream npm ci, node_modules/.bin/hereby dts and compiler builds passed. typescript.d.ts and tsserverlibrary.d.ts are byte-identical. Exact hashes: declarations.json.
- All 604 emitting source files match in both JavaScript comment modes. All 9 actual upstream compiler JavaScript outputs match byte-for-byte.
- Final normal apply passed and matches the measured candidate in all 747 source files. The whole lane was run unfiltered with unchanged workers and settings; its own verdict is PASS, with 106,366 passing, only the sanctioned API row and no pending tests.
- Final controls passed: drift, duplicates, partial adaptations, mixed endings, LF/CRLF, exact restoring mutant and zero-write replay.
- Resolution engine, restoring witness and original whole-file pin are unchanged. The createProgram amendment stays in 95-checker-axolotl-h; the utilities import amendment stays in core-multimap, including exact input/output variants.
- All 29 sanctioned proof inputs remain unchanged. No sanctioned-api.json, shard-plan runtime hash or generated shard test refresh was required. No README convention required a new ledger line.

See FINAL.json, census-gate.json, pair-candidate.json, raw full census archives and all result records.
