Built: first complete original pair, SymbolTracker.moduleResolverHost, with eight Node controls and pinned runtime refusals.
Commits: this batch follows 59eb65c8; its commit is recorded by the branch tip.
Commands and outputs: checked-view oracle PASS 108.015s; lower/native/javascript PASS 2.553/4.982/0.926s; IR PASS 14.737s; vet PASS.
Mutants: skip, wrong shape, nested omission, required-field absence, outer check omission, and forbidden absence each caught by three backend pins.
Not covered: 16 object candidates and 65 reads remain; recursive unions and array/callback compounds are next; exact allocation reachability remains unmeasured.

Full-family estimate remains October 11, 2026, 23:00 UTC. Candidate inventory is 17 object pairs/66 reads plus 28 delegated primitive/array pairs/703 reads. One original pair/one read is now certified. The private SymbolTrackerImpl pair is not counted through the public tracker witness.

The fixtures import complete declarations emitted from official Microsoft/TypeScript at 050880ce. All 78 emitted declaration hashes, complete inherited field sets, and original UTF-16 read-site positions are checked. The checkout and emitted declarations stay outside the repository. This certifies the original field contract through representative reads; it does not claim the enclosing compiler function or a checker-clean whole compiler build. Empty-map production satisfies the present collection value; branded map operations and callable bodies retain separate lazy read/call obligations.

Run setup against a pristine official upstream checkout with stage3/adapt/00-setup/adapt.cjs, then lane4b/original/prepare.cjs, then this directory's prepare.cjs. Set ADAMIC_INTERSECTION_ORIGINAL_DECLS to the emitted declaration directory and run go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginalPairs$' -count=1. run-mutants.py accepts that directory and a log directory. Test output was captured under /tmp/intersections-original-*.log; six mutant logs are under /tmp/intersections-original-mutants.

Minimal shared hooks listed in docs/checked-views-plan.md: recognize physical callable fields in the JavaScript validator and preserve an emitted intersection read check when a wider carrier contributes an unrelated unsupported field-name fallback. Callable signature/body proof remains at its own read/call.
