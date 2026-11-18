# checker.ts's error output annotations name ErrorOutputContainer

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 6 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C08 (G05, 6 sites)

`checkTypeAssignableTo` and four `resultObj` annotations erase ErrorOutputContainer's `skipLogging`. Their values are forwarded containers or a fresh `{}`, and the real containers at checker.ts:36175 and :45952 write `skipLogging: true`. Every named write or initializer is Boolean-compatible, so the existing owned annotations name ErrorOutputContainer and keep the flag, rather than claiming it absent.

- `src/compiler/checker.ts`, function `checkTypeAssignableTo`: `errorOutputObject?: { errors?: Diagnostic[] | undefined; }` becomes `errorOutputObject?: ErrorOutputContainer` (1 time).
- `src/compiler/checker.ts`, function `elaborateDidYouMeanToCallOrConstruct`: `const resultObj: { errors?: Diagnostic[] | undefined; } = errorOutputContainer || {};` becomes `const resultObj: ErrorOutputContainer = errorOutputContainer || {};` (1 time).
- `src/compiler/checker.ts`, function `elaborateArrowFunction`: `const resultObj: { errors?: Diagnostic[] | undefined; } = errorOutputContainer || {};` becomes `const resultObj: ErrorOutputContainer = errorOutputContainer || {};` (1 time).
- `src/compiler/checker.ts`, function `elaborateElementwise`: `const resultObj: { errors?: Diagnostic[] | undefined; } = errorOutputContainer || {};` becomes `const resultObj: ErrorOutputContainer = errorOutputContainer || {};` (1 time).
- `src/compiler/checker.ts`, function `elaborateIterableOrArrayLikeTargetElementwise`: `const resultObj: { errors?: Diagnostic[] | undefined; } = errorOutputContainer || {};` becomes `const resultObj: ErrorOutputContainer = errorOutputContainer || {};` (1 time).

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.
