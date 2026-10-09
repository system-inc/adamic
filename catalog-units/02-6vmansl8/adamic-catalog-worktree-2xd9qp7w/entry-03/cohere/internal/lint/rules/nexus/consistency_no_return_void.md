# `nexus/consistency-no-return-void`

| | |
|---|---|
| **Recommendation** | **Yes** (Kirk's rulings on `no-confusing-void-expression` and the `void`/`undefined` return rule, 2026-10-01) |
| Findings in ahra | **10** at HEAD on 2026-10-01, all in `modules/phi/PhiSocialMediaCommandLineInterface.ts`; **0** in the working tree after the syntax wave |
| Measured precision | 10 of 10 |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | yes, for the canonical text |
| Needs type information | no |

## What it checks

A `return` statement whose expression is a `void` unary, through parentheses:
`return void console.log('Usage: ...');`. Write the expression as its own statement, then `return;`.

## Out of scope, deliberately

- **Arrow expression bodies.** `onClick={() => void handleAdd()}` has no `return` statement. It is the
  idiom for a callback that starts a promise and returns nothing, it appears on every button in
  `app/(os-layout)/see/`, and the ruling was about the statement. Widening to it would turn a
  decided idiom into findings.
- **`void` expression statements** (`void startServer();`), the dropped-promise marker.
- **`void` elsewhere inside a returned value** (`return [void a(), 1];`).

`return void 0;` is reported: it is `return undefined;` written to look like something else.

## The fix

`return void E;` becomes `E;` then `return;` on the next line at the same indentation when the
`return` is in a block or a case clause, `E; return;` when a case label shares its line, and
`{ E; return; }` when the `return` is the whole body of an `if`, `else`, loop or label (without the
braces the `return;` would run unconditionally). Offered only when the statement is exactly
`return void E;` (no comment between tokens, closing semicolon present, no parentheses around the
`void`) and `E` starts with an identifier character other than `function`, `class`, `let` or
`async`, so the new statement cannot join the previous line under automatic semicolon insertion or
parse as a declaration or a block.

The fix always writes `return;`. In a function typed to return a value,
`nexus/consistency-require-matching-return-type` then rewrites that bare return to
`return undefined;`, so the two converge in one run.

## Divergence from ESLint

None to record in `internal/differential/acknowledged.go`: the gate has no rule of this name.
`@typescript-eslint/no-confusing-void-expression` also reports these ten lines; that overlap is two
rules agreeing, not a divergence.

## How the measurement was taken

As for `consistency_no_for_in.md`: real binary on the working tree with a re-anchored scratch config
(3,745 files, 0 findings), and HEAD by parsing a `git archive` (10 findings, all ten carrying a fix).
