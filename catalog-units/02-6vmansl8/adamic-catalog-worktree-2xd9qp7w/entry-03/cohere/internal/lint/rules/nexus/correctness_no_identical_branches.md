# `nexus/correctness-no-identical-branches`

| | |
|---|---|
| **Recommendation** | **Yes**, at `error` once ahra's 9 findings are fixed. Registered `off` in ahra until then, because the Kling pair needs its owner's call (`#d5ryqgk`) |
| Findings | **ahra 9, www-phi-health 4, www-connected-app 4** (measured 2026-10-02; the Structure pair is shared by phi and connected) |
| Measured precision | 17 of 17 true: in every finding the condition chooses nothing |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, for the narrowing guard; without a checker the rule registers nothing |

## What it checks

A ternary, or an `if` with an `else`, whose branches are the same code, so the condition is silently
ignored. The rule's doc comment carries the precise version: structural equality with whitespace and
comments ignored, the `else if` chain judged from its last `if`, and the type guard that keeps
`typeof v === 'string' ? format(v) : format(v)` silent when the two calls resolve different overloads.

## Where it came from

`modules/kling/KlingApi.ts:156` and `:187` in ahra, `highQuality ? 'pro' : 'pro'`: a quality flag that
did nothing in either request builder (`#d5ryqgk`). Found independently by two research passes of the
new-rules sweep (`#tevhg3f`), built as `#052jffy`.

## The findings, read one by one

**ahra (9)**

| Site | Shape | Reading |
|---|---|---|
| `modules/kling/KlingApi.ts:156`, `:187` | `highQuality ? 'pro' : 'pro'` | **Bug.** The flag is ignored; the false side was meant to be `'std'` (`#d5ryqgk`) |
| `modules/meta/MetaMessagingApi.ts:52`, `:142` | `appKey === 'instagramApp' ? '/me/conversations' : '/me/conversations'` (and `/me/messages`) | Dead condition, possibly a bug: the Instagram branch may have been meant to differ |
| `modules/midjourney/MidjourneyApi.ts:590` | `.includes('.mp4') ? 'mp4' : 'mp4'` | Dead condition |
| `app/(os-layout)/finance/_components/FinanceChart.tsx:134` | `left: horizontal ? 8 : 8` | Dead condition |
| `modules/facets/FacetsWeeklies.ts:410` | `if(isCurrentWeek) { ... } else { ... }`, six identical lines each | Dead condition |
| `libraries/structure/code-quality/lint/rules/NetworkRequireHookOptionsParameterRule.ts:92` | the last three arms of an `else if` chain all set index 2, `'third'` | Two conditions ignored; reported from the first of the three |
| `libraries/structure/libraries/nexus/code-quality/lint/rules/ImportRequirePathAliasRule.ts:158` | `node.type === ImportDeclaration ? node.source : node.source` | Dead condition; merging compiles (a fixture) |

**www-phi-health (4)**: `HealthTimelineFormatters.ts:123` and `:202`, `Number.isInteger(count) ?
count.toString() : count.toString()`, which may have meant a fixed precision on the false side; plus the
Structure pair below.

**www-connected-app (4)**: `WeatherBadge.tsx:38`, `isDay ? 'Clear' : 'Clear'` (dead; the icon beside
it does differ); `PortsPage.tsx:97`, `isCommonPort ? '' : ''` (a class that was never written); plus
the Structure pair.

**Shared Structure (both apps)**: the ESLint copy of `NetworkRequireHookOptionsParameterRule.ts:83`
(the chain above) and `LocalStorageService.ts:223`, `typeof parsed.expiresAt === 'number' ? new
Date(...) : new Date(...)`, where the field is a `string`, so the first branch is `never` (a fixture).

## Declined, by measurement

**`switch`** is not covered (the doc comment gives the reason). A throwaway probe reporting a `switch`
with a `default` whose every non-empty case is token-identical once a trailing `break` is dropped
fired on a planted case and found **0** on all three trees.

## Verification

- Fixtures both ways: the real KlingApi builders before (2 findings) and after (`'std'`, silent), 20
  firing shapes, 19 silent shapes, and the JSX text pair.
- Mutation check, each mutant through `go test -overlay` and each killed: the type guard, the
  between-children token gap, the empty-branch skip, `never` matching anything, the `else if`
  widening, leaves compared as written, the spread-argument type, and the intrinsic-JSX skip.
