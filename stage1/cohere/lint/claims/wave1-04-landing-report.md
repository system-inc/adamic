Built: helper branch rebased, re-greened and pushed; rule branch landing is blocked by shared foundation conflicts, so no new helper is claimed.
Commits: helper landing b272eab3; rule source restored to 3aa2bc62 after aborting the conflicting rebase; this report commit follows.
Commands and outputs: setup PASS 17s, nproc 5; rebased helper comparison/mutants PASS 18.910s; input oracle PASS 11.104s; helper vet PASS; rule rebase stops on 29175443.
Mutants: all five helper mutants compile and are caught only by actual Go byte comparison on source Node, emitted JavaScript and sanitized native after rebase.
Not covered: rule branch oracle against current main, full repository gate, new helpers or rule integration gaps already documented in the owned rule reports.

# Landing cap

Fetched all origin heads. Neither owned branch was already an ancestor of current
origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965. The helper branch
codex/lint-helpers-from-codex/lint-wave1-04 rebased cleanly, and both its existing
Go differential oracles and all five semantic mutants passed again. Its bounded
gate is recorded in helpers/from_wave1_04/REPORT.md and evidence/landing-*.log on
that branch. The branch was pushed under its own name as b272eab3 with an exact
force-with-lease on its previously published 0618bb18 tip. No main or area branch
was pushed. cloud/setup.sh passed: Go, clang and Node ready at 0s; submodules 1s;
cache warm and done 17s; nproc 5, cpu.max 400000 100000.

# Rule branch blocker

Command: git checkout codex/lint-wave1-04, then git rebase origin/main.
It stops at the first inherited registration foundation commit:
2917544391494545fad4c86f84ce2229ccd8b164, Discover lint rules from independent
directories. The conflicted files are:

- stage1/cohere/lint/README.md
- stage1/cohere/lint/lint.ts
- stage1/cohere/lint/lint_test.go
- stage1/cohere/lint/testdata/oracle.go

These are shared files outside this unit's rule directories. This is a substantive
foundation integration conflict: main has additional monolithic listeners and
option adapters while registration replaces the monolithic dispatch with generated
per-directory listeners/adapters. Taking the registration side wholesale would
lose main's added listeners; taking main's side would leave our per-directory rules
unregistered. The shared tests also disagree about the portFiles slice/function
API, with profile_test.go on main still requiring the slice. Either choice needs
a shared registration/harness integration by its owner, rather than a blind conflict
resolution or a rule-local source fix.

Ahra's explicit scope instruction remains: "Keep your changes inside your own rule
directories. Don't edit the shared registration generator or the test harness."
Ahra also says, "If anything else blocks you, say exactly what it is and stop,
rather than editing shared files." The new landing-first cap requires every
previously pushed branch to be on main or rebased, green and repushed before any
new claim. The rule branch does not satisfy that cap. No additional rule or helper
was claimed, and no conflicted shared file was edited.

The complete failed rebase and combined conflict diff are preserved as
wave1-04-evidence/landing-rebase.log and landing-conflicts.log. git rebase --abort
restored the prior rule source tip 3aa2bc62003d7bd9ade4ee20f418bee5666ef286. Both
original branch tips have local backup refs. This report changes no rule source.
Earlier passing comparisons and integration limits remain historical evidence;
they are not described as a new green rule gate against current main.

The five repeated helper mutants were: accepting an empty configured filename,
omitting parent-directory cleaning, normalizing the literal cwd fallback,
discarding the recording snapshot, and taking that snapshot before the load.
Every mutant completes cleanly on all three execution paths before the Go byte
comparison catches it. These two helpers remove two prerequisites from each of
the six Tailwind canonical/class-order/variant-order/shorthand/conflicting/unknown
rules; they do not remove those rules' other dependencies or supply a native CSS
engine. Current owned work stops at the rule foundation landing blocker.
