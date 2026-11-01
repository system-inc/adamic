# 95g: stopped at the reference gate

Status: not built, no adapter shipped, not mergeable as an adaptation.
Base: 95a, ef4d26a4a38b7890317e1ba71f8ad928fd8588f6.

lane: run as part of a Correctness wave

The requested behavior retains a filtered index symbol and leaves its parent
unset when its declaration list is empty. The required reference check found
an observable disagreement before any G implementation or build:

- Official microsoft/typescript-go commit
  `89d5d5b2849a0db0957065889ca58536fa6d2e4a`,
  [internal/checker/checker.go](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/checker.go#L32209),
  getApplicableIndexSymbol creates the symbol only if len(declarations) != 0.
  With an empty list it leaves info.indexSymbol nil and returns nil. It does
  not return a parentless symbol.
- TypeScript main commit `6161772c1370bdd4a82f808e6823bd755eb1870d` now has
  its checker in Go; src/compiler/checker.ts is absent. Its equivalent
  [tsc/internal/checker/checker.go](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/checker.go#L32895)
  has the same empty-list behavior. The current implementation has therefore
  already changed from 6.0.3, but does not implement the requested result.

The supporting getObjectLiteralIndexInfo function creates the broad computed
index with declaration nil, and getIndexInfosOfIndexSymbol propagates it.
These source checks, including the complete functions, were read before
building. No typescript-go execution is claimed. The distinction is visible
through the symbol API: no symbol and a symbol whose parent is absent are
different query results.

The instruction says "If either disagrees with the behavior above, stop and
report before building." G stops here. No G source patch, build, after witness,
diagnostic reduction, full oracle or mutant is claimed. No claim is made that
every downstream reader tolerates a retained parentless symbol; that audit is
not completed after the earlier reference gate failed. This README-only branch
is a stop report, not an executable adaptation and must not be merged into apply.
The adapter directory contains only this reader-facing report.

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
stage3/evidence/95-checker-axolotl-g/. Pinned evidence commit: `fa5ce4ad30f5fa06000075f116b1d66838d46290`.
[Reference excerpts, baseline witness and full stacks](https://github.com/system-inc/adamic/tree/fa5ce4ad30f5fa06000075f116b1d66838d46290/stage3/evidence/95-checker-axolotl-g/).
The drafted issue in stage3/upstream/LEDGER.md documents the 6.0.3 crash and the
unresolved behavior choice. Nothing has been filed. F and G independently
append to 95a's ledger end; whoever merges second must resolve that append
conflict and preserve both entries.
