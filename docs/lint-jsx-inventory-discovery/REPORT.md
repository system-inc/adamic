# JSX inventory discovery

The rule branch remains at `55a2730268fbefaf19385ff38f5d6f65291bef46`.
The independent fix commit is `d997c3a9257c9823b3b0a5f7871fe95a2e58e518`,
on `lint-fix/jsx-inventory-discovery`, created from
`origin/area/stage1-lint` at `1f9e223d`.
The rule was merged onto this branch at
`a03bf84a979beb83a6a1737db49b6e6b091db00c` for integration validation.
No shared test change was made on the rule branch.

## Original failures

At the rule commit above, both failures reached the shared hard-coded map
assertion in `stage1/cohere/lint/jsx_integration_test.go:61`:

- `TestJsxLintReleaseAndThroughput`, reported at
  `stage1/cohere/lint/jsx_integration_test.go:93`.
- `TestJsxLintTrees`, reported at
  `stage1/cohere/lint/jsx_integration_test.go:138`.

Both had exactly this assertion message:

```text
captured JSX cases by rule map[nexus/consistency-no-abbreviated-identifier:5 nexus/consistency-no-ambiguous-identifier:4 react/jsx-curly-brace-presence:159 react/jsx-no-comment-textnodes:40 react/no-find-dom-node:9 react/no-is-mounted:5], want map[nexus/consistency-no-abbreviated-identifier:5 nexus/consistency-no-ambiguous-identifier:4 react/jsx-no-comment-textnodes:40 react/no-find-dom-node:9 react/no-is-mounted:5]
```

The prior map encoded five rule names and their JSX-case counts. Discovery
correctly added 159 cases for the new rule, which the map rejected before
runtime or parser-tree comparisons ran.

## Change

`jsxSources` now gets descriptors from the same registry discovery used by the
rest of the harness, obtains captured cases from `upstream`, and classifies each
source with the unchanged Go JSX oracle. No public rule names or case counts are
pinned in a shared test. Sources from generic listeners are included whenever
they contain JSX; listener kinds are not used as the corpus filter.

The helper still rejects malformed rows and unregistered captured names. A
discovered listener subscribing to any JSX kind must have captured JSX coverage,
and a completely empty JSX corpus is refused. This guards disappearance without
requiring shared list updates when a rule directory is added. Focused regressions
exercise a new descriptor with two cases, generic-listener JSX, missing coverage,
unknown rules, malformed rows and an empty corpus.

## Validation

The two formerly failing tests and discovery regressions passed with the rule
merged (55.752 seconds). Discovery found 222 JSX sources, including all 159 new
rule cases; parser trees held 134,737 identical bytes against Go on Node and
sanitized native. JSX release output, emitted JavaScript and throughput counts
also matched Go.

The whole lint package passed on the merged integration commit: 120 pass, 0 fail,
1 skip (`TestCheckerBridgeRefusalPending`, the existing pending checker bridge
refusal at `stage1/cohere/lint/checker_pending_test.go:49`). Package exit status
was zero. Wall time was 892.715 seconds; nproc was 5; load before was
0.35/0.28/0.71 and after was 1.96/2.31/1.66.

Every external input was supplied: a clean TypeScript checkout at
`050880ce59e30b356b686bd3144efe24f875ebc8`, WASI SDK 27, throughput enabled,
and one fresh directory used for both profile variables. Exact paths, command,
counts and timing are in `whole-summary.json`. The full-run log is omitted
from the pushed evidence bundle to keep it small.
The profile and runtime checks passed, including all mutants.
