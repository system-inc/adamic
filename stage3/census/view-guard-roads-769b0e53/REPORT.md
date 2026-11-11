# View guard roads paired census

Task 3zex15c. Measurement only. Continuation in progress. The prior blocked report is preserved in commit 19d28a6cefb5ad4b2e7e0c569699028159e7b3db.

A scratch-only observer now calls the real guard once per recorded view, at each attempted unit end. Both overlays use the identical guard-hook.go.txt and instrument-guard.py. A fresh cycleFinder has l set to a disposable unit snapshot and an empty where map. No extra graph state is populated: guard helpers lazily initialize field/function caches and read the unit checker, type mapper, staticGlobals and representation metadata. GraphTypes, selected graph nodes, shapes, reachability and whole-program graph membership are neither computed nor claimed. findCycles and graphTypes never run.

Each singleton viewSources slice lets the real guard report that view even if an earlier view refused. Its real viewPairFailure path is queried only after a refusal to retain step source/target names and classify roads; the acceptance decision is never reimplemented. A side stream records GuardRefusal, GuardError and GuardUnit rows. Probe errors are explicit. A snapshot isolates lowering-owned state; checker caches remain shared and will be checked against original normal census output.

Small-control verification passed on both sides: one view, zero real refusals, one refuse-every-pair mutant refusal, zero probe errors. Entire normal streams are byte-identical to the original binaries with both the real and mutant observer. See raw/control-verification.json.

First-error parent/head seam runs and original-binary baselines are active on the same parent-adapted tree. GOMAXPROCS=2 for each; runs overlap on the four-CPU box. Full mode is conditional on new first-error guard refusals. Counts are pending.

The matched first-error prefix now has head-only guard refusals. The smallest witness in identity.ts.txt is a Narrow-to-Wide view with a field declared as the exact same alias Choice on both interfaces. Parent accepts; head refuses through `Choice as Choice`, then `Choice as A`, then `B as A`. This is a guard false refusal: writing Choice through that field introduces no value outside the source field Choice. The ordinary control file records remain byte-identical; the checker relation-report cache sizes differ, as expected from the additional guard queries. Full corpus noninterference verification remains pending.

Full-mode controls also passed on both compilers: one recorded view, one forced-refusal mutant record, zero probe errors, and unchanged serialized normal findings against each original binary. File-specific full checks will use a common original-parent normal baseline for both seam sides; the original latent call-path proof establishes that the original head differs only in unreachable guard code.

Original head baseline completed: child exit 0; 4058 timed non-generic unit attempts; 4869.733 seconds; peak RSS 5123252 KiB. The first-error guard pair and corpus noninterference checks remain in progress.

Original parent baseline completed: child exit 0; 4966.145 seconds; peak RSS 5101784 KiB. Both original baseline finding arrays and all non-telemetry census bytes agree exactly. The common original-parent full-mode baseline is therefore backed by both the source call-path proof and this complete first-error comparison.

Parent seam first-error census completed: child exit 0; 4058 observed units; 635 views; zero guard refusals and zero probe errors; 4888.397 seconds; peak RSS 4624656 KiB. All ordinary headers, file inventories, unit records and findings are byte-identical to its original no-seam baseline; only checker-cache telemetry may differ. parent-noninterference.json preserves the proof hashes.
