Built: rebased twelve completed rule ports onto updated main; numeric listeners and three React ports remain blocked.
Commits: main e8ba3d5d; rebased tip 202a1b46; published ancestry retained at d974b225; this report committed separately.
Commands and outputs: four wave-22 oracles PASS (682.845s); bridge PASS (147.768s), checker PASS (1.123s).
Mutants: all twelve per-rule mutants, scope-export and released-handle guards caught again on updated main.
Not covered: numeric listener declarations/dispatch, three pending React ports, full repository gate and new benchmark medians.

## Landing first

Only codex/typeaware-wave-22 has been pushed for this unit. Fetching all origin
heads found that main advanced from e011f8f6 to
e8ba3d5d81de4d3773c723914fccd4c76248b965. Rebase onto that main completed.
The previous history-preserving merge caused original commits to replay after
their rebased copies. The resulting conflicts were duplicate registrations,
claims and reports, not new implementation conflicts. Only identified duplicate
original commits were skipped. Logs preserve each skip. Comparing the rebased
rule and bridge trees with the previous published tip produced no differences.

Commit d974b225 retains the previous published tip 843e9505 as an additional
parent without changing the rebased tree, so the own branch can be pushed
normally without force-pushing. Main and the previous published tip are both
ancestors. No push to main or any area branch was made. No new claims were made.

## Re-green commands and evidence

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -timeout=30m -v
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
```

Both commands wrote output directly to logs. Results saved in
validation-wave-22-speed-landing:

| Check | Result | Seconds |
| --- | --- | ---: |
| Fourth preference batch | PASS | 197.90 |
| Continuation promises/spread/concurrency batch | PASS | 219.76 |
| Original three rules | PASS | 144.86 |
| Third global-binding batch | PASS | 120.31 |
| Entire focused package invocation | PASS | 682.845 |
| Bridge package | PASS | 147.768 |
| Checker package | PASS | 1.123 |

All four rule suites matched production Go findings, repairs and suggestions
byte for byte on positive controls and the frozen repository/compiler corpora,
in normal and sanitizer builds. The corpora emitted zero findings: repository
18485 bytes and compiler 5857 bytes. Positive controls prevent these checks from
being vacuous.

All twelve rule mutants again built and ran successfully, with exit zero and
empty stderr; the byte comparison alone caught numeric-prefix,
has-own-fix-span, spread-parens, promise-condition, spread-await-edit,
lost-write-span, nullish-suggestion, qualifier-fix, private-read,
global-provenance, global-declaration-span and timer-string. The scope-export
mutant was caught too. Released handles panic with exit 70; the retained-registry
mutant exits zero and is caught by the required refusal. Bridge outputs matched
54982 bytes across 1600 positions in four files under ASan, UBSan and LSan;
its existing bridge mutants were rerun.

The original suite measured native/Go repository wall time 0.384357/0.190968s
and compiler time 2.734073/0.401556s. These are single observations on the
unchanged rules, not a numeric-dispatch performance claim or new medians.
Native remains slower. The existing toolchain setup was reused (total 118s,
nproc 5); the full repository gate was not run.

## Numeric listener prerequisite

Observed API in stage1/typescript/parser/nodes.ts:

```ts
readonly kind: string;
constructor(kind: string, pos: number, end: number, children: number[])
```

Parser.make also accepts a string kind. Searches in the native parser and
native type-aware sources found no SyntaxKind enum, numeric kind field or
numeric kind constants. The inspected main, bridge-base and shared-harness
refs also expose the string field. Exact ref SHAs and parser blobs are saved
in listener-gap.json. This observation is limited to those checked refs.

The existing rule implementations still compare string kinds and scan/refetch
nodes. They do not meet the newly requested numeric listener contract. No
numeric declarations were added: there is no parser numeric contract to use,
and inventing local IDs would not establish compatibility with the coming
shared driver. Numeric declarations and node-only handlers remain unfinished.
No speed-contract mutant was run, because no speed-contract code was changed.

Adding numeric SyntaxKind storage/constants to the shared parser or changing
the shared driver is outside this unit's permitted files. Ahra instructed:
“If anything else blocks you, say exactly what it is and stop, rather than
editing shared files.” This is a parser prerequisite, beyond the anticipated
driver ignoring declarations. Work stopped at it after completing the landing
requirement. No shared parser, driver, generator or test harness was edited.

## Reserved React rules

react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components remain pending their JSX parser and native React
HIR/SSA prerequisites. See WAVE_22_FIFTH_REPORT.md for the runtime reproduction
and independent production Go tests. No new native coverage is claimed for them.

## Latest continuation check

Fetched every origin head again successfully. Main remains e8ba3d5d, already
an ancestor of published tip 4758f5c4. The rule/compiler sources were not
changed after the passing 682.845s oracle run and 147.768s bridge run above;
those checks were not repeated for this documentation-only update. No other
branch was pushed for this unit.

A prerequisite has progressed: origin/codex/stage1-jsx-lint at
a8a62d62ca49db7415e14c3887dd305022b17309 contains jsx.ts and its JSX_REPORT.md
records native JSX parser parity and batch-8 integration. This is published
worker evidence, not a new validation run by this unit. It is not on the
inspected main. The earlier parser panic remains evidence about this branch,
not a claim that no JSX implementation exists anywhere on origin.

The inspected JSX branch still has no native React HIR/SSA paths or matches
for ForFunctionWithoutManualMemoization, ControlDominators or
UnconditionalBlocks under stage1/cohere. The three pending rules require these
graph and capture/value analyses, as described in WAVE_22_FIFTH_REPORT.md.
Bringing JSX syntax alone into this branch would not unblock their ports.
Ahra's instruction to stop at other blockers without editing shared files
still applies. No shared files were edited and no new claims were made.

The latest instruction requires new rules to declare rule.json kinds and
consume the provided node without string relevance dispatch. No new rule was
written in this continuation, and no compliant new listener is claimed.
The message supplies no shared Diagnostic integration SHA, so no rebase onto
a speculative batch-8 worker tip was performed. Exact inspected tips and
ancestry are in validation-wave-22-speed-landing/latest-prerequisites.json.
