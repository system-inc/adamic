# View guard roads paired census

Task 3zex15c. Measurement only. Continuation in progress. The prior blocked report is preserved in commit 19d28a6cefb5ad4b2e7e0c569699028159e7b3db.

A scratch-only observer now calls the real guard once per recorded view, at each attempted unit end. Both overlays use the identical guard-hook.go.txt and instrument-guard.py. A fresh cycleFinder has l set to a disposable unit snapshot and an empty where map. No extra graph state is populated: guard helpers lazily initialize field/function caches and read the unit checker, type mapper, staticGlobals and representation metadata. GraphTypes, selected graph nodes, shapes, reachability and whole-program graph membership are neither computed nor claimed. findCycles and graphTypes never run.

Each singleton viewSources slice lets the real guard report that view even if an earlier view refused. Its real viewPairFailure path is queried only after a refusal to retain step source/target names and classify roads; the acceptance decision is never reimplemented. A side stream records GuardRefusal, GuardError and GuardUnit rows. Probe errors are explicit. A snapshot isolates lowering-owned state; checker caches remain shared and will be checked against original normal census output.

Small-control verification passed on both sides: one view, zero real refusals, one refuse-every-pair mutant refusal, zero probe errors. Entire normal streams are byte-identical to the original binaries with both the real and mutant observer. See raw/control-verification.json.

First-error parent/head seam runs and original-binary baselines are active on the same parent-adapted tree. GOMAXPROCS=2 for each; runs overlap on the four-CPU box. Full mode is conditional on new first-error guard refusals. Counts are pending.
