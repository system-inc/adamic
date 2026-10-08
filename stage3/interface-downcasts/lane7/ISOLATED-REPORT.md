Built: isolated original helper certificates for 9454.argumentExpression and 9657.class.expression, with complete receiver/member provenance and own counts.
Commits: follows pushed d5c2ecce; this report accompanies the isolated member batch.
Commands and outputs: original declaration preparation PASS fourteen rows; four original controls and own counts update PASS 13.952s; final verification PASS 13.096s.
Mutants: 9454 and 9657 member contract omissions expected FAIL 3.630s and 3.000s; each wrongly prints true and exits zero in sanitized native, release native and JavaScript, independently caught.
Not covered: three remaining lane 7 pairs / six reads, including private 68704 excluded as instructed; no broad intersection or array admission change.

Working date October 11, 2026. Lane 7 now certifies fourteen original pairs / 60 reads of seventeen / 66. Assigned lane 4 Identifier 9477 is separately certified (one pair / one read) in d5c2ecce. Lane 4b seven / 45 overlap these certificates and are not added again.

Each generated .a imports the full pinned original compiler/types.d.ts. An interface CheckedReceiver inherits the complete original intersection (BindableStaticElementAccessExpression, or JSDocAugmentsTag['class']). Existing admission suffices. The oracle checks every original receiver field against the emitted contract, and the exact selected descendant field set and bounded intersection flag against the original manifest. There is no earlier read of Binary.left, JSDoc.class, or either helper member. All inherited Node and Identifier field sets remain checked. The wrong element argument is Identifier instead of StringLiteral; the wrong heritage expression is This instead of Identifier. Node prints true in each control. Checked modes stop with exit 70 naming node.argumentExpression or node.expression. Good counterparts match Node and pass leak checks.

Each mutant clears only that helper Property.ViewContract; it retains presence, kind, readiness and subsequent end read. All three modes now print true and exit zero. This establishes the helper's own deferred member selection is necessary, independently of an earlier read. No compiler changes made in this batch.

Own counted rows cover all four generated controls. Good allocations/frees 9/9, peak 9; wrong allocations/frees 9/0 at panic, peak 9. All have zero regions and graph merges. Rows are verified by the scoped original oracle; the user's instruction permits these own rows without a global counts refresh.

Commands, test output directly redirected:
- node stage3/interface-downcasts/lane7/original/prepare.cjs /tmp/lane7-upstream /tmp/lane7-declarations > /tmp/lane7-isolated-prepare.log: PASS fourteen original rows and exact source spans.
- ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane7-declarations go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginalIsolatedMembers$' -count=1 -v -args -update-counts > /tmp/lane7-isolated-counts.log: PASS 13.952s.
- Same test with ADAMIC_INTERSECTION_ISOLATED_MUTANT=9454 and -run '^TestCheckedViewIntersectionOriginalIsolatedMembers$/^9454$/^wrong$' > /tmp/lane7-isolated-9454-mutant.log: expected FAIL 3.630s, all three pins catch successful wrong output.
- Same test with ADAMIC_INTERSECTION_ISOLATED_MUTANT=9657 and -run '^TestCheckedViewIntersectionOriginalIsolatedMembers$/^9657$/^wrong$' > /tmp/lane7-isolated-9657-mutant.log: expected FAIL 3.000s, all three pins catch successful wrong output.

Final scoped counts verification: same original isolated test without update-counts > /tmp/lane7-isolated-final.log, PASS 13.096s. git diff --check clean.
