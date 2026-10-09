# `nexus/correctness-no-nullish-stripping-assertion`

| | |
|---|---|
| **Recommendation** | **Yes, registered `off` until the sites are fixed.** 87 findings on ahra, 0 false. Turning it on at `error` today puts 87 findings on ahra |
| Findings | **ahra 87** (55 in `app/` and `modules/`, 32 in `libraries/structure/`, 11 of those in tests), measured 2026-10-03 |
| Measured precision | 87 of 87 true by construction: in each, the checker's type for the asserted expression holds `undefined` or `null` and the asserted type cannot. Live bugs include the 21 `AnalyticsApi` / `GoogleAds*` flag reads and `AhraOsReactions.ts:837`; 30 sit in front of a `??` / `||` fallback and are harmless for the `undefined` itself |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes; without a checker the rule registers nothing |

## What it checks

`x as T` (and `<T>x`) where the type of `x` at the assertion, after narrowing, is a union holding
`undefined` or `null` beside at least one other constituent, and `T` has no constituent that could be
nullish. It is the zero-`!` ruling (`@typescript-eslint/no-non-null-assertion`, at error) applied to
the other door: `x as T` makes the same unchecked presence claim `x!` does.

## The partition with `@typescript-eslint/non-nullable-type-assertion-style`

That rule owns the **pure removal**: an assertion whose asserted constituents are exactly the original's
non-nullish constituents, by type identity (`maybe as string` over `string | undefined`). This rule
declines exactly that case, using the style rule's own predicate (`isSameTypeWithoutNullish`, copied
because it is unexported), so the two never report the same cast. A test runs both rules on the same
casts and asserts each reports only its own.

The style rule is **off** in ahra (`NexusCohereSettings.json`). Enabled for this measurement it found
**0**: the 55 pure removals its audit doc lists were already cleaned up. Its message tells the author to
write `!`, which `no-non-null-assertion` bans, so if it is ever turned on its message needs to change.

## Where it came from

The cross-language pass of the new-rules sweep (`#tevhg3f`, after clippy `unwrap_used` and Kotlin `!!`),
built as `#c3afxe9`. Fixtures are taken from three real sites, before and after:

- `modules/google/analytics/AnalyticsApi.ts:175` (`getDateRange`, `getDaysLabel`): `parseInt(options.days
  as string, 10)`. The research's lead site, and a live bug: AnalyticsApi's own `parseOptions` stores
  `--days 30` (space form) as `true`, so `parseInt(true)` is NaN and the default silently wins. Fixed with
  `numberFlag`.
- `modules/os/sensation/AhraOsReactions.ts:837`: `typeof x.configuration === 'object' ? (x.configuration as
  Record<string, unknown>)`. `typeof null === 'object'`, so a null configuration passes as a record. The
  outer read eight lines up guards with truthiness and is the silent counterpart in the same fixture.
- `libraries/structure/source/components/dialogs/DialogMenu.tsx:53`: `event.target as HTMLElement` over
  `EventTarget | null`. Fixed with `instanceof`.

## What it declines

- the pure removal (the style rule's)
- `any` / `unknown` originals, and `any` / `unknown` / `void` targets
- an all-nullish original (`null as ResponseType`, `undefined as 'Ready'`): a placeholder, not a strip
- `as never`: an escape hatch like `as any`; `no-unsafe-type-assertion` owns it. 3 ahra sites
  (`FinancePositionCommandLineInterface.ts:487`, `:648`, `GoogleAdsKeywordApi.ts:282`)
- a generic target whose constraint could be nullish, or that has none (`source.get(key) as ValueType`)
- `void` in the original: a result to ignore, not an absent one
- an assignment target, `(holder.value as string) = 'x'`
- `as const` (no guard needed: the asserted type is the expression's own, so it keeps the nullish)

It **reports** `(x as T) ?? fallback` and `(x as T) || fallback`, though the fallback catches the
`undefined` at runtime. That matches the house's `!` contract (`x! ?? y` is an error under
`no-non-null-asserted-nullish-coalescing`), and the cast makes the fallback look dead to the compiler.
Declining it would also be exact and would drop about 30 findings (the `|| default` flag reads in
`GoogleAds*`, `Kling*`, `OpenAi*`, `Markdown.tsx`, `OpsSupport*`, `TranslationUtilities.ts`); that is a
judgment for the Square.

## Reconciling with the research count

Research counted **62** (60 general, 2 on `.get()`) with a TypeScript-API probe (`probe6.js`) over
`app/` and `modules/` only. cohere counts **55** there, **87** with `libraries/structure/`:

- **-3, `as never`**: declined by design (above). The research counted them.
- **-1, `null as ResponseType`** (`LinkedInClient.ts:394`): all-nullish placeholder into a type
  parameter. The research counted it.
- **-1, `FigmaApi.ts:89`**: the code changed since the probe; the cast is gone.
- **-2, the `.get()` pair**: not in the probe's saved output (`ahra_out6.txt`, 60 lines, all "other");
  the 62 in the report includes two the file does not list. Unreconciled; the probe's file list was
  not saved.
- Every other probe line matches a finding here after line drift (`MetaMarketing` -2, `OpenAi` and
  `Kling` renumbered, `XAdsApi.ts:68` is now `:101`, `AhraOsTriggers.ts:1474` is now `:1491`).
- **+32, `libraries/structure/`**: outside the probe's file list. 11 in tests.

The research also cited `FinanceSqliteDatabase.ts:5390` with the reason "a missing row becomes
undefined". The row is `Record<string, unknown>`, so the plain `row.value_cents as number` reads are
silent (unknown is not this rule's). `:5390` reports for a narrower reason: `=== null` narrows `unknown`
to `{} | undefined` on the other branch, and the cast drops that `undefined`.

The per-site list with fixes is `sites.tsv` in the build scratch directory.

## Verification

- 6 real-site tests (3 before, 3 after), 12 firing shapes, 23 silent shapes, 7 partition cases run
  against both rules: 51 passing subtests.
- Mutation sweep, each mutant a copy through `go test -overlay`, all 10 killed: pure-removal exclusion
  removed, original-has-nullish removed, all-nullish decline removed, asserted-side test removed,
  `never` decline removed, assignment-target decline removed, instantiable branch removed,
  unconstrained-answers-nullish flipped, `void` target removed, `any`/`unknown` target removed.
