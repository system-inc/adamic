# Before a lint worker pushes

From the repository root, after cloud setup and sourcing its tool environment:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript-6.0.3 \
  python3 cloud/lint-wave-check.py > /tmp/lint-wave-check.txt 2>&1
```

The script exits nonzero when a check fails. It never pushes. Commit the candidate
first: tracked edits, untracked files and modified submodules invalidate the run.
The pinned compiler checkout is the one required by the lint harness,
`050880ce59e30b356b686bd3144efe24f875ebc8`.

Each worker owns one claim file, with its Git branch name mirrored in the path.
For branch `codex/lint-example`, commit
`stage1/cohere/lint/claims/codex/lint-example.json`:

```json
{
  "version": 1,
  "branch": "codex/lint-example",
  "rules": ["example-rule", "@typescript-eslint/another-rule"]
}
```

Use exact public rule names. A claim must be nonempty, have no duplicate names,
and belong to the current branch. No worker appends to a central claim list.
Claims are reservations: keep them on origin until integration or coordination
retires them. The final-push checker requires completed implementations and tests;
it is not a command for publishing an unfinished reservation.

Existing wave Markdown reservations are supported without rewriting them. The
script discovers the file by its `Branch:` or `Owner:` header, or select it:

```sh
python3 cloud/lint-wave-check.py --claim stage1/cohere/lint/claims/wave1-12.md
```

Keep the assignment list before the first `##` report section. Numbered lists
and bullets, with optional backticks and position prefixes, are read as rule
names; entries explicitly marked `skip` or `skipped` are excluded. A claim's
own `<stem>-evidence/` and `<stem>-report.md` files are permitted as owned
attachments. They are excluded from origin claim parsing. Other workers' claims
and attachments remain outside the worker's territory. JSON claims are the
strict format for new reservations.

The checks are:

1. The committed claim exists and matches the worker's branch.
2. Every origin head is fetched, even with a main-only `remote.origin.fetch`.
   Deleted remote heads are pruned; submodule histories are not fetched. Every
   other head is searched for claims and directory descriptors with the same
   public name. Existing Markdown claims are read too; evidence JSON is excluded.
   Legacy batch selection JSON and exact backtick names in BATCH
   reports count as reservations, including selections not yet implemented.
   Legacy production .ts/.a modules and their volume-rule metadata are searched too.
   A quoted name in a legacy production module or batch report blocks the rule
   conservatively; inspect that path rather than assuming it is unported.
   Witnesses, negative fixtures and prose mentions are excluded. The worker's own
   origin head is allowed, as is an identical inherited copy of its reservation.
3. Every claimed rule has its own `rules/<slug>/rule.json`, omits `order`, and is
   new relative to the registration baseline. Changes under the lint tree may
   touch only the claimed rule directories, this worker's claim and owned evidence. Changes
   to the registry command, shared imports/dispatch, Go oracle, copy/corpus lists,
   shared test harness or another worker's rule fail. Renaming a shared file
   counts as changing it. Generated registration must remain ignored. The actual
   registry generator validates the descriptors before tests.
4. `TestRulesAgree`, `TestOwnedWitnesses`, `TestCompilerAndStage1Agree`,
   `TestMutants` and `TestRegistrationMutant` run with `-count=1 -timeout=30m`.
   All must run and pass; a skipped compiler corpus is a failure. Each parity
   test must report a nonempty Go/Node/native comparison, and the upstream corpus
   must be nonempty. Each claimed rule's owned mutant must have its own passing
   subtest and report that its successfully executed output differed on both
   Node and sanitized native. Use distinct mutant names made of ASCII words,
   spaces, underscores or hyphens. A green command with absent tests or an
   uncaught mutant is rejected.

Main initially still uses shared lint registration. Until the accepted
`codex/lint-registration` migration is integrated, workers must include its entire
published tip. The checker recognizes that exact origin prerequisite as the
baseline. After integration, it compares against the common ancestor with
origin/main. There is no arbitrary baseline override that could hide shared edits.
This checker branch does not import that other unit's implementation.

Successful runs save Go JSON output, separate stderr, generator output and an
atomic receipt inside the checkout's Git directory (`git rev-parse --git-path
lint-wave-check`). These local files never create shared source lines. The
receipt records HEAD, the claim, mutant names, command, origin snapshot and the
test-log SHA-256. HEAD and cleanliness are checked again after execution. All
origin heads are fetched and collision-checked again before success.
`GOFLAGS -overlay` is rejected: a passing unapplied compatibility proposal does
not establish that the committed candidate builds and passes. Commit the needed
foundation changes through their owners before checking readiness.

Immediately before pushing the same tested commit, verify the receipt without
rerunning the expensive comparisons:

```sh
python3 cloud/lint-wave-check.py --verify > /tmp/lint-wave-verify.txt 2>&1
git push origin HEAD
```

Verification fetches origin again, repeats ownership/directory checks, checks the
current commit and log hash, and reads the test events again. A new commit needs
a new run, even if only its documentation changed. Do not treat a receipt as an
adversarial signature or a toolchain attestation. It proves this cooperating
worker ran the expected tests on the recorded commit, rather than trusting a
handwritten PASS line. Git fetch/check and push are separate operations; two
simultaneous claims can still race. Coordinators must resolve reservations before
releasing overlapping work, and integration still owns the complete uncached gate.

Verification of the checker itself:

```sh
python3 -B cloud/lint-wave-check-test.py > /tmp/lint-wave-tests.txt 2>&1
```

These CLI tests use disposable Git origins and synthetic Go events to attack each
check. Synthetic events are parser tests, not semantic parity evidence. The
[unit report](reports/lint-wave-check/REPORT.md) separately records a complete
real Go/Node/sanitized-native worker run and each failing control.
