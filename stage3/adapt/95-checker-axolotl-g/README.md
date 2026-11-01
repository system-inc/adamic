# 95g: no-symbol fix blocked by a service reader

Status: reference choice resolved; stopped before building at an unsafe reader.
The candidate is a JavaScript-changing fork fix, not a soundness assertion.
No adapter is shipped. This report-only branch is not mergeable as an adaptation.
Base: 95a, ef4d26a4a38b7890317e1ba71f8ad928fd8588f6.

lane: run as part of a Correctness wave

Correctness resolved the reference choice: on an empty filtered declaration
list, return undefined before creating an index symbol. This follows both
reference implementations:

- Official microsoft/typescript-go commit
  `89d5d5b2849a0db0957065889ca58536fa6d2e4a`,
  [internal/checker/checker.go](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/checker.go#L32209),
  getApplicableIndexSymbol creates the symbol only if len(declarations) != 0.
  With an empty list it leaves info.indexSymbol nil and returns nil. It does
  not return a parentless symbol.
- TypeScript main commit `6161772c1370bdd4a82f808e6823bd755eb1870d` now has
  its checker in Go; src/compiler/checker.ts is absent. Its equivalent
  [tsc/internal/checker/checker.go](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/checker.go#L32895)
  has the same empty-list behavior. Its current implementation already uses
  the newly ruled no-symbol result.

The supporting getObjectLiteralIndexInfo function creates the broad computed
index with declaration nil, and getIndexInfosOfIndexSymbol propagates it.
These source checks, including the complete functions, were read before
building. No typescript-go execution is claimed. The distinction is visible
through the symbol API: no symbol and a symbol whose parent is absent are
different query results.

## Blocking getSymbolAtLocation reader

The new caller audit found a reader that does not tolerate undefined:
`src/services/findAllReferences.ts:580-581`,
`getImplementationReferenceEntries`, in its super-property arm:

```ts
const symbol = checker.getSymbolAtLocation(node)!;
return symbol.valueDeclaration && [nodeEntry(symbol.valueDeclaration)];
```

It is reachable through the public language service method
getImplementationAtPosition, services.ts:2416, then
FindAllReferences.getImplementationsAtPosition. An identifier naming the right
side of super.xyz takes this exact arm. The assertion erases on Node; there
is no absence guard before valueDeclaration.

A real language-service witness, with zero syntactic or semantic diagnostics,
uses a broad computed static index and a derived class:

```js
/** @type {string} */
let key = "xyz";
class Computed { static [key] = 1; }
class Derived extends Computed { static read() { return super.xyz; } }
```

Before the candidate guard, getImplementationAtPosition on xyz returns
undefined without throwing. With `if (!declarationList.length) return undefined;`
evaluated in memory at the proposed early-return site, that same service query
throws TypeError reading valueDeclaration in getImplementationReferenceEntries.
The query on the super keyword itself is unchanged. No compiler source or
built file on disk was modified, and no G compiler was built for this audit.

The original unit's stop rule says that if a reader does not tolerate the
missing result, stop and report that reader instead of building. The newly
requested check that every reachable caller already takes undefined is false.
G therefore stops before implementation, counts, oracle or mutant claims.
The full caller inventory is retained, but a complete safe-caller proof is not
claimed after finding this concrete blocker. A follow-up must authorize and
audit the service consumer change before the Go-style fork fix can ship.

There is also a scope distinction: a direct Computed.xyz symbol query can
currently return an __index symbol with zero declarations and parent Computed,
without diagnostics or a throw. The unconditional Go-style guard makes that
result undefined too. The report does not claim that all previously successful
queries remain unchanged. This is distinct from, and weaker than, the service
regression above, which demonstrates an actual new exception.

## Baseline witness

The existing 95a built compiler was queried, without rebuilding or modifying it.
A JavaScript class Computed has a broad computed static string key. A separate
empty class expression is cast through JSDoc to typeof Computed and exported
via module.exports; an extra export member removes the original type symbol.
A TypeScript consumer imports it and queries v.xyz with getSymbolAtLocation.
The fallback has one index info but no index declaration. At both noImplicitAny
settings the program has zero diagnostics and the query throws TypeError:
Cannot read properties of undefined (reading 'parent'). Both complete stacks
and all diagnostic output are retained with the witness.

Evidence is on typescript/axolotl-checker-g-evidence under
stage3/evidence/95-checker-axolotl-g/. Pinned evidence commit: `e139d2ebc8d75e48307c0e9e9f62a828a4705e87`.
[Reference excerpts, baseline witness and full stacks](https://github.com/system-inc/adamic/tree/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/).
The drafted issue in stage3/upstream/LEDGER.md documents the 6.0.3 crash, the
resolved Go-style behavior and the unsafe service consumer. Nothing has been filed. F and G independently
append to 95a's ledger end; whoever merges second must resolve that append
conflict and preserve both entries.

## Candidate evidence, not a built fix

An in-memory compile hook inserts the early empty-list return in the existing
95a JavaScript bundle, without building it. The original two getSymbolAtLocation
crashes become undefined results, with zero diagnostics preserved. The same
hook exposes the unsafe service reader above. Complete before/after service
outputs and stacks, the exact hook and witness, the original query results,
caller inventory and blocking source excerpt are on the evidence branch.

A suggested reader guard is retained as an unbuilt text diff for review.
It has not been applied. Fixing that reader alone is not presented as a completed
audit of every other caller.

Reader regression artifacts:

- [Exact language-service fixture and runner](https://github.com/system-inc/adamic/blob/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/super-witness.cjs).
- [Before: zero diagnostics, undefined result](https://github.com/system-inc/adamic/blob/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/super-before.json).
- [Candidate: zero diagnostics, TypeError stack](https://github.com/system-inc/adamic/blob/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/super-with-guard.json).
- [Blocking reader source](https://github.com/system-inc/adamic/blob/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/blocking-reader.ts.txt).
- [Unbuilt proposed consumer guard](https://github.com/system-inc/adamic/blob/e139d2ebc8d75e48307c0e9e9f62a828a4705e87/stage3/evidence/95-checker-axolotl-g/proposed-reader-guard.diff.txt).

The evidence update does not represent a built 95g fix. No own-file or
whole-program diagnostic delta, oracle verdict or mutant result is claimed.
