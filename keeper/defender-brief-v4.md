# Defend these tests: does each one catch something no other test catches?

The test audit of github.com/system-inc/adamic marked the rows below in __PACKAGE__ as subsumed (every mutant it caught was also caught by another row) or untrue (no mutant made it fail). Before anyone may delete or rewrite one, a defender has to try hard to prove it worth keeping. You are that defender. Your bias is to keep: a row you can defend stays.

The audit's evidence for this unit is on branch test-audit/__SLUG__, under review/test-audit/__SLUG__/ (its mutants, matrix, and report); fetch it with a full refspec (`git fetch origin test-audit/__SLUG__:refs/remotes/origin/test-audit/__SLUG__`; a bare branch name only sets FETCH_HEAD), then read it first. Its report may be spread over several files (rows.json, report.md, code-and-oracle notes): read them all. The package may have gained tests since the audit: include them in your matrix.

## Your rows
__ROWS__

## For each row
1. Name the CODE UNDER TEST and the ORACLE, as the audit did. Never mutate the oracle, the harness or the test.
2. Find what this row reaches that its subsumer (or, for an untrue row, the rest of the package) doesn't: a per-test coverage profile of the row and of its subsumer (`go test -run '^Name$' -coverprofile`, with `-coverpkg` on the code under test), then the lines only this row covers. Exclusive lines are a lead, not the whole answer: two rows can share every line and still differ in what they feed it (an input history, an edit sequence, a boundary value). A defense on shared lines counts if it is semantic and you say what differs. For an untrue row, look for the behavior it claims to check (its name, its assertions) that the audit's mutants never touched.
3. Write up to three mutants AIMED at that difference. Aiming is allowed here, unlike the audit: your question is whether any break exists that only this row sees. Use the same menu (flip a condition, change a constant, drop a statement, return early, off by one, swap arguments); an implicit bound counts as a bound (writing one byte fewer is an off by one).
4. For each mutant, run the whole package (or, for a big package, the rows that reach the mutated function, listed) under `timeout 120 go test -json -count=1 -timeout 90s`, with its own `ADAMIC_BUILD_CACHE_DIR`. Record which rows fail.
Twins are not each other's cover. When two rows run one program through two executors (TestXNode runs the port under Node, TestXNative runs the natively compiled port, both against the same oracle), a port mutant fails both by construction, because both run the port. The Node row guards the port; the Native row also guards Adamic's compiler, which no port mutant touches. So never report either twin "not defended" against the other on port mutants alone: put `twin` in `defense`, name the sibling, and spend your attempts on the other rows. Only with budget left, try a compiler mutant that miscompiles this program (its own build cache) to defend the Native half.

5. Verdict:
   - defended: a mutant that this row fails on and no other row in the package does. Give the mutant's standalone diff, the failing line, and the list of rows that passed.
   - not defended: three honest attempts, each caught by another row too or by nobody. Say what each attempted and what it showed.
   - cannot-judge: why.
   An untrue row that you make fail with a real production mutant becomes "defended" if nothing else catches it, or "subsumed" (name the catcher) if something does.

## Rules
- Start clean from origin/main: `git fetch origin && git checkout -b test-defend/__SLUG__ origin/main`. Read CLAUDE.md, the audit's report, the rows' tests whole, then the code.
- Toolchain: skip setup if /workspace/adamic-tools/env.sh works; run `npm ci` in stage3/api before the baseline; never defend on a red baseline.
- Never delete, rewrite or weaken a test. Never push to main or open a pull request. Push your evidence (diffs, logs, coverage diffs) on test-defend/__SLUG__ under review/test-defend/__SLUG__/.
- Every diff applies to origin/main and compiles with the tool that builds it.
- No em-dashes in what you write.

## Final message
A three-line summary, then a JSON array, one object per row:
{"test","package","prior_verdict","subsumed_by","defense":"defended" | "not defended" | "cannot-judge","unique_mutant":"<id and file:line, when defended>","attempts":[{"mutant","file_line","change","rows_failed"}],"evidence":"<command and failing line>"}
Then every place this brief was unclear or cost you time. Also say, for any row you could not defend, whether its name promises something its assertions don't check (a row named for performance that asserts no threshold): that is a finding for its owner either way.

## Budget
About 20 minutes for the unit; 30 for a stage1 port. The test binary's budget is 90 s; the outer `timeout 120` only bounds compiling. A run past 90 s is cooked: stop it, narrow it, say so. A whole-package matrix near a minute means at most seven mutants: spend them on the rows with the most exclusive behavior first.
