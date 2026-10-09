Defense blocked by a red clean baseline.
No production mutant was planted and no test was edited.
Evidence is retained for the owner to rerun after freeing temporary build space.

Starting commit: 157a43552015f41a79331949c2e82b6f8c7caaab. Audit metadata read: README.md, rows.json and production-plan.json from the full fetched audit ref. Current list has 6760 Test names. All eight requested family members still exist, now in node_table_oracle_floor_test.go rather than node_table_split_test.go.

Code under test: TypeScript lint port Linter.run, ancestry, walk, fixed and main.ts entry, compiled through Adamic. Current oracle: Go cohere findings, asserted nonempty, then native plain-versus-junk equality. The audit reported only the latter self comparison. Current nodeTableRunShard checks Go output at lines 339-345 before the junk comparison at lines 347-349. Commit ca87bd7a added this oracle floor. This changed test cannot inherit the audit's empty-output finding without replay.

Warm env.sh worked; setup skipped; npm ci completed before baseline. nproc=5. The whole-package baseline ran 90.265 binary seconds and timed out, but it was also red before the timeout: TestLegacyMutants failed its Go oracle build with no space left on device, and TestMutants failed because that build had failed earlier. The timeout does not erase those earlier failures. Therefore the rule never defend on a red baseline stops this defense.

A family coverage control had already been launched with -coverpkg=./internal/load,./internal/lower,./internal/native when the failed-test events were noticed. Its process tree was stopped; its log is incomplete and is not coverage or baseline evidence. No rest-of-package profile, coverage difference, mutation, native rebuild proof or uniqueness matrix was completed. Go coverage would measure Adamic compiler work, not execution of TypeScript port functions; additional native or source instrumentation would be needed for port coverage claims.

Semantic lead, untested: only node_table_oracle_floor_test.go passes --junk-rows, so this family uniquely exercises duplicate unattached rows. Possible aimed bound errors in Linter.run were considered but never planted. No honest attempt count or defense verdict is fabricated.

Name/assertions finding: the family explicitly compares native output before and after unattached rows and checks a nonempty independent oracle. Those assertions check the promised observable invariance. They do not assert that junk-row perturbation actually appended the intended rows; disabling that perturbation could still pass. This is an inspection finding, not a survivor demonstrated in this session.

Brief costs and ambiguities: audit evidence describes an older test body and filename; main has since strengthened and moved it. The package now lists 6760 Tests, so full execution exceeded 90 seconds. The baseline's oracle build exhausted the separate /tmp filesystem despite free space on /workspace. The requested Go coverage mechanism cannot directly instrument TypeScript. Approximately five minutes were spent fetching, inspecting, installing npm dependencies, running the failed baseline and stopping the coverage control. No tests were deleted, rewritten or weakened. Only evidence is committed; no main push or pull request.
