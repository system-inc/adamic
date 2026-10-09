# `@typescript-eslint/no-misused-spread`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **2** (measured 2026-10-01, after dropping the string branch; the gate adds the four string sites below) |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow using the spread operator when it might cause unexpected behavior

## Deliberate divergence: no string branch

Upstream also reports a string spread into an array or a call (`noStringSpread`) and asks for
`Intl.Segmenter`, because spread yields code points rather than the graphemes a reader sees. Cohere
drops that branch, by Kirk's ruling of 2026-10-01 (review option C). Whether a string should be
walked by code point or by grapheme is intent the rule cannot see; spread is the correct code-point
iteration, and upstream's only repair, `Array.from(text)`, is the same iteration under another name,
so every finding either launders or is wrong. The object cascade (Promise, function, Map, array,
iterable, class instance, class declaration), which catches silent data loss, is unchanged.

So cohere is silent where the gate reports, on the four string sites ahra had:

| site | why the code wants code points |
|---|---|
| `libraries/structure/libraries/nexus/source/geography/Countries.ts:13` | `ISO` letters mapped to regional indicators |
| `modules/openai/PngTextMetadata.ts:64` | a Latin-1 filter |
| `modules/pensieve/PensieveDailyQuotes.ts:163` | quote-mark scanning |
| `modules/pensieve/PensieveDailyQuotes.ts:201` | ignored characters filtered out |

ESLint draws the same line through Nexus's `NexusTypeScriptEsLintPlugin`, whose wrapper of this rule
drops `noStringSpread` (Nexus 4e3395d), so the engines agree on all four (#cn8sthd). The fixture is
`TestNoMisusedSpreadLeavesStringSpreadAlone`: upstream invalid 0 through 13 verbatim, the four ahra
sites, the constructor shape, and a control proving the object cascade still reports in the same
harness. The `allow` option is kept for parity of the option surface; nothing in ahra sets it.

## Deliberate divergence: a class whose copy is complete

Upstream reports every spread of a class instance, because the copy drops the prototype. Cohere
reports one only when the prototype holds something to lose. By @system_cohere's ruling of
2026-10-03, from api's decorated GraphQL and Serializable data classes (#ynneze5), an instance is
exempt when its class and every base class in its chain:

- are declared only as classes in the program's own source: not ambient (`declare class`), not in a
  declaration file, not in a package, and not merged with an interface, which can describe methods a
  mixin installs;
- declare no instance method, get or set accessor, or auto-accessor (static members live on the
  constructor, which an instance never carried);
- declare no `#private` member, which a spread does not copy;
- carry only decorators (on the class, a member or a constructor parameter) that resolve to our own
  source.

A union is exempt only when every class instance in it is. The remaining trust is the last clause's:
one of our own decorators could install a prototype accessor at runtime, which no type shows, and
the copy would silently lack it. Base's and api's decorators register metadata only, and third-party
decorators such as MobX's `@observable` or Lit's `@property` stay outside the exemption.

So cohere is silent where ESLint reports on upstream's invalid 69 through 76 and 83, every one a
fields-only class, and on api's 24 spreads of its data classes. ahra and www have none, measured
before and after on 2026-10-03. The fixture is `TestNoMisusedSpreadExemptsAClassWhoseCopyIsComplete`:
the nine upstream cases verbatim, then each clause with the case it exempts and the case it must
still report.

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

5 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/security/secrets/Secret.test.ts:70`**

```
const spread = { ...secret };
```

> Using the spread operator on class instances will lose their class prototype

**`libraries/structure/source/modules/account/hooks/useAccount.ts:106`**

```
...previousAccountState.data,
```

> Using the spread operator on class instances will lose their class prototype

**`libraries/structure/source/services/network/NetworkService.ts:359`**

```
...options?.headers,
```

> Using the spread operator on an array in an object will result in a list of indices

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

