# Internal/native deletion-set replay

Main: 7b9d4272c28f59530ab13daa5c49067e47933b06, detached while testing. Cohere submodule: 7945d102a6c18dd36adf9114a758ce646e8b2359.

The candidate list was recorded before the baseline. All 24 selected diffs applied without modification. The green baseline took 134.958 seconds: 190 top-level passes and 85 default skips. Replay wall time totaled 1452.857 seconds. Every run used ADAMIC_GATE_UNCACHED=1 and a fresh per-run ADAMIC_BUILD_CACHE_DIR. No tests were deleted or changed.

Nineteen mutants retain a clean non-witness catcher outside the set. Four completed package runs failed only TestOptionalMethodThunksMatchNode, a test that plants a missing-thunk mutant and verifies that Node catches it. Its production-mutant precondition failures do not count as catches. Consequently TestParserHasNoUnusedOptionalMethodThunks retains the last shown non-witness catch for those diffs.

R013/D2 from test-defend/internal-native-number applies but cannot compile on current main: dtoa.c:427:13 reports unused function bignum_less under -Werror. TestDecodeASCIIUnit42 and Unit41 reported that build error. They are recorded as build-failure rows, not behavioral catches. The number-format candidate therefore remains unresolved and is conservatively retained. The other five candidates satisfy the replay's deletable criterion. This is an evidence report, not an authorization to delete.

No Go package-aborting panics occurred. Native-program panic diagnostics are ordinary observed outputs, not Go test-binary panics. Caught runs stopped after the first completed non-witness top-level failure; passing or witness-only runs completed. Later rows in early-stopped runs are unknown. Default skips limit this evidence to enabled checks.

The inventory covers audit M*.diff and defender D*.diff files in each named branch's own review directory, including archived sessions. Defender B*/R* files fall outside the requested filename patterns. Repeated IDs are disambiguated by replay IDs R001 through R024 and source paths in mutant-list.json. Superseded temporary gather artifacts remain untracked and are not part of the published evidence.

Evidence: mutant-list.json records historical candidate and other-row failures plus branch provenance. selection-inventory.json records all 87 eligible-name diff files considered. matrix.json records current replay runs and catches. report.json records the deletion classification. failure-evidence.json provides commands and failing lines. logs/ contains the original Go JSON output, run records, and apply-check results. baseline-rows.json lists baseline passes and default skips. Source was restored after every diff.

Initial disk availability was 8.8 GB on both mounts. The previous unit's named scratch cache was removed; /tmp's total capacity is below 15 GB, so the requested free-space threshold is unattainable on this workspace. No disk failure occurred.
