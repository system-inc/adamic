# Regex capture review: stopped before adaptation

This is a review deliverable, not a completed adaptation. No adapt.cjs is installed.
The 29 upstream-config diagnostics have not been removed. No adapted-tree census,
stage 3 baseline oracle, or idempotence test was run. No native compilation was
attempted, so there is no native miscompile observation.

The requested all-29 result cannot be established using mandatory-capture
assertions alone: parser.ts has a real empty-single-quote bug, and semver.ts
already accepts an absent operator despite declaring its parameter as string.
I stopped after the review and counterexamples instead of silently fixing the
bug or asserting optional groups. The mandatory subset remains unimplemented.

Source: TypeScript v6.0.3, commit
050880ce59e30b356b686bd3144efe24f875ebc8. Census: origin/codex/tsc-census,
profile upstream-config. All positions below refer to that unadapted source.
Counts: TS2345 15, TS18048 6, TS2322 4, TS2532 4; 29 total.

## Proof rule

A successful match is not proof that every capture participated. A group must
participate on every successful path through the expression. A quantifier with
minimum zero on the group or an enclosing group, or an alternative which
bypasses it, defeats that proof. Alternation *inside* a mandatory capture does
not defeat it. An empty captured string is still a present string. These are
manual, expression-specific proofs, not an implemented general regex analyzer.

Split is a separate contract: regex split can insert undefined only through
nonparticipating captures in its separator. A separator with no capturing
groups inserts only substring strings. This proves element values, not the
presence of an arbitrary index in the resulting array.

## All 29 diagnostics

M = mandatory capture; D = split element density; H = absence already handled;
B = real runtime bug; R = required positional argument checked at the use.
No assertion or declaration patch has been applied to any row.

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
claim something demonstrably false. No such declaration change was made.

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
would change behavior, so it is deliberately not patched here.

parser.ts:10794 is not a capture proof. Current non-XML pragma definitions in
types.ts:10269-10294 have one required factory argument, and missing required
args return "fail" before assignment. The helper's general optional-argument
path could store undefined in a string-valued map, but current non-XML callers
have no optional argument. Any adaptation should document that caller invariant
or truthfully represent absence, not infer bounds from the separator regex.

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
The manual proof rule declines group 1 because the `anonymous` alternative
bypasses it. Node successfully matches `anonymous` with group 1 undefined.
A deliberately incorrect expected capture is caught by the probe assertion.
This validates the counterexample observation, not an automatic adaptation
engine or a native compiler check.

## Deliberately unfinished

No stock TypeScript compiler-API adapter, fixture bucket, stage-0 fixture status,
census rerun, baseline oracle, idempotence proof, or full gate is provided.
No compiler implementation files or other adaptation territories were edited.
The branch starts at current origin/main; prerequisite branches were read,
not merged. Completion needs the mandatory and split adaptations plus an
explicit decision on the reported bug; it cannot honestly be claimed by adding
assertions to every diagnostic location.

## Recorded verification

`node --check probe.cjs` exited 0; the Node probe command above exited 0.
`logs/probe.log` records its exact stdout. An independent Python comparison
loaded the upstream-config census rows from the fetched census branch and
compared the multiset of (file, line, column, code) to the Markdown table:
29 matched. Removing one table row failed that equality assertion; see
`logs/ledger.log`. `git diff --check` exited 0. No package tests were run for
this review-only deliverable.

`bash cloud/setup.sh > /tmp/regex-captures-setup.log 2>&1` exited 0.
The printed environment file `/workspace/adamic-tools/env.sh` was sourced.
Go go1.27.1 ready at 0s; clang 20.1.8, Node v24.19.0, and submodules ready at
1s; build cache warm at 187s; total 187s. `nproc` printed 5; cgroup cpu.max
was `400000 100000`. The complete setup stdout is in `logs/setup.log`.
