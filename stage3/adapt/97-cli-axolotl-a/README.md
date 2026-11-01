# 97a: program.ts tests a paths substitution where it narrows

Axolotl (#axolotl), task #qg1ry2t. This adaptation is JavaScript-changing in text only: one condition's text changes, its behavior doesn't. It changes no public API. It edits only the substitution loop in `verifyCompilerOptions` in src/compiler/program.ts.

lane: run as part of the Axolotl wave Correctness cuts it into; Loom's run of that wave is the lane's proof.

## The cause

`verifyCompilerOptions` checks each substitution of each `paths` pattern:

- **Read:** `const subst = options.paths[key]![i];` is `string | undefined`.
- **Stored:** `const typeOfSubst = typeof subst;` keeps the result of `typeof` in a local.
- **Tested:** `if (typeOfSubst === "string")` tests the local, not `subst`.

TypeScript narrows `subst` through `typeof subst === "string"` written in the condition, never through a stored `typeof` result. So inside the then branch `subst` is still `string | undefined`, and three calls that take a `string` reject it. Those are the three TS2345 diagnostics at 4157:64 (`hasZeroOrOneAsteriskCharacter(subst)`), 4160:69 (`pathIsRelative(subst)`) and 4160:95 (`pathIsAbsolute(subst)`).

## The repair

- The condition `typeOfSubst === "string"` becomes `typeof subst === "string"`.
- `const typeOfSubst = typeof subst;` stays: the else branch reports it as the `got_2` argument of `Substitution_0_for_pattern_1_has_incorrect_type_expected_string_got_2`.

No cast, no assertion. This is the repair Correctness and Adamic ruled.

## Why it's sound

- `typeOfSubst` is a `const` set to `typeof subst` on the line before the test, and `subst` is a `const` too. Neither can change between the two lines, so `typeof subst` in the condition is the same string `typeOfSubst` holds, and the test takes the same branch for every entry.
- `typeof` has no side effects and `subst` is a local, so evaluating it a second time observes nothing new.
- The else branch is untouched and still reads `typeOfSubst`, so its message is unchanged.

The emitted JavaScript differs in the text of that one condition and nowhere else.

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3 and finds the site by its AST: the one `for` loop in the file whose body declares `const subst`, which must sit in `verifyCompilerOptions`. That body must be exactly `const subst = options.paths[key]![i]`, `const typeOfSubst = typeof subst` and an `if` with an else. The condition must be the binary `typeOfSubst === "string"` (before) or `typeof subst === "string"` (after), spelled exactly so. The else branch must still read `typeOfSubst`, and every `subst` and `typeOfSubst` reference must agree with the condition (7 and 3 before, 8 and 2 after), or the file is refused as partly adapted. It plans the edit, re-reads its own result as the adapted shape, checks that only those bytes moved, and then writes once.
- **Refusals**, each probed on a scratch copy, each exit 1 with no write: `typeOfSubst !== "string"`, `"string" === typeOfSubst`, single quotes, `typeof subst === "string" && true`, the declaration removed with the condition adapted, the else branch no longer reading `typeOfSubst`, a second substitution loop, a stray `typeOfSubst` test, a stray adapted test, `options.paths[key][i]` without its `!`, and a fourth statement in the loop.
- **Diagnostics:** `adamic types` on the adapted tree's program.ts, at main 7b9c0fe7, on Golem:
  - whole program, 144 before and 141 after
  - program.ts own-file, 11 before and 8 after
  - The three removed are exactly 4157:64, 4160:69 and 4160:95, with no new diagnostic anywhere (the two sorted diagnostic lists differ in those three lines alone).
- **Source:** the adapter's only change to program.ts is line 4156, `if (typeOfSubst === "string") {` to `if (typeof subst === "string") {`.
- **JavaScript:** `ts.transpileModule` of program.ts before and after (ES2020, CommonJS) gives 240,346 and 240,347 bytes. The only difference is that condition, at output line 3073, one byte longer.
- **Idempotence:** a second run edits nothing and leaves the file's SHA-256 unchanged. An input already holding the adapted condition is accepted with no write.
- **Witnesses:** `node witnesses.cjs <program.ts before> <program.ts after>` lifts the substitution loop out of each file, transpiles it, and runs it with stock TypeScript's own `hasZeroOrOneAsteriskCharacter`, `pathIsRelative`, `pathIsAbsolute`, `Diagnostics` and `formatStringFromArgs`, with `baseUrl` unset. Before and after report the same diagnostics, by message and argument:

  | Entry | Diagnostics, before and after |
  | --- | --- |
  | `"*/*"` | 5062 "Substitution '\*/\*' in pattern '@app/\*' can have at most one '\*' character." args `*/*`, `@app/*`; 5090 "Non-relative paths are not allowed when 'baseUrl' is not set. Did you forget a leading './'?" no args |
  | `42` | 5064 "Substitution '42' for pattern '@app/\*' has incorrect type, expected 'string', got 'number'." args `42`, `@app/*`, `number` |
  | a hole, `[, ]` | 5064 with args `undefined`, `@app/*`, `"undefined"`; formatting it throws `Debug Failure.` |

  The hole's throw is upstream's, the same before and after: `formatStringFromArgs` refuses an undefined argument. JSON has no holes, and a `null` entry's `typeof` is `object`, so this case is reachable only through the API.
- **Mutant:** a copy of the adapter that writes `typeof subst !== "string"` is refused by its own re-read before writing ("not an inventoried shape"). With that re-read removed, it writes the wrong condition, and the witnesses fail: `"*/*"` reports 5064 "got 'string'" instead of 5062 and 5090, `42` throws in `pathIsAbsolute`, and the hole throws in `hasZeroOrOneAsteriskCharacter`.

Not yet run: the sharded lane oracle. It runs in Loom as part of the wave, per Correctness' rule on #axochk.
