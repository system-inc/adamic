Built: rebased twelve completed ports onto current main; no new rules or claims.
Commits: main f8013f0ba; rebased/published-ancestry tip 4fe121961; this report committed separately.
Commands and outputs: four owned oracles PASS (673.723s); bridge PASS (150.960s), checker PASS (1.998s).
Mutants: all twelve rule mutants, scope-export and released-handle guards caught again.
Not covered: three reserved React ports, new node-listener contracts, full repository gate and new benchmark medians.

## Landing refresh

Fetching all origin heads found main at
f8013f0baac41ddc340d76f83bddde38536a8f07. The linear rule history was rebased
onto it; the two later report commits were retained. The previous published
tip 9871aa7f0 was retained as an additional parent without changing the rebased
tree. Both current main and the previous published tip are ancestors of
4fe121961f92095c3ad36b7ac54b3ae47a5bb839, allowing a normal push.
This avoids replaying duplicate commits from prior preserved ancestry.
The rule and bridge trees are identical to the previous published version.
Only codex/typeaware-wave-22 is pushed; main and area branches are untouched.

## Verification

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -timeout=30m -v
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
```

Output went directly to the saved logs in validation-wave-22-landing-refresh.
Fourth batch passed in 193.18s; continuation in 204.05s; original in 143.63s;
third in 132.85s. The full focused invocation passed in 673.723s.

All twelve per-rule mutants again compiled, exited zero with empty stderr, and
were caught by byte comparison alone: numeric-prefix, has-own-fix-span,
spread-parens, promise-condition, spread-await-edit, lost-write-span,
nullish-suggestion, qualifier-fix, private-read, global-provenance,
global-declaration-span and timer-string. Scope-export was caught too.
Released handles panic with exit 70; the registry-retention mutant exits zero
and is caught by the required refusal. Normal and sanitizer runs match Go on
positive controls and both frozen corpora, including findings, repairs and
suggestions. Repository emitted 18485 identical bytes and compiler 5857;
both have zero findings, with positive controls checked separately.

The bridge matched 54982 bytes across 1600 positions in four files under
ASan, UBSan and LSan and reran its existing bridge mutants. Its package passed
in 150.960s; checker passed in 1.998s. No shared harness was edited.

Single native/Go wall-time observations from the original rule suite are
0.400286/0.177394s on repository and 2.832997/0.500128s on compiler. Native
remains slower in these observations; these are not new benchmark medians.
The existing setup was reused (118s total, nproc 5). No full repository gate
was run.

## Outstanding prerequisites

The three reserved rules remain react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components. Their
required native React HIR/SSA, capture/value propagation and dominator analyses
are still absent. The fetched shared-ssa branch afa0cb0b7 has no corresponding
stage1 native React graph implementation. JSX parsing is published separately
on codex/stage1-jsx-lint, as recorded in the prior report, but syntax support
alone does not unblock these graph-dependent ports.

Ahra instructed workers to stop at other blockers rather than edit shared
files. No shared parser, graph infrastructure, generator or harness was edited.
No additional rule was written or claimed. No new compliant rule.json kinds
or node-only listener implementation is claimed. The user supplied no shared
Diagnostic integration SHA, so no speculative rebase onto a worker's finding
model was performed. Earlier reports retain the exact JSX reproduction and
production Go tests; none proves coverage of these three unimplemented rules.
