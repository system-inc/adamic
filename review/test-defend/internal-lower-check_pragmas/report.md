CODE UNDER TEST: Lower and lowering.property for getter reread refusal; cycleFinder.slotsOf, cycleFinder.reaches and cycleFinder.unproven for static constructor-parent cycle detection. ORACLE: self-written diagnostic substring assertions in class_features_test.go. No oracle, harness or test was modified.

Current main: 7b9d4272c28f59530ab13daa5c49067e47933b06. Baseline passed (37.167 test-binary seconds). Every matrix ran all 276 currently listed top-level tests, including additions since audit. Default opt-in skips remain visible in logs. No bounded inference is needed for completed runs.

Coverage: NarrowedAccessor has 401 exclusive covered blocks versus StaticInterfaceCycle. StaticParentCycle has three: cycles.go:329, fresh.go:44 and fresh.go:61. Coverage-differences.json lists each exact block. Both rows use Lower; no loader mutation was made.

D01 changes only the narrowed-accessor diagnostic label. Its uniqueness proves the wording contract, not execution behavior. D02 ends the slot search at a safe earlier field; D03 ends the write search at a nonmatching write; D04 drops traversal of the static-parent ownership edge. These target the direct typeof Child path versus the construct-signature interface path, with no test names or source spelling in production edits.

Brief and cost findings: The audit's relevant M18 results were isolated after a package panic, so its subsumption was bounded. Tests have moved from lines 141/149 to 117/125. Disk free remains below the requested 15 GB after deleting the known prior-unit cache: /tmp has only 8.8 GB total, 3.0 GB free, and /workspace has 5.8 GB free. Shared tools and repository files were preserved. Warm env.sh worked; npm ci ran before baseline. No setup installation was needed.

Name/assertion finding for StaticParentCycle: it requires an error containing cycle, which matches the named refusal behavior, but does not check Refused type, code, location, or the precise parent ownership edge. A different cycle refusal can satisfy it. This is a limitation, not a name promising an untested performance threshold.

Location correction: fresh.go:44 is the exclusive unknown-write return, not the mismatching-write continue. D03 changes the shared continue at fresh.go:50. D02 changes continue at cycles.go:329. The selection used a shared-code semantic lead for D03. Matrix wall time totaled 182.706 seconds, with 5 processors. All four go vet validations passed; no panic or timeout occurred.
