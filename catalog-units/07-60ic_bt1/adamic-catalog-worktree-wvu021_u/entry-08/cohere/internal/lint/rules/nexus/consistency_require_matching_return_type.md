# `nexus/consistency-require-matching-return-type`

| | |
|---|---|
| **Recommendation** | **Yes**, in place of `consistent-return` (already `off` in ahra) |
| Findings in ahra | **9** (measured 2026-10-01): 2 bare `return;` in a value function, 7 `return undefined;` in a void function |
| Subject matter in ahra | 2,291 bare returns and 252 `return undefined;` judged; 32 of them in functions whose type cannot say |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | yes, both directions, canonical text only |
| Needs type information | yes, the verdict is the function's return type |

## What it checks

**A bare `return;` is allowed only where the function's return type says `void`. Everywhere else
every `return` carries a value, so `undefined` is written `return undefined;`.** And the converse,
so there is one shape everywhere: in a function whose type says `void`, the early exit is `return;`,
never `return undefined;`.

```ts
function find(): string | undefined { if(!ready) return; return name; }           // reports
function find(): string | undefined { if(!ready) return undefined; return name; } // clean
function save(): void { if(!ready) return; write(); }                              // clean
function save(): void { if(!ready) return undefined; write(); }                    // reports
useEffect(function() { if(!ready) return; start(); return function() { stop(); }; }); // clean
```

`void` means no value; `undefined` is a value. The spelling at the return tells the reader which kind
of function they are in without scrolling to the signature.

It judges `return` statements only. A function's fall-off end is never reported; that is
`noImplicitReturns`' job, and it is why an exhaustive `switch` with no `default` is clean here.

## Why this replaces `consistent-return`

`consistent-return` cannot see types. It fires on an exhaustive `switch` whose end the checker proves
unreachable, and it cannot tell a `void` callback (a React effect that exits early and otherwise
returns its cleanup) from a function that returns `undefined` as a value. In ahra it is `off`.

## Which return type

1. **Declared**: the function's own annotation, final even when it is `any`.
2. **Contextual**: for a function expression, an arrow function, or an object-literal method with
   no annotation, the return type of the single call signature its context expects. A union context
   contributes each member's single signature; a member with none (the `undefined` of an optional
   callback property) is skipped; a member with several makes the context ambiguous.
3. **Inferred**: the function's own signature. Overloads are used only when they all agree.

A contextual type that is `any`, `unknown`, `never` or a bare type parameter says nothing and is
passed over for the inferred type.

## Decisions

| Case | Verdict | Why |
|---|---|---|
| Declared `any` / `unknown` | no finding either way | the author declined to say; the rule cannot know |
| `never` | no finding | no `return` is reachable without a type error, so there is no spelling to offer |
| A bare type parameter `T` | no finding | it could be instantiated as `void` |
| Async | the awaited type | `Promise<void>` is void, `Promise<string \| undefined>` a value; a non-async function returning a promise is not unwrapped |
| Generator | its TReturn (second type argument of `Generator`, `Iterator`, ...) | that is what `return` sets; an unannotated generator with bare returns infers `void` |
| Constructor, setter | void | `return` in them is only an early exit; `return undefined;` reports |
| Getter | an ordinary function | its type is the property's |
| Declared `(): undefined` | a value | the author said `undefined` is the value |
| Contextual or inferred exactly `undefined` | no finding | see below |
| `return void expr;`, `return (undefined)` | `void expr` is left alone; `(undefined)` reports in a void function, without a fix | only the identifier `undefined` resolving to the global counts |

### Why an undeclared `undefined` is unknowable

An inferred type is the function's own returns read back, and a generic context (`forEachChild`'s
`T | undefined`, `Array.from`'s mapper `U`) is instantiated from the callback it types. When either
comes out as exactly `undefined`, it records the spelling rather than an intent, and both readings
were wrong on ahra:

- Read as a **value**, the first measurement reported 28 of 36 findings as noise: 27 exits of one CLI
  dispatcher (`modules/phi/PhiSocialMediaCommandLineInterface.ts`) whose `return void console.log(...)`
  made its inferred type `Promise<undefined>`, and one `forEachChild` visitor.
- Read as **void**, the second measurement reported `Array.from({ length }, function() { return
  undefined; })` (`useGraphQlInfiniteScroll.ts:342`) and two stub methods standing in for
  `() => X | undefined` (`useForm.tsx:97`, `:100`). There `undefined` is the value, and the fix would
  retype each callback as `void` and break its consumer.

## The fixer

Both directions are behaviour-preserving: `return;` and `return undefined;` return the same value.
The fix is offered only on canonical text (`return;`, or `return undefined;` with only whitespace
between), so no comment is moved and no automatic-semicolon-insertion boundary changes: inserting
` undefined` after a semicolon-less `return` could join the next line into the expression. In the
value direction it is also withheld when the type does not admit `undefined` (`(): number`), where
`return undefined;` would surface a type error.

## Findings in ahra (2026-10-01)

    modules/google/ads/GoogleAdsCampaignApi.ts:630                       bare, Promise<GoogleAdsCreateCampaignResultInterface | undefined>
    libraries/structure/libraries/nexus/source/collections/Array.ts:57    bare, T | undefined (getRandom's implementation)
    libraries/structure/source/components/animations/LineLoadingAnimation.tsx:55   undefined, useEffect
    libraries/structure/source/components/animations/LineLoadingAnimation.tsx:56   undefined, useEffect
    libraries/structure/source/ops/developers/assets/components/dialogs/UploadAssetsDialog.tsx:120   undefined, useEffect
    libraries/structure/source/ops/developers/metrics/MetricsBottomPanel.tsx:114   undefined, useEffect
    libraries/structure/source/ops/developers/metrics/DataSource.tsx:86   undefined, useEffect
    libraries/structure/source/modules/account/authentication/components/dialogs/AccountMaintenanceDialogBody.tsx:91   undefined, useEffect
    libraries/structure/source/utilities/react/hooks/useMeasuredElement.tsx:210   undefined, useEffect

Measured with a scratch copy of `CohereSettings.json` enabling the rule, `--no-fix --lint`, 0 crashed
files. The subject-matter split came from a temporary build that reported every bare and
`return undefined;` return with its verdict: 2,262 bare in void functions, 2 bare in value
functions, 27 bare unknowable; 240 `return undefined;` in value functions, 7 in void functions, 5
unknowable. A text grep over the same 3,605 files finds 2,298 `return;` and 254 `return undefined;`,
the difference being text in comments and strings.

## Status: registered, NOT enabled

Enabling it is adding `"nexus/consistency-require-matching-return-type": "error"` to
`CohereSettings.json`. All nine findings sit on canonical text in a type that admits the rewrite, so
each should carry the fix; that was read from the source, not observed through a fix run, which
would have written to the tree.
