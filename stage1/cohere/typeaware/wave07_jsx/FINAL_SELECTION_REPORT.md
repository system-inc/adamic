Built: refreshed twelve existing ports on integrated JSX support; no new claims.
Commits: tested source 794926f08 and green evidence 209768ba2; final audit follows on the same branch.
Commands and outputs: all 17 landing steps pass; fresh all-head audit examines 609 refs and 33 claim blobs, with zero eligible rules.
Mutants: every rule, named dispatch, legacy listener, graph/fact/data/ownership and shared-model witness is recorded in AREA_LANDING_REPORT.md.
Not covered: parked IR/SSA/capture-dependent hooks, four production bridge registrations, shared checker context, own emitted JavaScript and the full repository gate.

The final ranking audit excludes the original production ports, native filenames
and rule.json descriptors on origin/main and origin/codex/tsgo-c-library, and
rules named in claims on every origin branch. It inspects all 197 ranked names.
No unported, unclaimed name remains. No additional wave-07 claim was made.

The green worker branch is rebased onto tested integration commit 7481e032,
which contains current main 39638d9e. area/stage1-lint advanced to d65a8f93 during
validation; no result is claimed against that later integration revision. Main
remains an ancestor. Only codex/typeaware-wave-07 is pushed. All code changes
were tested before the first push; this final commit records only audit evidence
and documentation.

[AREA_LANDING_REPORT.md](AREA_LANDING_REPORT.md) gives exact outputs, every
mutant detector, remaining blockers, setup lines and native/Go timing tables.
The JSX medians are 0.324–0.329s native versus 0.135–0.137s Go on repository
sources, and 2.059–2.245s versus 0.319–0.326s on compiler sources. Source and
evidence integrity were checked; 2,097 command streams are archived beside the
fresh final audit and its independent byte hashes.
