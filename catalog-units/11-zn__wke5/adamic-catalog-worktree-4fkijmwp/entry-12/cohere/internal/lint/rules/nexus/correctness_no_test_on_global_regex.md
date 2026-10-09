# `nexus/correctness-no-test-on-global-regex`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` once the one ahra site is fixed.** 1 finding, 1 true |
| Findings | **ahra 1** (measured 2026-10-03) |
| Measured precision | 1 of 1 true: a module-level `g` regex tested from a function, patched by hand with `lastIndex = 0` |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve the receiver to its declaration and `test` to RegExp's; without a checker the rule registers nothing |

## What it checks

`.test()` (the default library's `RegExp.test`) on a regex with the `g` flag held by a `const` or a
`readonly` property of the same file, where the call can run more than once against one regex: a function
boundary, or a loop that repeats the call but not the regex's creation, lies between them. `.test()` never
needs `g`; with it, a true leaves `lastIndex` past the match and the next call searches from there.

Exempt: `.exec()` (never reported; the `for(let m = re.exec(s); m; m = re.exec(s))` loop needs `g`),
`.test()` as a loop condition (`while(re.test(s)) count++`), a regex whose `lastIndex` is set to anything
but the literal `0` somewhere in the file (searching from a position), and a sticky regex without `g`.

## Where it came from

`modules/os/wisdom/AhraOsWisdom.ts:2437` in ahra: `compLeakDollarPattern` is declared at module scope with
`g` for the `replace` in `maskCompTokens`, and `textCarriesCompCommitment` calls `.test(text)` on the same
regex, with `compLeakDollarPattern.lastIndex = 0` after a true and a comment explaining why. Found by the
JavaScript-catalog pass of the new-rules sweep (`#tevhg3f`, after sonarjs `stateful-regex`), built in
`#j03vwm6`.

## Reconciling with the research count

Research counted **8 raw, 1 non-loop**: 7 were the `.exec()` loop in `MediumConversationSource.ts`, which
this rule does not report by construction, and the 1 is `AhraOsWisdom.ts:2437`. cohere counts **1**, the
same site. The task believed it fixed; it is still live at HEAD on 2026-10-03. A text probe over every
`.ts`/`.tsx` in ahra for `<name>.test(` on a binding declared with a `g` regex found that one site and no
other.

## A judgment, stated

The reported site is correct code today: the hand reset makes it work. It is reported because the reset
is the patch for a flag `.test()` does not need, and a later edit that adds an early return between the
`.test()` and the reset brings the bug back. Likewise a find-first loop that returns on the first true,
or a function called once per process, is reported though it never tests again after a true: proving
either is a whole-program question, and the fix (a regex without `g` for the `.test()`) is correct and
free at every one of them. If the Square wants only sites that can answer wrong today, the narrowing is
to exempt a binding with any `lastIndex = 0` write, and the one ahra finding goes away with it.

## Verification

- Fixtures from the real site: the comp-leak guard as commit `6a4b7896` added it, with the hand reset
  (fires) and without it (fires), and fixed with a separate non-global regex for the `.test()` (silent).
  8 firing shapes, 14 silent shapes, including `MediumConversationSource`'s `.exec()` loop on a `gi` regex.
- Mutation check, all killed: no repeat check (a regex made in the same run, inside the loop, or tested
  in a `for...of`'s iterated expression), no loop-condition exemption (`while` and `do` conditions), no
  positioning exemption (`lastIndex = position`), the zero reset exempting (the AhraOsWisdom fixtures),
  sticky-only regexes reported, a `let` accepted, a `for...of`'s iterated expression counted as repeating.
