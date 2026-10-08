# React package validation

Base: `1f9e223d8ef6d0a914407d4b37d0e56568521f4d` (`origin/area/stage1-lint`).
Triage: `d7ab0bc4`; source ownership and duplicate decisions are in [README.md](README.md).
Go cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`.

[Helper package run](testdata/helpers-full.log): all three packages pass, zero
failures and zero skips. React takes 217.207 seconds. It captures 4,064 actual
asserted consumer cases from 31 rules, adds one syntax witness, projects 91,855
nodes, and tests 5,685 names. All 97,540 output rows agree byte-for-byte with Go
on source Node, emitted JavaScript and sanitized native. The namespaced-member
callback additionally tests constant true and false. Ten distinct semantic
mutants compile, execute and disagree with Go on both Node and sanitized native.

[Selected rule run](testdata/rule-selected.log): `react/state-in-constructor`
passes all 82 unique captured source/rule/options combinations and two owned
witnesses, byte-for-byte across Go, Node, emitted JavaScript and sanitized native
(39,473 output bytes). The selected run takes 101.889 seconds including its
`accept-plain-classes` mutant, caught on all three port runtimes. The reproducible
selector is [run-selected.sh](../../rules/react-state-in-constructor/testdata/run-selected.sh).

The witnesses preserve Go's surprising report on `#state`: its hashed private
identifier is treated as state even though it is not React state. Parenthesized
bases use the upstream rule's fallback after the strict shared helper declines.
The shared helper itself remains strict about outer parentheses. Hook calls
also retain Go's rejection of `(React.useState)()`, and hook/component names
retain Unicode-uppercase acceptance (`useΩ`, `Émile`) while rejecting `use2`.
The rule
preserves the constructor-ancestor walk through nested classes and functions.
Typed capture options (`Mode`) and public string options are both retained.

The clean TypeScript checkout is `050880ce59e30b356b686bd3144efe24f875ebc8` at
`/workspace/scratch/typescript-6.0.3`. WASI SDK 27 is at
`/workspace/adamic-tools/wasi-sdk`. The full lint run sets `WASI_SYSROOT`,
`ADAMIC_TEST_WASI=1`, `ADAMIC_TYPESCRIPT_SOURCE`, `ADAMIC_LINT_BENCH=1`, and a
single fresh directory for both `ADAMIC_LINT_PROFILE_DIR` and
`ADAMIC_LINT_PROFILE_SNAPSHOTS`. The registry is regenerated before the run.

[Full lint events](testdata/lint-full.jsonl) and [machine/count summary](testdata/full-summary.json):
112 passing test events, 2 failing test events, 1 skipped test; overall package
fails. Wall time is 1,828.336 seconds (package elapsed 1,825.912 seconds), nproc 5
with a 4-CPU cgroup. Load before: 4.195 / 4.253 / 2.201; after:
1.873 / 2.083 / 2.260. There are no missing-input skips.

First failure: `TestJsxLintReleaseAndThroughput` at
`stage1/cohere/lint/jsx_integration_test.go:93`. Its captured JSX rule-count map
contains this rule's 26 JSX cases, while the shared expected map omits this
new registration. `TestJsxLintTrees` fails on the same frozen map at line 138.
The rule itself passes all 82 upstream cases, including all 26 JSX cases. The
full upstream/inherited comparison, 888-file compiler/stage1 comparison,
throughput, 1/2/5-worker sharding, all registered mutants, profile artifacts,
compilation, snapshots, witnesses and registration checks pass. Changing the
shared JSX count assertion is outside the requested commit scope, so it is
reported rather than filtering any upstream cases out of the rule.

The sole skip is `TestCheckerBridgeRefusalPending` at
`stage1/cohere/lint/checker_pending_test.go:49`, whose explicit prerequisite is
the unlanded `TSGoError` bridge API (`codex/tsgo-errors-as-values`).

Only `react/require-optimization` and `react/state-in-constructor` lose all frozen
helper blockers from this package alone. The latter now has its rule port.
There is no stopped helper or language gap. The unrelated checker-bridge test
has an explicit pending skip for `codex/tsgo-errors-as-values`; this package does
not modify that test or the missing bridge API.

## JSX inventory fix rerun

Merged `origin/lint-fix/jsx-inventory-discovery` at `e7c196a9` in merge
`2986e318b2ceca14e9f4238436b62dc88898c3ea`. The dynamic inventory checks repair
the two frozen JSX count failures described above.

[Full lint rerun](testdata/jsx-merge-lint-full.jsonl) and
[summary](testdata/jsx-merge-summary.json): package PASS; 121 passing test
events, zero failures, one skip (`TestCheckerBridgeRefusalPending`, the
unlanded TSGoError bridge prerequisite). No missing-input skips.
Wall time 2,265.266 seconds; package elapsed 2,257.901 seconds; nproc 5.
Load before 0.361 / 0.544 / 1.403; after 1.691 / 2.409 / 3.479.
Registry generation passed. All input variables described above were set,
with one fresh directory `/tmp/react-merge-profile.h4hZqK` for both profiles.
The 4,120 captured cases, 890-file comparison, throughput, sharding,
registered mutants, profile snapshots and registration checks pass.
