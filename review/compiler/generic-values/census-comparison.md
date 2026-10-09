# Step 16 generic-value remeasurement

Same source bytes, checker diagnostics, root identities, statuses and body spans. Full isolated lowering with production Load/Lower disabled and no-output guard enabled. This is a blocker observation, not successful compilation or a claim of retired hidden bytes.

| Root | Historical value sites | Base generic blockers | Head generic blockers | Other head lowering blockers |
| --- | --- | ---: | ---: | ---: |
| binder.ts:1373:5 | core.ts:220:112 | 1 | 0 | 2 |
| binder.ts:635:5 | core.ts:220:112 | 1 | 0 | 2 |
| checker.ts:2698:5 | core.ts:220:112 | 1 | 0 | 33 |
| checker.ts:2791:9 | core.ts:220:112 | 1 | 0 | 0 |
| checker.ts:5229:5 | core.ts:220:112 | 1 | 0 | 16 |
| checker.ts:5249:9 | core.ts:220:112 | 1 | 0 | 12 |
| checker.ts:5737:5 | core.ts:220:112 | 1 | 0 | 21 |
| checker.ts:5762:9 | core.ts:220:112 | 1 | 0 | 0 |
| checker.ts:6101:9 | core.ts:220:112 | 1 | 0 | 0 |
| checker.ts:6769:9 | core.ts:220:112 | 1 | 0 | 152 |
| checker.ts:7718:9 | core.ts:220:112 | 1 | 0 | 0 |
| checker.ts:9845:13 | debug.ts:251:45 | 1 | 1 | 7 |
| core.ts:1370:1 | core.ts:1370:140 | 1 | 1 | 1 |
| core.ts:2377:1 | core.ts:2378:40 | 1 | 0 | 0 |
| core.ts:2442:1 | core.ts:2442:107 | 1 | 0 | 0 |
| core.ts:814:1 | core.ts:814:162 | 1 | 1 | 1 |
| factory/parenthesizerRules.ts:675:1 | factory/parenthesizerRules.ts:676:54 | 1 | 1 | 3 |
| factory/utilities.ts:1370:5 | factory/utilities.ts:1372:18 | 1 | 1 | 1 |
| factory/utilities.ts:1392:5 | factory/utilities.ts:1394:34 | 1 | 1 | 3 |
| path.ts:1039:1 | path.ts:1049:126 | 1 | 1 | 12 |
| path.ts:1045:1 | path.ts:1049:126 | 1 | 1 | 12 |
| path.ts:1047:1 | path.ts:1049:126 | 1 | 1 | 12 |
| program.ts:1075:1 | program.ts:1076:14 | 1 | 0 | 4 |
| sourcemap.ts:820:1 | sourcemap.ts:821:24 | 1 | 0 | 0 |
| transformers/classFields.ts:1956:5 | transformers/classFields.ts:2071:33 | 2 | 1 | 27 |
| utilities.ts:10629:1 | utilities.ts:10629:92 | 1 | 1 | 1 |

Summary: `{"base_generic_root_count": 26, "base_generic_site_count": 14, "base_lowering_panics": 29, "base_scan_findings": 10, "checker_diagnostics": 324, "head_generic_root_count": 11, "head_generic_site_count": 8, "head_lowering_panics": 29, "head_roots_without_observed_lowering_blockers": 7, "head_scan_findings": 10, "original_distinct_sites": 14, "roots": 26, "source_files": 79}`

The JSON companion retains every selected-root finding, module-scan findings, and recovered panic counts. Both runs have ten unchanged computed-property module-scan panics and 29 unchanged isolated-statement panics. Module scan failures and other lowering boundaries can hide children. A root without an observed lowering blocker is not certified accepted.

Audit mutants: changed source hash rejected: source bytes differ from historical ledger; missing root rejected: root coverage changed.
