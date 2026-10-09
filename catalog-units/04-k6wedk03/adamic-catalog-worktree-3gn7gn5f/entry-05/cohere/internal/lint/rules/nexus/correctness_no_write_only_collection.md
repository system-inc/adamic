# `nexus/correctness-no-write-only-collection`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` once the two ahra sites are fixed.** 2 findings, 2 true |
| Findings | **ahra 2** (measured 2026-10-03) |
| Measured precision | 2 of 2 true: each map is filled with `.set` and never read |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to find every reference by symbol and to tie `Map`, `Set` and each mutating method to the default library; without a checker the rule registers nothing |

## What it checks

A `const` declared inside a function, initialized to an array literal or the default library's
`new Map(...)` / `new Set(...)`, whose every reference is a write: a dropped call of a mutating method
(`push`, `unshift`, `pop`, `shift`, `splice`; `set`, `delete`, `clear`; `add`, `delete`, `clear`, each
resolved to the default library's declaration) or `name[key] = value`. At least one write is required; a
collection never referenced is `no-unused-vars`'s.

## Where it came from

`modules/os/database/AhraOsProfileCommitsStore.ts:728` and `:729` in ahra: `bmRanks` and `vecRanks` in
`searchProfileCommitsHybrid` are filled with `.set(row.rowid, rank)` during rank fusion and never read; the
ranks the result carries come from `hitData`. Found by the JavaScript-catalog pass of the new-rules sweep
(`#tevhg3f`, after sonarjs `no-unused-collection`), built in `#j03vwm6`.

## Reconciling with the research count

Research counted **2** (`AhraOsProfileCommitsStore.ts:728-729`). cohere counts **2**, the same two. The
task believed them fixed; both are still live at HEAD on 2026-10-03.

## What it declines

A module-level collection (it could be exported through `export { name }`, an alias this rule would have
to resolve), a `let`, a write in an arrow's expression body (`rows.forEach((row) => list.push(row))`
hands the result back to the caller), a mutating call whose result is used, and a compound element write
(`counts[0] += 1` reads). Each is a missed finding.

## Verification

- Fixtures from the real site: the rank fusion as it stands (`bmRanks` and `vecRanks` fire; `scores`,
  `hitData` and `ranked` stay silent) and fixed (the two maps deleted, silent). 7 firing shapes including
  shadowing (only the outer, written-only binding fires), 17 silent shapes.
- Mutation check, all killed: a used result counted as a write (a mutating call whose result is used, the
  arrow body), any method counted as a writer (a typed recorder whose `push` is its own), module level
  accepted, a compound element write counted, references matched by name rather than symbol (shadowing),
  any `Map` constructor accepted (a `Map` subclass declared under the library's name that registers itself).
