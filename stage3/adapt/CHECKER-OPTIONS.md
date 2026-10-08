# Checker measurement option alignment

The production loader and checker ledger 3f0926c0 use erasableSyntaxOnly false and omit noImplicitReturns and noFallthroughCasesInSwitch. All current measurement profiles under stage3/adapt and stage3/census now follow those three policy settings. Strictness and each probe's existing project inputs remain intact. Named counterfactual experiments and historical evidence are retained.

The validator now rejects all three stale settings. The historical census policy overlay also accepts the already-false production setting and copies the loader byte for byte. The read/write renderer states the effective setting.

## Observed diagnostic counts

The identical completed apply tree, before npm installs upstream dependencies. Stock TypeScript 6.0.3 comes from the stage3 API cache. Existing project-specific inputs and options are retained; these counts are not production Adamic acceptance counts.

These are paired measurements using each script's actual first createProgram root and option expressions on the same completed adapted tree. They isolate the option change. They do not claim that every downstream inventory or coverage workload was rerun. Adaptation 75 audit.cjs was also executed in full and reproduced its 7699 to 7519 counts.

| Script | Before | After |
|---|---:|---:|
| stage3/adapt/10-type-imports/adapt.cjs | 751 | 238 |
| stage3/adapt/20-optional-declarations/adapt.cjs | 751 | 238 |
| stage3/adapt/47-host-errors/adapt.cjs | 752 | 239 |
| stage3/adapt/70-readonly-views/adapt.cjs | 8192 | 7240 |
| stage3/adapt/70-readonly-views/field-audit.cjs | 8192 | 7240 |
| stage3/adapt/70-readonly-views/survey.cjs | 8192 | 7240 |
| stage3/adapt/70-readonly-views/writers.cjs | 8192 | 7240 |
| stage3/adapt/75-optional-widening/adapt.cjs | 417 | 237 |
| stage3/adapt/75-optional-widening/audit.cjs | 7699 | 7519 |
| stage3/adapt/75-optional-widening/read-write.cjs | 7699 | 7519 |
| stage3/adapt/75-optional-widening/rebucket.cjs | 7699 | 7519 |
| stage3/adapt/75-optional-widening/coverage/run.py | 751 | 238 |
| 70-readonly-views/consumer.cjs | 0 / 2 | 0 / 2 |
| 70-readonly-views/narrow-range.cjs diagnostic portion | 0 | 0 |
| 75-optional-widening/coverage/probe.py, all five variants and both checkers | 0 | 0 |

The complete apply rerun measured adaptation 20 at 2728, 2071 and 2063 diagnostics before, versus 2213, 1556 and 1548 after. Both edited exactly 373 declarations.

## Output identity

Both fresh apply runs passed. All 81,504 files, including generated inputs and patch-set.md, have identical bytes; only Git metadata is excluded. The sorted path-to-SHA-256 map has aggregate SHA-256 `abb32a2145ee2cc4a75978f8509cdcbdaf7c7fdafa28f769b26fb4e009bf6435`. Full file manifests and run logs are preserved in /tmp/stage3-batch2-before-hashes.json.gz and /tmp/stage3-batch2-after-hashes.json.gz.

The regenerated checked-in patch table also records the merged Error sites: adaptation 40 is 17 files, 106 added and 106 removed lines; total is 78 files, 5128 added and 5102 removed. This table correction comes from the merged adaptations, not the option change.

## Mutants and limits

- actual core.ts one-byte addition caught by file SHA-256 comparison, then restored.
- erasableSyntaxOnly true rejected by check-strict.py.
- noImplicitReturns true rejected by check-strict.py.
- noFallthroughCasesInSwitch true rejected by check-strict.py.
- introduced TS9999 diagnostic rejected by check-strict.py.
- renamed loader option rejected by census overlay singleton guard.
- dropped enum owner prefix rejected against all 14 recorded field reasons.

The meter owner row uses the exact observed prefix `a checked numeric enum object view with an incompatible structured or optional payload field`; all 14 suffix variants in run 20261007T230852Z.jzKzWs map to 01a1183d-8e09, and reading Debug maps to 01a113e7-bfed.

This proof does not rerun native scanner execution, exhaustive read/write inventory propagation, full coverage instrumentation, macOS, or the full Go repository test suite. Requested Linux gate results are reported separately. See checker-options-evidence.json for exact per-code counts and measurement scope.
