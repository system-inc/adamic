# `nexus/correctness-no-leaked-number-render`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` once the three Structure sites are fixed.** 3 findings in ahra, 0 false |
| Findings | **ahra 3**, all in `libraries/structure/` (measured 2026-10-03) |
| Measured precision | 3 of 3 true to the type. 2 are real bugs, 1 is true but harmless (a `Date.now()` timestamp that is never `0`) |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, the whole rule is a question about the operand's type; without a checker it registers nothing |

## What it checks

An `&&` whose result is rendered as a JSX child (`<p>{...}</p>`, `<>{...}</>`), where the operand `&&`
returns when it is falsy is typed as a number that can be `0`: `number`, `bigint`, a literal union holding
`0`, a numeric enum with a zero member, a branded number, or a type parameter constrained to one of these,
optionally beside `null`, `undefined`, `void` and `boolean`. React renders that `0` (or `NaN`) as text. The
doc comment carries the precise walk: which operands of a chain, a ternary, `||` and `??` are followed.

## Where it came from

`libraries/structure/source/components/navigation/pagination/PaginationControls.tsx:109`:
`{properties.itemsTotal && (` prints a bare "0" in the table footer when a table has no records. Found by
the JS-catalog pass of the new-rules sweep (`#tevhg3f`, after sonarjs S6439 and Biome `noLeakedRender`),
built as `#1g2wqpw`.

## Not `react/jsx-no-leaked-render` configured

eslint-plugin-react's rule asks the same question without types. It was measured and declined at 428 ahra
findings (`react/jsx_no_leaked_render.md`, a document with no `.go` beside it: never ported), because it
reports every string, object and boolean operand too. Its options choose a fix strategy (`ternary`,
`coerce`), not which operand types count, so no configuration of it reaches this rule's three findings.
This is a new, typed rule; the untyped one stays declined.

## What it declines

- **A union with any part that is not a number, a boolean or nullish**: `string`, an object, `any`,
  `unknown`, an unconstrained type parameter. `ReactNode` contains `number`, and reporting it would add 34
  findings in ahra today (`Button.tsx:184` `iconLeft &&`, `AccordionItem.tsx:62` `properties.icon &&`,
  and so on), every one a node or label that is meant to be rendered, `0` included. `string | number` is
  declined for the same reason. The research probe excluded these as its 39 false positives.
- **A non-zero literal**: `1 | 2 | 3`, a constant `3`, an enum with no zero member. These cannot leak.
- **An attribute** (`visible={count && true}`), **a spread child**, an `&&` outside JSX, and the right side
  of `&&` or the left side of `||` (both rendered only by intent).
- A number known positive only by a guard the type cannot carry (`if(count > 0) { ... {count && 'x'} }`)
  is still reported: the type says `number`. No such site exists in ahra; the fix states the intent.

## Reconciling with the research count

Research counted **3** (0 in ahra's own code, 3 in Structure, plus 39 `ReactNode` hits excluded by hand).
cohere counts **3**, the same three sites. The `ReactNode` exclusion is now made by the type test rather
than by hand: the same rule with the decline removed counts 37 in ahra, 3 plus 34 `ReactNode` and
`string | number` operands. The 5 between 34 and 39 is drift since the research ran.

## The findings, read one by one

| Site | Type | Reading |
|---|---|---|
| `libraries/structure/source/components/forms/fields/multiple-checkbox-grid/FieldInputMultipleCheckboxGrid.tsx:118` | `number \| undefined` | **Bug**: `0` means no cap (the toggle handler at `:54` tests it the same way) and prints "0" under the grid |
| `libraries/structure/source/components/navigation/pagination/PaginationControls.tsx:109` | `number \| undefined` | **Bug**: an empty table prints "0" where "0 records" or nothing was meant |
| `libraries/structure/source/ops/developers/web-sockets/WebSocketsPage.tsx:181` | `number \| null` | **True, harmless**: disconnected is `null` and the value is `Date.now()`, so the `0` never occurs; `!== null` says what it means |

## Verification

- Fixtures from the three real sites, each before (one finding on the operand) and after (a boolean
  condition, silent). 21 firing shapes and 25 silent shapes beside them.
- Mutation check, each mutant a copy through `go test -overlay -count=1`, each killed: a declining part read
  as silent (`a ReactNode`, `a string or a number`); the JSX-child test removed (`an attribute, not a
  child`); non-zero literals read as leaking (`a number literal union with no zero`, `a non-zero
  constant`, `a numeric enum with no zero member`); the left of a nested `&&` not followed (`both operands
  of a chain`); the left of `||` followed when rendered (`an and that an or catches`) and when falsy (`a
  number on the left of an or in the operand`); the intersection arm dropped (`a branded number`); the
  constraint arm dropped (`a generic constrained to number`); a ternary's true branch not followed
  (`inside a ternary branch`); the spread test removed (`a spread child`).
