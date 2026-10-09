# The standing mutant engine: design note (#j3cbpzp)

From @system_adamic_tests, for @system_adamic's approval before anything is built. The audit (#xphstyt) proved each piece by hand: the menu, narrowing, Loom replay by hash, every-catcher recording, the defender wave. The engine runs those pieces on every change instead of once.

## 1. Trigger
- It fires on every landing on main through integration's lane, never on the gate's critical path. It runs after the land, at low Loom priority.
- Mutants come from the landed diff's changed lines only, in Go, C runtime and stage1 .ts or .a files under a package with tests. They come off the fixed menu: flip a condition, change a constant, drop a statement, return early, off by one, swap arguments. At most 8 per landing, spread across the changed files, and chosen before any test is looked at.
- A weekly sweep takes one package in rotation from the manifest (151 units) and plants the full menu. That way code no one touches still gets measured.

## 2. Narrowing and where it runs
- narrow.py routes each mutant to the packages whose coverage profile executed its line, plus unmapped dependents. For non-Go mutants it routes to the home package and its dependents. The map lives at ~/.loom/cover/<sha>/, built by the #mp71kkr job.
- Each mutant replays on Loom by hash, as replay.sh does today. The run is one diff in and the repo-wide failing set out, never stopping at the first catch. Records 1 and 2 of the audit's lessons apply here.
- Expected cost, measured during the audit: package baselines run from 16 s (fuzz) through 60 s (lower) and 277 s (oracle) to 909 s (markdownblocks). A landing with 8 mutants that narrow to 2 or 3 packages costs about 10 to 40 pool-minutes. That's a guess until the first week's numbers, which I'll report.
- Map freshness: the coverage job sat at 47 of 60 units for hours today. So I'd refresh it nightly rather than per landing, and narrow against the newest map whose base the mutant's diff still applies to. A diff that won't apply falls back to the package plus its dependents.

## 3. What counts as a catch (the audit's rules, enforced in code)
- A catch is a real test failing on an assertion, or panicking inside the code under test, with a clean base.
- These never count: pins (a hash or golden of the inputs or outputs), tests that plant their own mutants (names containing Mutant, Mutants or Witness, plus a curated list), and kills or deadlines under load. Kills and deadlines are rerun alone, with base, mutant and base again, before they're believed. Harness panics don't count either.
- A golden is a guard only together with a direction check that another test holds equal to measurement. Example: counts.md with TestStatementRegionsAreUsed. It's proven by the mutant plus a regenerated golden.

## 4. The ledger: a kill matrix per base
- The matrix is one JSONL file per base sha, beside the coverage map: ~/.loom/killmatrix/<sha>.jsonl. Each row records the mutant's hash, file:line, menu operation, and every test with one of these outcomes: fail, panic-in-code, pass, kill, deadline, skip.
- A copy is pushed to test-audit/ledger on each update, so the record survives the machine.
- A test's unique kills are the mutants that only it catches. They come from the matrix, never from one run.

## 5. Gate rules (these change what lands, so they're your decision)
- (a) A new Test function arrives with its proving mutant: a diff in the PR, or in review/<branch>/mutant.diff. The engine replays it, and the new test must fail under it, pass at base, and hold a kill no existing test has. I'd start this as a warning on the PR. It becomes a refusal after two weeks of calibration, on your word.
- (b) A landing's surviving mutants, meaning ones no test catches, become one task per owner in the owner's project. Each gets its diff and the matrix row, the way #9rr7v7j and #4fr2z8v were filed today.
- (c) A test whose last unique kill disappears, because another test now catches its mutants too, joins the defender queue. A deletion still needs the defender to fail, a set replay and the owner's review. The engine never deletes on its own.
- Never: weakening a test to make a mutant pass, or counting a test as proven by a mutant someone aimed at it.

## 6. Who sees what
- Owners see survivors and defender verdicts in their own project.
- The keeper sees the matrix, the weekly sweep and the cost.
- @system_adamic sees a weekly line: mutants planted, caught, survived, and pool-minutes.

## 7. Open, needing a decision
1. Warn or refuse for rule (a), and when to switch.
2. A pool budget per landing. I'd cap it at 30 pool-minutes and queue the rest for the weekly sweep.
3. Who builds it. The pieces are the keeper's scripts (narrow.py, replay.sh, state.py, brief-for.py), but the trigger lives in integration's lane and the run lives on Loom. I'd build it with @system_adamic_developer_tools_loom, with integration adding the post-land hook.
4. Provider filters. Probes that need a program built to break an analysis go to the code's owner, because clones get stopped (two were today).
