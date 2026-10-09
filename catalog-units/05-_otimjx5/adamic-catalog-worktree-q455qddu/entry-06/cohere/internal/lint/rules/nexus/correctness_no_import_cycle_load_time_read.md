# `nexus/correctness-no-import-cycle-load-time-read`

| | |
|---|---|
| **Recommendation** | **Yes, but registered `off` until the sites are fixed.** 8 findings on ahra, 0 false, all from two cycles in Nexus, each fixed by moving one helper to a file of its own |
| Findings | **ahra 8** (measured 2026-10-03) |
| Measured precision | 8 of 8 true: each is a binding read while its module loads, declared in a module of the same runtime import cycle, and each throws `Cannot access ... before initialization` when the cycle is entered at the declaring file |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, for symbol identity and for TypeScript's elision rule; reads other files (`ReadsOtherFiles`), so the findings cache never replays it |
| Cost on ahra | the program's graph is built once per run: 24 to 124ms wall across five runs (20ms resolving every import, the rest asking the checker about the 70 files inside cycles). `--timing` sums it to 0.4 to 2.7s, because every other worker's wait on the build's lock is counted as its setup |

## What it checks

A binding read while the reading module loads (top-level code, `export default <expression>`, a class's
`extends`, class decorators, `static` initializers and blocks, enum initializers, `typeof`, a namespace
member `ns.X`), when the binding, followed through every re-export to its declaration, is declared in
another module of the same runtime import cycle and is created when that module runs: `const`, `let`,
`var`, a `class`, a non-`const` `enum`, `export default <expression>`. Whether it throws today depends on
which module of the cycle the program enters first, which one new import anywhere can change. The doc
comment carries the precise version.

## How cross-file reading works here

cohere has no program-level hook: a rule runs per file. Rules that need the whole program
(`consistency-require-constant-casing`'s importer index, `boundary-no-project-theme-value`'s theme map)
build a derived index once per program, under a package lock, keyed by `ProgramIdentity`, and declare
`ReadsOtherFiles`. This rule does the same. The graph is built in two passes: a syntactic one over every
import and re-export not written `type` (cheap, a resolution per specifier), and a checker pass that keeps
only the edges TypeScript would emit, run only on files inside a cycle of the first pass. On ahra both
passes found the same 20 cycles over 70 files, so the second pass dropped nothing today; its fixtures are
what show it is needed.

## Which imports execute

Kept, because they run: `import './x'`, `export * from`, `export * as ns from`, an import whose binding is
read as a value and lands on a value (TypeScript's elision rule, which tsc, SWC and esbuild share), and an
`export { a } from` whose specifier lands on a value. Dropped, because each doubt can only invent a cycle:
`import type`, `export type`, inline `type` specifiers, an import used only in types, a `const enum`, a
dynamic `import()`, `require()`, anything resolving outside the project, and any import of a `'use server'`
module (a Next.js server action is a reference from client code, so a cycle through one may run in neither
bundle).

## Where it came from

`libraries/structure/libraries/nexus/source/protocols/base/errors/BaseErrorIdentifiers.ts:62` spreads
`...RateLimiterModuleErrors` at module top level. `RateLimiterModuleErrors.ts` imports
`BaseErrorIdentifierKeys` back for `isRateLimitErrorData`, and `isBaseErrorData` from `BaseError.ts`,
which imports the identifiers file too. Found by the JS catalogue pass of the new-rules sweep (`#tevhg3f`,
after Biome's `noImportCycles`, refined), built as `#c8armec`.

## Reconciling with the research count

Research counted **1** dangerous cycle (and 19 cycles in all). cohere reports **8** reads in **2** cycles,
out of 20 cycles in all. The gap is two things, neither a false positive:

- **The research counted cycles, the rule counts reads.** Its probe emitted one hit per cycle with reads.
- **The research probe skipped `extends`.** It stopped at `ts.isTypeNode`, which is true for the
  `ExpressionWithTypeArguments` a class heritage clause holds, so `class ArraySchema extends BaseSchema`
  never counted. That is the textbook form of this bug: `BaseSchema.ts` imports
  `normalizeValidatorMessageOverrides` from `Schema.ts`, which imports every concrete schema, each of which
  extends `BaseSchema` at load. Entered at `BaseSchema.ts`, all seven class definitions throw. The rule
  reports them: `ArraySchema.ts:11`, `BooleanSchema.ts:8`, `DateSchema.ts:8`, `FileSchema.ts:8`,
  `NumberSchema.ts:8`, `ObjectSchema.ts:11`, `StringSchema.ts:25`.
- 20 cycles against 19: the tree has moved since the research run; every one of the 20 was read.

## The findings, read one by one

| Site | Reading |
|---|---|
| `nexus/source/protocols/base/errors/BaseErrorIdentifiers.ts:62` | **Bug**: `...RateLimiterModuleErrors`, cycle `BaseErrorIdentifiers → RateLimiterModuleErrors → BaseErrorIdentifiers` (and through `BaseError.ts`) |
| `nexus/source/validation/schema/{Array,Boolean,Date,File,Number,Object,String}Schema.ts` | **Bug** x7: `extends BaseSchema`, cycle `XSchema → BaseSchema → Schema → XSchema` |

**Recall spot check.** Every load-time read in all 20 cycles that matched an import from the same cycle
and was declined was logged: one, `Tabs.tsx` reading `TabItem`, a `function` declaration, which is hoisted
and initialized before any module in the cycle runs. Correctly silent.

## What it declines, on purpose

- **A cycle with no read at load.** 18 of the 20 cycles are components rendering each other, or modules
  calling each other inside functions. Harmless, and a plain cycle rule would report all of them.
- **A hoisted `function`,** even though calling one at load can still reach an uninitialized binding inside
  its body. Following calls is not done; a missed finding.
- **Reads deferred one step:** an IIFE at top level, a method's computed name or decorators, `export =`
  through `import x = require()`. Missed findings, never false ones.

## Verification

- Fixtures both ways from the real site: the three-file `BaseErrorIdentifiers` cycle reports the one
  spread and not the two outside the cycle, the other two members stay silent, and the fix (the guard moved
  to `RateLimitErrorData.ts`) is silent. 26 firing shapes and 29 silent ones, including every runtime and
  erased back edge, a barrel, a type-only re-export, `'use server'`, and three JSX component cases.
- Mutation check, each mutant through `go test -overlay`, all killed: the load-time boundary removed
  (function bodies, methods, shadowing), elision skipped (a value import used only in types, a shadowing
  parameter, an inline type specifier, a type re-export), functions not hoisted, `extends` not read,
  instance fields read at load, the type-only alias check (a type-only re-export), the ambient check, the
  `const enum` declaration check, the `'use server'` boundary, and the JSX closing-tag guard.
