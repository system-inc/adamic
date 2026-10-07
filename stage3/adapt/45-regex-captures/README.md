# 45: Proven regex captures and split density

`adapt.cjs` resolves all 29 upstream-config regex diagnostics with 19 AST-planned
edits in eight compiler files. It uses stock typescript@6.0.3, reparses current
source, preserves line endings, and refuses unreviewed shapes before writing.
The regex participation analyzer is in proof.cjs. It checks regex syntax,
counts captures, unions guarantees through concatenation, intersects them
across alternatives, and clears guarantees under zero-minimum quantifiers.
Lookarounds, named groups, backreferences, unreviewed escapes and flags decline.
Only the reviewed ordinary regex forms and g/i/m flags are accepted.

Mandatory reads receive `!`. Two required major bindings move to `match[1]!`
without changing optional-capture defaults. Two split callbacks assert their
string element. Four capture-free split results receive a proven `as string[]`
density assertion; those casts erase and do not allocate a replacement array.
The range operator parameter gains `| undefined`, reflecting its existing
`case undefined` handling, rather than asserting an optional group.

The user explicitly authorized `(matchResult[2] || matchResult[3])!` as a
checked boundary for the reported latent bug. This layer preserves `||` and
its Node behavior. Kirk subsequently directed the fork bug fix to a separate
layer: adaptation 46 changes that operator to `??`. Adaptation 45 recognizes
that later checked expression when reapplied and leaves it untouched. Neither
individual alternative capture receives a mandatory assertion.

Validation uses an isolated checkout of origin/codex/stage3-type-imports at
`a3ef0dc`; adaptation 20 is excluded as requested. The branch itself keeps
only this unit's files and the separately requested fix and upstream ledger.

Source: TypeScript v6.0.3, commit
050880ce59e30b356b686bd3144efe24f875ebc8. Census: origin/codex/tsc-census,
profile upstream-config. All positions below refer to that unadapted source.
Counts: TS2345 15, TS18048 6, TS2322 4, TS2532 4; 29 total.

## Proof rule

A successful match is not proof that every capture participated. A group must
participate on every successful path through the expression. A quantifier with
minimum zero on the group or an enclosing group, or an alternative which
bypasses it, defeats that proof. Alternation *inside* a mandatory capture does
not defeat it. An empty captured string is still a present string. The expression-specific proofs below justify the analyzer decisions.
The analyzer is conservative; it is not a general ECMAScript regex optimizer.

Split is a separate contract: regex split can insert undefined only through
nonparticipating captures in its separator. A separator with no capturing
groups inserts only substring strings. This proves element values, not the
presence of an arbitrary index in the resulting array.

## All 29 diagnostics

M = mandatory capture; D = split element density; H = absence already handled;
B = real runtime bug; R = required positional argument checked at the use.
The table keeps original census locations. Shared source edits resolve multiple
diagnostics, so there are fewer edits than diagnostic rows.

| File | Line:column | Code | Decision and proof |
| --- | --- | --- | --- |
| checker.ts | 8062:153 | TS18048 | D: `/\r\n|\n|\r/` has no captures; callback receives strings. |
| debug.ts | 378:28 | TS2322 | M: `/^function\s+([\w$]+)\s*\(/` requires group 1 on its only successful path. |
| emitter.ts | 4111:27 | TS2345 | D: `/\r\n?|\n/` has no captures; iteration receives strings. |
| emitter.ts | 4977:46 | TS2345 | D: same separator; array passed to guessIndentation contains only strings. |
| emitter.ts | 4979:40 | TS18048 | D: same separator; lineText is a string. |
| emitter.ts | 4980:17 | TS18048 | D: lineText and its sliced result are strings. |
| emitter.ts | 4982:23 | TS2345 | D: same string result passed to write. |
| parser.ts | 10717:22 | TS2532 | M: `/^\/\/\/\s*<(\S+)\s.*?\/>/m` requires group 1, including at least one nonspace. |
| parser.ts | 10733:74 | TS2532 | M: named-argument group 1 precedes the quote alternatives and is mandatory for fixed pragma names. |
| parser.ts | 10735:29 | TS2322 | B: quote groups 2 and 3 are alternative branches; `group2 || group3` loses empty single quotes. |
| parser.ts | 10737:45 | TS18048 | B: same value can be undefined; captureSpan reads value.length. |
| parser.ts | 10741:25 | TS2322 | B: same value can be undefined; non-span argument receives it. |
| parser.ts | 10769:18 | TS2532 | M: both callers require group 1; see pragma expressions below. |
| parser.ts | 10794:9 | TS2322 | R: `/\s+/` has no captures, but indexes can be absent; existing required-argument guard proves current calls. |
| semver.ts | 149:25 | TS2345 | M: versionRegExp group 1 is required; optional enclosing suffix starts after it. |
| semver.ts | 283:30 | TS2345 | M: hyphenRegExp groups 1 and 2 each require `[a-z0-9-+.*]+`; neither branch is optional. |
| semver.ts | 287:48 | TS18048 | D: whitespaceRegExp `/\s+/` has no captures. |
| semver.ts | 288:48 | TS2345 | H: rangeRegExp group 1 is optional; parseComparator explicitly handles case undefined, but its parameter is declared string. Group 2 is mandatory. |
| semver.ts | 302:20 | TS2345 | M: partialRegExp group 1 is required, even with its internal alternative. |
| semver.ts | 302:42 | TS2345 | M: same major capture at parseInt. |
| semver.ts | 303:20 | TS2345 | M: same major capture at isWildcard. |
| semver.ts | 304:20 | TS2345 | M: same major capture at isWildcard. |
| semver.ts | 319:21 | TS2345 | M: leftResult.major originates in mandatory partialRegExp group 1. |
| semver.ts | 323:21 | TS2345 | M: rightResult.major originates in mandatory partialRegExp group 1. |
| semver.ts | 339:21 | TS2345 | M: result.major originates in mandatory partialRegExp group 1. |
| sourcemap.ts | 394:20 | TS2532 | M: `/^\/\/[@#] source[M]appingURL=(.+)\r?\n?$/` requires group 1; trailing newline optionality does not bypass it. |
| utilities.ts | 1330:53 | TS18048 | D: `/\r\n|\n|\r/` has no captures; callback receives strings. |
| utilitiesPublic.ts | 713:94 | TS2345 | M: locale group 1 `[a-z]+` required; optional territory suffix starts after it. |
| utilitiesPublic.ts | 714:36 | TS2345 | M: same required language capture. |

There are 16 M diagnostics, 8 D diagnostics, 3 B diagnostics, 1 H diagnostic,
and 1 R diagnostic. These are diagnostic counts, not independent edit counts.

## Expressions and legitimate absence

versionRegExp is
`/^(0|[1-9]\d*)(?:\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*)(?:-([a-z0-9-.]+))?(?:\+([a-z0-9-.]+))?)?)?$/i`.
Group 1 is mandatory. Groups 2 through 5 are optional; destructuring defaults
for minor, patch, prerelease and build already handle their absence.

partialRegExp is
`/^([x*0]|[1-9]\d*)(?:\.([x*0]|[1-9]\d*)(?:\.([x*0]|[1-9]\d*)(?:-([a-z0-9-.]+))?(?:\+([a-z0-9-.]+))?)?)?$/i`.
Group 1 is mandatory. Minor and patch already default to `"*"`; prerelease and
build are allowed absent by Version's constructor defaults. Required major
should be asserted at its binding source rather than at every downstream use.

hyphenRegExp is `/^\s*([a-z0-9-+.*]+)\s+-\s+([a-z0-9-+.*]+)\s*$/i`.
Both captures are mandatory. rangeRegExp is
`/^([~^<>=]|<=|>=)?\s*([a-z0-9-+.*]+)$/i`: group 1 is optional, group 2 mandatory.
The existing `case undefined:` at semver.ts:378 handles plain comparators.
A truthful operator parameter would include undefined; asserting group 1 would
claim something demonstrably false. The adapter makes precisely that parameter declaration change.

The locale expression is `/^([a-z]+)(?:[_-]([a-z]+))?$/`.
Language is mandatory; territory is optional and already passed to a parameter
of type string | undefined.

addPragmaForMatch has two caller expressions:
`/^\/\/\/?\s*@([^\s:]+)((?:[^\S\r\n]|:).*)?$/m` and
`/@(\S+)(\s+(?:\S.*)?)?$/gm`. Each requires group 1 and permits group 2 absent.
getNamedPragmaArguments accepts optional text and handles missing text with {}.

The named-argument constructor is
`new RegExp("(\\s" + name + "\\s*=\\s*)(?:(?:'([^']*)')|(?:\"([^\"]*)\"))", "im")`.
The pragma names used upstream are fixed names without regex metacharacters.
Group 1 participates in either quote branch. Group 2 is absent on double
quotes; group 3 is absent on single quotes. Neither can receive a mandatory
assertion. For ` path=''`, group 2 is `""`, group 3 is undefined, and the original
`matchResult[2] || matchResult[3]` evaluates to undefined. With captureSpan,
value.length throws; without it, an undefined value is stored. A nullish choice
would change behavior, so this soundness layer retains `||` and only adds the
authorized whole-expression assertion. Adaptation 46 owns the subsequent fix.

parser.ts:10794 is not a capture proof. Current non-XML pragma definitions in
types.ts:10269-10294 have one required factory argument, and missing required
args return "fail" before assignment. The helper's general optional-argument
path could store undefined in a string-valued map, but current non-XML callers
have no optional argument. The split density assertion removes undefined from the element type under
upstream's config. The required-argument guard remains unchanged; density is
not claimed to prove arbitrary index bounds.

## Probes and mutant

Run against a pristine pinned checkout:

```sh
source /workspace/adamic-tools/env.sh
node stage3/adapt/45-regex-captures/probe.cjs /tmp/regex-captures-upstream > /tmp/regex-captures-probe.log 2>&1
```

The probe reads the actual named-argument constructor and value expression from
upstream rather than substituting a repaired expression. It checks nonempty
single quotes, empty double quotes, and empty single quotes, and confirms the
last produces undefined and a TypeError on value.length. It also reads the
actual rangeRegExp and confirms the absent operator and existing handler.

The mutant of debug.ts's name expression is
`/^(?:function\s+([\w$]+)\s*\(|anonymous)$/`.
The actual analyzer and adapter decline group 1 because the `anonymous`
alternative bypasses it; the adapter exits before any writes. Node successfully matches `anonymous` with group 1 undefined.
A deliberately incorrect expected capture is caught by the probe assertion.
test.cjs additionally runs this mutant through the actual adapter, checks all
input hashes remain unchanged, declines an optional-capture split mutant, and
proves idempotence with both zero edits and identical file hashes.

## Latent bug and scope

The public Node counterexample is `/// <reference path='' />`, passed to
`ts.createSourceFile("empty.ts", text, ts.ScriptTarget.Latest)`. Upstream and
adaptation 45 alone throw `TypeError: Cannot read properties of undefined
(reading 'length')` at src/compiler/parser.ts:10737, inside extractPragmas.
`logs/parser-public-before.log` records the real adapted compiler stack.
Adaptation 46 fixes it; [the upstream ledger](../../upstream/LEDGER.md) contains
the issue draft for Kirk. No issue has been filed.

No native tsc or full uncached Adamic integration gate is claimed. The base
compiler still refuses non-null assertions and module cycles during lowering;
the requested checked native unwrap is a compiler-owner dependency, not a
runtime result measured by this source unit. No compiler implementation files,
other worker's adaptation, or reference baselines were changed. This adaptation
unit does not add a fixture bucket or edit fixtures_test.go.

## Recorded verification

`node --check probe.cjs` exited 0; the Node probe command above exited 0.
`logs/probe.log` records its exact stdout. An independent Python comparison
loaded the upstream-config census rows from the fetched census branch and
compared the multiset of (file, line, column, code) to the Markdown table:
29 matched. Removing one table row failed that equality assertion; see
`logs/ledger.log`. `git diff --check` exited 0. Those logs record the initial review. Final validation and proof-engine tests
are recorded below; they supersede the earlier stopped status.

`bash cloud/setup.sh > /tmp/regex-captures-setup.log 2>&1` exited 0.
The printed environment file `/workspace/adamic-tools/env.sh` was sourced.
Go go1.27.1 ready at 0s; clang 20.1.8, Node v24.19.0, and submodules ready at
1s; build cache warm at 187s; total 187s. `nproc` printed 5; cgroup cpu.max
was `400000 100000`. The complete setup stdout is in `logs/setup.log`.

## Soundness-layer default oracle

Before the later fork fix, the 10+45 tree ran the full unfiltered default oracle:
106,367 passing, 0 failing, 0 pending; all install/build/test exits 0;
`baseline.diff` 0 bytes. The run completed before an attempted stop for the
superseding fix, so no process was stopped. Its independent report is saved in
`evidence/soundness-oracle-report.json`; elapsed wall time 700.927s. Install
7.341s, build 97.760s, tests 595.757s. The final combined proof follows below.

## Upstream-config census

The unchanged census driver and loader overlay from `a3ef0dc` checked the
matched 10-only control and the post-build 10+45+46 tree. The control reproduced
exactly 29 diagnostics: TS2345 15, TS18048 6, TS2322 4, TS2532 4. The final tree
has **0 checker diagnostics**, on the whole program and on each of its 78 source
entries. Each run has 81 per-file attempts plus the 78-root whole-program row.
Raw compressed records are `evidence/census-before.jsonl.gz` and
`evidence/census-after.jsonl.gz`; summary and precise pins are alongside them.

The 78 single-source entries now reach the existing import-cycle refusal.
The whole program reaches `lower: stage 0 compiles a program from one entry
file, got 78`. These are observed lowering limits, not checker diagnostics and
not native success. The three JSON input attempts remain extension errors.
No checker option or sound regex declaration was disabled.

The initial Go build encountered VCS-stamping failure because the isolated
worktree shares its pinned cohere checkout through a symlink. Retrying with
`-buildvcs=false` succeeded. That changes artifact stamping only; source,
checker configuration, and the census overlay are unchanged.

## Mutants and catches

| Mutant | Dedicated catch |
| --- | --- |
| Debug regex alternative `anonymous` bypasses group 1 | Participation proof and actual adapter decline; successful Node match has undefined group 1; all input hashes unchanged. |
| Split separator `/\r\n?|(\n)?/` inserts an optional capture | Capture-free density guard declines; `"a\r\nb"` really produces `["a", undefined, "b"]`; no writes. |
| JSX positional argument becomes optional | Caller-invariant guard declines before the map element assertion; no writes. |
| Pragma name adds `(extra)` and changes capture numbering | Both adapters' fixed-name guards decline; parser bytes unchanged. |
| Debug assertion removed after a zero-edit observation | Independent source-hash comparison fails. |
| Fork fix puts `||` back | Empty-single-quote selector expectation fails; real original compiler also fails the fixed public-API observation. |
| One review-table row omitted (initial review) | Exact census location/code multiset comparison fails. |
| Wrong expected alternative-capture value (initial review) | Node observation assertion fails. |

The direct adapter suites record their catches in `logs/adapter-tests.log` and
adaptation 46's logs. Unsupported regex constructs are also declined by direct
participation tests. The final evidence audit separately mutates the retained
diagnostic, baseline-diff bytes, and oracle filter fields; these are record
mutants of the audit, not claimed mutations of the upstream oracle itself.

## Final combined proof

The final 10+45+46 tree, built from the requested `a3ef0dc` checkout, passes
all default oracle runners: **106,367 passing, 0 failing, 0 pending**.
The run has no runner/test filter, uses 4 workers, and npm's test script supplies
`--light=false`; lint=false remains the base oracle's default. All phase exits
are 0. Install 2.924s, build 34.124s, tests 572.645s, total wall 609.787s.
Adaptation 46 saves its complete oracle report and the **0-byte baseline.diff**.
No reference baseline was edited, accepted, or updated. No existing baseline
was found to cover the buggy empty-single-quote behavior.

After the full build, suite and census, idempotence.py reapplied both adapters.
Each reported **0 files, 0 edits**. SHA256 maps of **82,834 files** were identical;
only .git and node_modules were excluded. `evidence/idempotence.json` records
that result. The audit in verify.py passes using the committed compressed
census records and adaptation 46's oracle evidence; it also catches independent
retained-diagnostic, nonempty-baseline, and filtered-oracle record mutants.

The initial guard/erasure test compared emitted JS text and failed on the
explicitly requested parentheses around `||`. Its corrected check compares
parsed JS ASTs modulo parentheses and passes. The first fix probe lacked an
initializer-presence guard in its AST walk; that probe was corrected and
rerun. Those were test-harness corrections, not changes to the fork result.

Commands run from the isolated validation checkout (after copying only these
unit directories there), with all observations redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
STAGE3_CACHE=/tmp/regex-captures-cache stage3/apply.sh /tmp/regex-captures-fixed > /tmp/regex-captures-fixed-apply.log 2>&1
stage3/oracle/run.sh /tmp/regex-captures-fixed /tmp/regex-captures-fixed-oracle > /tmp/regex-captures-fixed-oracle.log 2>&1
python3 stage3/census/make_overlay.py /tmp/regex-captures-overlay
go build -buildvcs=false -overlay=/tmp/regex-captures-overlay/overlay.json -o /tmp/regex-captures-census ./stage3/census/tool > /tmp/regex-captures-census-build-retry.log 2>&1
CENSUS_CONFIG=/tmp/regex-captures-baseline/src/compiler/tsconfig.json /tmp/regex-captures-census /tmp/regex-captures-baseline/src/compiler /tmp/regex-captures-census-before.jsonl > /tmp/regex-captures-census-before.log 2>&1
CENSUS_CONFIG=/tmp/regex-captures-fixed/src/compiler/tsconfig.json /tmp/regex-captures-census /tmp/regex-captures-fixed/src/compiler /tmp/regex-captures-census-after.jsonl > /tmp/regex-captures-census-after.log 2>&1
```

The control tree uses the same source pin, 00+10, regenerated diagnostics and
the same locked node_modules, without 45/46. The census build uses the shared
pinned cohere/typescript-go; it does not change their sources. The before run
finished before the after run began. The final census began after compiler
build and completed during the suite; all source files checked were the same
post-build source. Idempotence ran after both completed.

Commands from this branch for the independent proof checks:

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH=/tmp/regex-captures-cache/api/node_modules node stage3/adapt/45-regex-captures/test.cjs /tmp/regex-captures-upstream > /tmp/regex-captures-tests-reviewed.log 2>&1
NODE_PATH=/tmp/regex-captures-cache/api/node_modules node stage3/adapt/46-fix-pragma-empty-argument/test.cjs /tmp/regex-captures-upstream > /tmp/regex-captures-fix-tests-name-mutant.log 2>&1
NODE_PATH=/tmp/regex-captures-cache/api/node_modules python3 stage3/adapt/45-regex-captures/idempotence.py /tmp/regex-captures-fixed > /tmp/regex-captures-idempotence.log 2>&1
python3 stage3/adapt/45-regex-captures/verify.py stage3/adapt/45-regex-captures/evidence/census-before.jsonl.gz stage3/adapt/45-regex-captures/evidence/census-after.jsonl.gz stage3/adapt/46-fix-pragma-empty-argument/evidence > /tmp/regex-captures-verification.log 2>&1
```

Every command above exited 0. `node --check` on the adapters/proof/public probe
and `git diff --check` also pass. No Go package changed; no additional package
tests or full uncached Adamic gate were run. Setup remains the recorded 187s,
Go 1.27.1, clang 20.1.8, Node 24.19.0, `nproc=5`, CPU quota 4. No OOM or
OOM-kill event occurred during these runs.
