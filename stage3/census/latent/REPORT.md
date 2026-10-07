Built: a scratch-only census that measures eligible top-level functions and every top-level statement across six configurations.
Base: ef3d907ecdc4c771b016f7d9c52372def057a340; exact scratch heads and source/binary/overlay hashes are in REPORT.json.
Commands/results: six builds and six corpus measurements exit 0; all census counts are measured on a checker-rejected program.
Mutants: extra NotYet on probes and real tsc, body-scope, misattribution, output guards, and report artifacts caught.
Limits: first lowering error per unit, diagnosed-body skips, generic/isolation limits, and no native execution claim.

# Scope and method

**Every census number in this report and its JSON is measured on a checker-rejected program.**
This includes checker counts, unit counts, reason/file totals, skipped counts, and
feature deltas. Branch/source SHAs and setup timings are provenance. These
observations cannot establish which programs compile or preserve JavaScript semantics.

TypeScript 6.0.3 is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`.
The 78 compiler roots are the 77 original sources plus
`diagnosticInformationMap.generated.ts`. `apply.sh` ran with adaptations 10 and 20
merged: type imports changed 72 files / 3,719 lines; optional declarations changed
26 files / 406 lines; combined, 73 files / 4,125 lines. All configurations use the
same adapted bytes. `data/apply.log.gz`, `data/patch-set.md`, and the SHA-256 manifest
preserve this evidence. The two JSON input files are not source roots.

The project is checked once per configuration with all roots together. The overlay
preserves every checker diagnostic but exposes the rejected program only through
`LatentLoad`. It matches each top-level function body's byte interval against raw
checker diagnostic spans in that file. A diagnosed body is skipped and counted;
a signature-only diagnostic leaves its body eligible. Functions on the same line
remain distinct. The raw spans and unit eligibility ledger permit an independent recount.

For each file, the refusal visitor continues past findings and skips diagnosed
top-level functions. Every eligible function and every top-level statement gets a
fresh lowering state. A returned error or recoverable panic is recorded and the next
unit runs. Project declarations, globals, class static storage, and supported enum
values are registered without lowering sibling bodies. Sibling function signatures
are prepared on use. Checker-diagnosed dependency bodies encountered by generic/class
lowering become separate `SkippedDependency` boundaries, excluded from actual
NotYet/Refused counts. Registration failures are rediscovered at actual reads.

Counts deduplicate `(kind, location, reason, exact diagnostic text)` across attempts.
A shared dependency failure is attributed to its actual diagnostic file, and raw
events retain each owning unit and refusal/lowering phase. REPORT.json includes all
unique finding texts/locations, per-file reasons/deltas, and source/binary/overlay
hashes. `data/<configuration>.jsonl.gz` preserves all raw events, checker spans and
units. `outside_roots` accounts for unlocated ordinary errors as well as findings
outside the compiler roots. All NotYet/Refused sites have structured locations.

The refusal scan records every visited refusal site. Lowering still returns its
first error within each unit. These are observed latent findings, not an exhaustive
list of every potential error in every function.

# Measurement only

Only `stage3/census/latent/` is committed on the delivery branch. Measurement edits are
scratch Go overlays, never production edits. Ordinary `Load` and `LoadOverlay`
are disabled in the census binary; `lower.Lower` always returns nil IR and an explicit
measurement error. The driver imports no emitter/backend. Every corpus run enables
`LATENT_ASSERT_NO_OUTPUT`, checking both disabled APIs. The ordinary driver fails
unless overlaid. No scratch branches are pushed, and no native output is produced.

# Configurations and conflict resolution

Baseline is main plus the stage3 pipeline, original census and adaptations 10/20.
Each individual configuration adds the named feature alone to that same baseline.
The cumulative run adds taste, flags, namespaces and nested in that order. Feature
branches include their ancestors; comparisons concern their actual merged heads.

| Configuration | Feature commit(s) | Resolved scratch head | Never-pushed branch |
|---|---|---|---|
| main | baseline | fc3482f2605590e1aa8bc49b7b4fd356f3c25717 | scratch/latent-compare-main |
| taste | codex/taste-not-soundness: aa896b5d5ccc82210184fd01b8fe4d0ce0730a50 | f6fca9973fb2236f7b1ec3c0da5c23ae143a194f | scratch/latent-compare-taste |
| flags | codex/flag-enums: f7d62772fa8e52fcfae754e047ceb66dba88b782 | b81f81e81aaaf99b2a33e63b71cf18a9aaceb5f5 | scratch/latent-compare-flags |
| namespaces | codex/namespaces-tsc: ce8a2acf14e420a9c82345236845a377cd4c7a50 | 19b64b3b2029ee4f069c39c7ed79f5b3e9ecbb8e | scratch/latent-compare-namespaces |
| nested | codex/nested-functions: b15216dabf65ffaa7152f6e64709b7b062ea01a9 | b17f58168a8af23001b8e30bb0eb082cbcfc251b | scratch/latent-compare-nested |
| cumulative | codex/taste-not-soundness: aa896b5d5ccc82210184fd01b8fe4d0ce0730a50<br>codex/flag-enums: f7d62772fa8e52fcfae754e047ceb66dba88b782<br>codex/namespaces-tsc: ce8a2acf14e420a9c82345236845a377cd4c7a50<br>codex/nested-functions: b15216dabf65ffaa7152f6e64709b7b062ea01a9 | 15eb079bc200cb1f3f8137fa762fc53bdbc45f22 | scratch/latent-rejected-cumulative |

Individual flags/namespaces conflicts combine main's accessor and nominal checks
with the feature enum/namespace rules. Cumulative resolutions also combine taste's
syntax support, flag safeguards, namespace state/traversal, and nested sibling-call
and capture rules. `resolve_scratch.py` contains exact observed conflict recipes,
accepts only `/tmp` trees on `scratch/latent-*` branches, and refuses unknown conflicts.
Merge/resolution logs are in data/. These are scratch integration decisions, not
compiler changes committed by this unit. Conflicting oracle count rows retain the
current scratch ledger with new feature rows appended; that ledger is unused by the
census and is not claimed as validated oracle evidence. All resolved compilers build.

Dependency pins: cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`,
typescript-go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
Scratch trees share the initialized checkout through a symlink. Builds use
`-buildvcs=false` because initial VCS stamping did not recognize the symlink as a
worktree submodule. This changes build metadata, not compiler semantics.

# Totals

Every number below is **measured on a checker-rejected program**. Dependency skips
are separate measurement boundaries. Bodyless function declarations are attempted
and can produce the lowerer's ordinary missing-body diagnostic.

| Configuration | Checker diagnostics | Units | Functions attempted | Bodies skipped | Statements attempted | NotYet | Refused | Dependency skips | Errors/panics | Raw events |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| main | 2165 | 4530 | 2499 | 256 | 1775 | 1938 | 1813 | 23 | 0/0 | 4096 |
| taste | 2165 | 4530 | 2499 | 256 | 1775 | 1864 | 1708 | 25 | 0/0 | 3965 |
| flags | 1985 | 4530 | 2500 | 255 | 1775 | 1177 | 2106 | 23 | 0/0 | 3646 |
| namespaces | 1985 | 4530 | 2500 | 255 | 1775 | 1188 | 2129 | 23 | 1/0 | 3660 |
| nested | 2165 | 4530 | 2499 | 256 | 1775 | 1943 | 1809 | 23 | 0/0 | 4097 |
| cumulative | 1985 | 4530 | 2500 | 255 | 1775 | 1032 | 1983 | 25 | 1/0 | 3409 |

The namespace and cumulative runs each record one ordinary error:
`lower: src/compiler/parser.ts:1472:9: the checker gave a declaration no symbol`.
Its error type exposes no structured location, so it is retained in the unlocated
`outside_roots` bucket and its attempting unit remains in the raw event. It is not
counted as NotYet or Refused. No configuration records a panic.

# Feature deltas

All deltas are **measured on a checker-rejected program**, relative to main.
Negative count deltas mean fewer observed sites. Exact removed/added identities
expose shifted first errors. Common-unit deltas restrict events to units eligible
in both configurations, separating changes in eligibility from changes in blockers.
These observations do not prove successful compilation or feature semantics.

| Configuration | NotYet delta | Refused delta | Removed sites | Added sites | Skipped-body delta | Newly eligible units | Common-unit NotYet delta | Common-unit Refused delta |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| taste | -74 | -105 | 366 | 189 | 0 | 0 | -74 | -105 |
| flags | -761 | 293 | 1017 | 549 | -1 | 1 | -762 | 293 |
| namespaces | -750 | 316 | 1037 | 604 | -1 | 1 | -751 | 316 |
| nested | 5 | -4 | 40 | 41 | 0 | 0 | 5 | -4 |
| cumulative | -906 | 170 | 1433 | 700 | -1 | 1 | -907 | 170 |

Largest reason decreases and increases, all **measured on a checker-rejected program**:

- taste: Refused: an ExportDeclaration (-77); NotYet: a PrefixUnaryExpression on a value (-75); Refused: a value as a condition (-50); NotYet: a BinaryExpression with a value and a value (-21); NotYet: a BinaryExpression with a value and a boolean (-16); Refused: export * (+75); NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class (+24); NotYet: reading SyntaxKind (+22); NotYet: a field of type boolean \| undefined (+4); NotYet: reading ScriptKind (+3).

- flags: NotYet: reading SyntaxKind (-481); Refused: enum (-162); NotYet: an EnumDeclaration (-154); NotYet: reading CharacterCodes (-35); NotYet: reading ModifierFlags (-22); Refused: arithmetic assigned back into an enum; the result need not be one of its members (+42); Refused: a cast the runtime can't check (+34); Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown (+30); NotYet: a PrefixUnaryExpression on a number (+21); Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BinaryExpression; its members are a closed union (+20).

- namespaces: NotYet: reading SyntaxKind (-481); Refused: enum (-162); NotYet: an EnumDeclaration (-154); NotYet: reading CharacterCodes (-35); NotYet: reading ModifierFlags (-22); Refused: arithmetic assigned back into an enum; the result need not be one of its members (+50); Refused: a cast the runtime can't check (+34); Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown (+30); NotYet: a PrefixUnaryExpression on a number (+21); Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BinaryExpression; its members are a closed union (+20).

- nested: NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions (-3); NotYet: a function inside a function (a closure) (-3); NotYet: a BinaryExpression as a statement (-2); NotYet: reading ModuleResolutionKind (-2); NotYet: a BinaryExpression with a number and a number (-1); NotYet: a function value with an optional parameter (+6); NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class (+3); NotYet: a function value returning union of differently held members (+3); NotYet: an optional chain longer than one step (+2); NotYet: a function without a body (+2).

- cumulative: NotYet: reading SyntaxKind (-481); Refused: enum (-162); NotYet: an EnumDeclaration (-154); Refused: an ExportDeclaration (-77); NotYet: a PrefixUnaryExpression on a value (-75); Refused: export * (+75); Refused: a cast the runtime can't check (+46); Refused: arithmetic assigned back into an enum; the result need not be one of its members (+42); NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class (+34); Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown (+33).

# Per file

Every cell is **measured on a checker-rejected program**, showing `NotYet / Refused`.
JSON also includes checker/skip counts, per-file reasons, and per-file deltas.
Zero own findings do not imply the rejected program compiles.

| File under src/compiler | Main N/R | Taste N/R | Flags N/R | Namespaces N/R | Nested N/R | Cumulative N/R |
|---|---:|---:|---:|---:|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 |
| _namespaces/ts.performance.ts | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 |
| _namespaces/ts.ts | 0 / 75 | 0 / 73 | 0 / 75 | 0 / 75 | 0 / 75 | 0 / 73 |
| binder.ts | 8 / 2 | 7 / 2 | 5 / 4 | 5 / 4 | 8 / 2 | 3 / 4 |
| builder.ts | 30 / 28 | 29 / 26 | 27 / 26 | 27 / 38 | 30 / 27 | 24 / 24 |
| builderPublic.ts | 8 / 0 | 8 / 0 | 6 / 0 | 6 / 0 | 8 / 0 | 6 / 0 |
| builderState.ts | 3 / 20 | 3 / 16 | 2 / 20 | 1 / 19 | 3 / 20 | 1 / 15 |
| builderStatePublic.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| checker.ts | 34 / 36 | 32 / 36 | 9 / 19 | 9 / 17 | 34 / 36 | 5 / 17 |
| commandLineParser.ts | 71 / 63 | 65 / 61 | 65 / 75 | 65 / 75 | 70 / 63 | 58 / 73 |
| core.ts | 171 / 71 | 171 / 67 | 167 / 70 | 167 / 70 | 171 / 70 | 167 / 66 |
| corePublic.ts | 1 / 2 | 1 / 2 | 0 / 1 | 0 / 1 | 1 / 2 | 0 / 1 |
| debug.ts | 2 / 59 | 2 / 59 | 3 / 62 | 14 / 67 | 2 / 59 | 15 / 61 |
| diagnosticInformationMap.generated.ts | 2 / 0 | 2 / 0 | 1 / 0 | 1 / 0 | 2 / 0 | 1 / 0 |
| emitter.ts | 18 / 13 | 16 / 10 | 13 / 12 | 13 / 12 | 17 / 14 | 11 / 9 |
| executeCommandLine.ts | 21 / 23 | 18 / 21 | 19 / 22 | 19 / 22 | 22 / 23 | 18 / 20 |
| expressionToTypeNode.ts | 2 / 1 | 2 / 1 | 2 / 1 | 2 / 1 | 2 / 1 | 2 / 1 |
| factory/baseNodeFactory.ts | 1 / 0 | 1 / 0 | 1 / 0 | 1 / 0 | 1 / 0 | 1 / 0 |
| factory/emitHelpers.ts | 5 / 1 | 5 / 1 | 3 / 0 | 3 / 0 | 5 / 1 | 3 / 0 |
| factory/emitNode.ts | 26 / 7 | 26 / 6 | 26 / 7 | 26 / 11 | 26 / 7 | 25 / 7 |
| factory/nodeChildren.ts | 5 / 1 | 5 / 1 | 2 / 2 | 4 / 2 | 5 / 1 | 4 / 2 |
| factory/nodeConverters.ts | 1 / 11 | 1 / 11 | 1 / 11 | 1 / 11 | 1 / 11 | 1 / 11 |
| factory/nodeFactory.ts | 16 / 5 | 16 / 3 | 9 / 9 | 10 / 9 | 16 / 5 | 8 / 7 |
| factory/nodeTests.ts | 227 / 227 | 227 / 227 | 0 / 227 | 0 / 227 | 227 / 227 | 0 / 227 |
| factory/parenthesizerRules.ts | 1 / 2 | 1 / 2 | 1 / 2 | 1 / 2 | 1 / 2 | 1 / 2 |
| factory/utilities.ts | 65 / 98 | 69 / 94 | 37 / 114 | 36 / 107 | 65 / 98 | 41 / 101 |
| factory/utilitiesPublic.ts | 3 / 3 | 3 / 2 | 1 / 3 | 1 / 3 | 3 / 3 | 1 / 2 |
| moduleNameResolver.ts | 70 / 42 | 64 / 32 | 57 / 56 | 57 / 63 | 73 / 41 | 50 / 45 |
| moduleSpecifiers.ts | 16 / 22 | 15 / 20 | 9 / 21 | 10 / 21 | 16 / 22 | 9 / 19 |
| parser.ts | 43 / 101 | 44 / 96 | 42 / 188 | 37 / 196 | 44 / 101 | 36 / 179 |
| path.ts | 37 / 4 | 34 / 0 | 33 / 4 | 33 / 4 | 37 / 4 | 29 / 0 |
| performance.ts | 8 / 1 | 6 / 0 | 8 / 1 | 8 / 1 | 8 / 1 | 6 / 0 |
| performanceCore.ts | 1 / 2 | 1 / 1 | 1 / 2 | 1 / 2 | 1 / 2 | 1 / 1 |
| program.ts | 24 / 49 | 23 / 44 | 21 / 50 | 22 / 50 | 23 / 49 | 19 / 45 |
| programDiagnostics.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| resolutionCache.ts | 10 / 3 | 10 / 3 | 10 / 3 | 10 / 3 | 10 / 3 | 10 / 3 |
| scanner.ts | 45 / 12 | 46 / 10 | 28 / 12 | 28 / 12 | 45 / 12 | 28 / 10 |
| semver.ts | 2 / 3 | 1 / 2 | 2 / 3 | 2 / 3 | 2 / 3 | 1 / 2 |
| sourcemap.ts | 8 / 20 | 8 / 18 | 6 / 22 | 7 / 22 | 8 / 20 | 7 / 20 |
| symbolWalker.ts | 1 / 7 | 1 / 7 | 1 / 7 | 1 / 7 | 1 / 7 | 1 / 7 |
| sys.ts | 20 / 31 | 19 / 28 | 15 / 28 | 15 / 28 | 22 / 31 | 16 / 25 |
| tracing.ts | 1 / 8 | 1 / 8 | 2 / 9 | 3 / 8 | 1 / 8 | 3 / 8 |
| transformer.ts | 5 / 4 | 5 / 3 | 2 / 4 | 2 / 4 | 5 / 4 | 2 / 3 |
| transformers/classFields.ts | 7 / 5 | 7 / 5 | 4 / 3 | 4 / 3 | 7 / 5 | 4 / 3 |
| transformers/classThis.ts | 4 / 2 | 3 / 2 | 4 / 2 | 4 / 2 | 4 / 2 | 3 / 2 |
| transformers/declarations/diagnostics.ts | 2 / 12 | 2 / 12 | 2 / 12 | 2 / 12 | 1 / 13 | 1 / 12 |
| transformers/declarations.ts | 9 / 3 | 9 / 3 | 3 / 8 | 3 / 8 | 9 / 3 | 2 / 8 |
| transformers/destructuring.ts | 11 / 16 | 9 / 16 | 10 / 17 | 10 / 17 | 9 / 16 | 6 / 17 |
| transformers/es2015.ts | 6 / 6 | 6 / 6 | 0 / 0 | 0 / 0 | 6 / 6 | 0 / 0 |
| transformers/es2016.ts | 0 / 10 | 0 / 10 | 0 / 11 | 0 / 11 | 1 / 10 | 1 / 11 |
| transformers/es2017.ts | 3 / 2 | 3 / 2 | 1 / 1 | 1 / 1 | 3 / 2 | 1 / 1 |
| transformers/es2018.ts | 2 / 2 | 2 / 2 | 0 / 0 | 0 / 0 | 2 / 2 | 0 / 0 |
| transformers/es2019.ts | 1 / 0 | 1 / 0 | 1 / 1 | 1 / 1 | 1 / 0 | 1 / 1 |
| transformers/es2020.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/es2021.ts | 0 / 1 | 0 / 1 | 0 / 1 | 0 / 1 | 1 / 0 | 1 / 0 |
| transformers/esDecorators.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/esnext.ts | 5 / 2 | 5 / 2 | 0 / 1 | 0 / 1 | 5 / 2 | 0 / 1 |
| transformers/generators.ts | 6 / 6 | 6 / 6 | 1 / 1 | 1 / 1 | 6 / 6 | 1 / 1 |
| transformers/jsx.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/legacyDecorators.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/module/esnextAnd2015.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/module/impliedNodeFormatDependent.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 1 / 0 | 1 / 0 |
| transformers/module/module.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/module/system.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/namedEvaluation.ts | 16 / 2 | 14 / 2 | 16 / 2 | 16 / 2 | 16 / 2 | 14 / 2 |
| transformers/taggedTemplate.ts | 2 / 4 | 2 / 4 | 1 / 3 | 1 / 3 | 2 / 4 | 1 / 3 |
| transformers/ts.ts | 2 / 2 | 2 / 2 | 0 / 0 | 0 / 0 | 2 / 2 | 0 / 0 |
| transformers/typeSerializer.ts | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 | 0 / 0 |
| transformers/utilities.ts | 29 / 20 | 23 / 19 | 24 / 48 | 24 / 48 | 28 / 20 | 17 / 47 |
| tsbuild.ts | 3 / 2 | 3 / 2 | 2 / 1 | 1 / 0 | 3 / 2 | 1 / 0 |
| tsbuildPublic.ts | 52 / 85 | 52 / 83 | 48 / 83 | 48 / 83 | 52 / 85 | 48 / 81 |
| types.ts | 76 / 80 | 76 / 80 | 0 / 10 | 0 / 7 | 76 / 80 | 0 / 10 |
| utilities.ts | 492 / 271 | 464 / 241 | 312 / 457 | 313 / 457 | 493 / 270 | 227 / 429 |
| utilitiesPublic.ts | 136 / 132 | 124 / 128 | 51 / 183 | 51 / 183 | 136 / 132 | 26 / 179 |
| visitorPublic.ts | 33 / 19 | 33 / 18 | 33 / 19 | 33 / 21 | 33 / 19 | 33 / 18 |
| watch.ts | 16 / 57 | 17 / 55 | 16 / 57 | 16 / 57 | 16 / 57 | 17 / 55 |
| watchPublic.ts | 6 / 6 | 6 / 6 | 6 / 6 | 6 / 6 | 6 / 6 | 6 / 6 |
| watchUtilities.ts | 7 / 7 | 7 / 6 | 5 / 5 | 5 / 5 | 8 / 6 | 6 / 4 |

# Per reason

Every cell is **measured on a checker-rejected program**. Reasons are the exact
lowerer strings, including concrete types. SkippedDependency rows are measurement
boundaries, not compiler blockers. Complete diagnostic text/locations are in JSON.

| Reason | Main | Taste | Flags | Namespaces | Nested | Cumulative |
|---|---:|---:|---:|---:|---:|---:|
| NotYet: .length on a value | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: ?.[] on a value | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: RegExp with a nonconstant pattern | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a BinaryExpression as a statement | 6 | 0 | 6 | 6 | 4 | 0 |
| NotYet: a BinaryExpression with a boolean and a boolean \| undefined | 1 | 0 | 1 | 1 | 1 | 0 |
| NotYet: a BinaryExpression with a boolean and a number | 0 | 0 | 1 | 1 | 0 | 0 |
| NotYet: a BinaryExpression with a boolean and a string | 2 | 0 | 2 | 2 | 2 | 0 |
| NotYet: a BinaryExpression with a boolean and a value | 6 | 0 | 8 | 8 | 6 | 0 |
| NotYet: a BinaryExpression with a number and a boolean | 0 | 0 | 3 | 3 | 0 | 0 |
| NotYet: a BinaryExpression with a number and a number | 6 | 1 | 7 | 7 | 5 | 1 |
| NotYet: a BinaryExpression with a number \| undefined and a number | 6 | 0 | 7 | 7 | 6 | 0 |
| NotYet: a BinaryExpression with a string and a boolean | 4 | 0 | 4 | 4 | 4 | 0 |
| NotYet: a BinaryExpression with a string and a number | 1 | 1 | 2 | 2 | 1 | 2 |
| NotYet: a BinaryExpression with a string and a string | 5 | 0 | 7 | 7 | 5 | 0 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 0 | 0 | 1 | 1 | 0 | 1 |
| NotYet: a BinaryExpression with a value and a boolean | 16 | 0 | 34 | 34 | 16 | 0 |
| NotYet: a BinaryExpression with a value and a number | 3 | 0 | 3 | 3 | 3 | 0 |
| NotYet: a BinaryExpression with a value and a string | 2 | 0 | 2 | 2 | 2 | 0 |
| NotYet: a BinaryExpression with a value and a value | 30 | 9 | 31 | 32 | 29 | 8 |
| NotYet: a ClassExpression | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a Map of T | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 8 | 10 | 8 | 9 | 5 | 7 |
| NotYet: a ModuleDeclaration | 9 | 9 | 9 | 0 | 9 | 0 |
| NotYet: a NonNullExpression | 21 | 24 | 23 | 23 | 20 | 25 |
| NotYet: a PostfixUnaryExpression | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a PrefixUnaryExpression on a boolean \| undefined | 5 | 0 | 5 | 5 | 5 | 0 |
| NotYet: a PrefixUnaryExpression on a number | 7 | 0 | 28 | 28 | 8 | 1 |
| NotYet: a PrefixUnaryExpression on a number \| undefined | 6 | 0 | 6 | 6 | 6 | 0 |
| NotYet: a PrefixUnaryExpression on a string | 7 | 0 | 8 | 8 | 8 | 0 |
| NotYet: a PrefixUnaryExpression on a union of differently held members | 2 | 0 | 2 | 2 | 2 | 0 |
| NotYet: a PrefixUnaryExpression on a value | 75 | 0 | 77 | 77 | 74 | 0 |
| NotYet: a SatisfiesExpression | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: a SpreadElement | 0 | 0 | 0 | 0 | 0 | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a call returning T | 0 | 0 | 0 | 0 | 1 | 1 |
| NotYet: a call through ?. (an optional call) | 10 | 11 | 10 | 11 | 11 | 13 |
| NotYet: a case whose type differs from the switch's | 0 | 0 | 3 | 3 | 0 | 3 |
| NotYet: a computed field name | 3 | 3 | 4 | 4 | 3 | 4 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 2 | 3 | 1 | 1 | 2 | 2 |
| NotYet: a destructured parameter beside a parameter with a default | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field holding union of differently held members | 0 | 3 | 0 | 0 | 0 | 3 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a field of type boolean \| undefined | 25 | 29 | 31 | 31 | 25 | 37 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field of type false \| string[] \| undefined | 0 | 1 | 0 | 0 | 0 | 1 |
| NotYet: a field of type string \| DiagnosticMessageChain | 4 | 4 | 4 | 4 | 4 | 6 |
| NotYet: a field of type string \| false \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field of type string \| number \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a field of type true \| Node \| undefined | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a field of type true \| undefined | 6 | 6 | 6 | 6 | 6 | 6 |
| NotYet: a function inside a function (a closure) | 3 | 3 | 3 | 3 | 0 | 0 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning CanonicalKey | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning CompilerOptionsValue | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function returning HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning ModeAwareCacheKey | 1 | 1 | 1 | 1 | 2 | 2 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 | 4 | 4 | 4 | 4 | 4 |
| NotYet: a function returning Path | 5 | 5 | 5 | 5 | 5 | 5 |
| NotYet: a function returning Path \| undefined | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function returning PathPathComponents | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning ResolvedConfigFileName | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a function returning ResolvedConfigFilePath | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning T | 48 | 48 | 48 | 48 | 48 | 48 |
| NotYet: a function returning T \| T[] | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning T \| T[] \| undefined | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a function returning T \| readonly T[] | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a function returning T \| undefined | 61 | 61 | 61 | 61 | 62 | 62 |
| NotYet: a function returning T1 & T2 | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning TOut | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning TOut \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning U | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function returning U \| undefined | 13 | 13 | 13 | 13 | 13 | 13 |
| NotYet: a function returning V | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function returning __String | 10 | 10 | 10 | 10 | 10 | 10 |
| NotYet: a function returning __String \| undefined | 4 | 4 | 4 | 4 | 4 | 4 |
| NotYet: a function returning any | 6 | 6 | 6 | 6 | 6 | 6 |
| NotYet: a function returning object | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function returning object \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning string \| object | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning unknown | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 0 | 1 | 1 | 1 | 0 |
| NotYet: a function returning void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 0 | 1 | 1 | 1 | 0 |
| NotYet: a function returning void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 0 | 1 | 1 | 1 | 0 |
| NotYet: a function value returning boolean \| undefined | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a function value returning union of differently held members | 2 | 3 | 2 | 2 | 5 | 6 |
| NotYet: a function value taking string \| DiagnosticMessageChain \| undefined | 0 | 0 | 0 | 0 | 1 | 1 |
| NotYet: a function value taking string \| string[] | 0 | 0 | 0 | 0 | 1 | 1 |
| NotYet: a function value with an optional parameter | 0 | 0 | 0 | 0 | 6 | 6 |
| NotYet: a function with an optional or rest parameter, as a value | 4 | 4 | 4 | 5 | 4 | 5 |
| NotYet: a function without a body | 191 | 191 | 191 | 192 | 193 | 194 |
| NotYet: a generic function as a value | 6 | 6 | 6 | 7 | 6 | 8 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 76 | 100 | 78 | 79 | 79 | 110 |
| NotYet: a namespace merged with a function; callable object properties, identity and receivers are not represented | 0 | 0 | 0 | 1 | 0 | 1 |
| NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 0 | 0 | 0 | 5 | 0 | 5 |
| NotYet: a parameter that isn't a plain name | 21 | 21 | 21 | 21 | 22 | 22 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 1 | 2 |
| NotYet: a union of differently held members variable a function value captures | 0 | 0 | 0 | 0 | 1 | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type (CompilerHost \| ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type BindableStaticNameExpression | 0 | 0 | 2 | 2 | 0 | 2 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type CompilerOptionsValue | 2 | 2 | 2 | 2 | 3 | 3 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 0 | 1 | 0 | 0 | 0 | 1 |
| NotYet: a value of type HasJSDoc | 1 | 2 | 1 | 1 | 1 | 2 |
| NotYet: a value of type HasJSDoc \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type K | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a value of type NamedEvaluation | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral \| StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 0 | 0 | 0 | 0 | 0 | 1 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type Path | 19 | 21 | 19 | 19 | 18 | 20 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type RequireOrImportCall | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type ResolvedConfigFileName | 6 | 6 | 6 | 6 | 7 | 7 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type ResolvedConfigFilePath | 18 | 18 | 18 | 18 | 18 | 18 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type SourceFile | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type T | 27 | 27 | 27 | 27 | 27 | 27 |
| NotYet: a value of type T \| Program | 4 | 4 | 4 | 4 | 4 | 4 |
| NotYet: a value of type T \| T[] | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a value of type T \| null \| undefined | 0 | 0 | 0 | 1 | 0 | 1 |
| NotYet: a value of type T \| readonly T[] | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type T \| undefined | 10 | 10 | 10 | 10 | 10 | 10 |
| NotYet: a value of type T1 | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type TData | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type TEntry | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type TInArray | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a value of type T["kind"] | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 0 | 0 | 1 | 1 | 0 | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type V | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a value of type WrappedExpression<T> | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type __String | 17 | 18 | 22 | 24 | 17 | 26 |
| NotYet: a value of type __String & string | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type any | 27 | 29 | 27 | 28 | 28 | 31 |
| NotYet: a value of type never | 0 | 0 | 0 | 1 | 0 | 1 |
| NotYet: a value of type object | 4 | 4 | 4 | 4 | 4 | 4 |
| NotYet: a value of type object \| undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 | 3 | 3 | 3 | 3 | 3 |
| NotYet: a value of type string \| null \| undefined | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: a value of type undefined | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: a value of type unknown | 9 | 9 | 9 | 11 | 9 | 11 |
| NotYet: a void call used as a value | 1 | 1 | 1 | 6 | 1 | 8 |
| NotYet: an ElementAccessExpression | 3 | 6 | 3 | 3 | 3 | 6 |
| NotYet: an EnumDeclaration | 154 | 154 | 0 | 0 | 154 | 0 |
| NotYet: an array of T | 4 | 6 | 4 | 4 | 4 | 6 |
| NotYet: an array of T \| U | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: an array of never | 5 | 5 | 5 | 5 | 5 | 6 |
| NotYet: an array of unknown | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: an enum inside a function or block; declare it at module scope | 0 | 0 | 9 | 3 | 0 | 3 |
| NotYet: an optional chain longer than one step | 0 | 0 | 1 | 1 | 2 | 4 |
| NotYet: assigning a field of a value | 1 | 1 | 1 | 1 | 2 | 2 |
| NotYet: assigning an element of a value | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: destructuring inside a namespace; use plain singleton bindings | 0 | 0 | 0 | 1 | 0 | 1 |
| NotYet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 0 | 1 | 0 | 0 | 0 | 1 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 4 | 4 | 4 | 4 | 4 | 4 |
| NotYet: for...of over an object | 7 | 9 | 8 | 8 | 7 | 11 |
| NotYet: lastIndexOf with these arguments | 1 | 2 | 1 | 1 | 1 | 2 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: new a ParenthesizedExpression | 2 | 2 | 2 | 2 | 3 | 3 |
| NotYet: new an Identifier | 7 | 7 | 7 | 7 | 7 | 7 |
| NotYet: optional chaining to .size on a value | 1 | 1 | 1 | 1 | 2 | 2 |
| NotYet: reading AccessKind | 3 | 3 | 0 | 0 | 3 | 0 |
| NotYet: reading AssignmentDeclarationKind | 6 | 6 | 0 | 0 | 6 | 0 |
| NotYet: reading AssignmentKind | 0 | 1 | 0 | 0 | 0 | 0 |
| NotYet: reading BuilderFileEmit | 2 | 3 | 0 | 0 | 2 | 0 |
| NotYet: reading BuilderProgramKind | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading CharacterCodes | 35 | 37 | 0 | 0 | 35 | 0 |
| NotYet: reading Comparison | 9 | 11 | 0 | 0 | 10 | 0 |
| NotYet: reading DiagnosticCategory | 3 | 3 | 0 | 0 | 3 | 0 |
| NotYet: reading EmitFlags | 7 | 7 | 0 | 0 | 7 | 0 |
| NotYet: reading Error | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: reading Extension | 20 | 23 | 0 | 0 | 20 | 0 |
| NotYet: reading Extensions | 5 | 6 | 0 | 0 | 5 | 0 |
| NotYet: reading FileIncludeKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading FileWatcherEventKind | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading ForegroundColorEscapeSequences | 0 | 2 | 0 | 0 | 0 | 0 |
| NotYet: reading GetLiteralTextFlags | 0 | 1 | 0 | 0 | 0 | 0 |
| NotYet: reading ImportsNotUsedAsValues | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading Instruction | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading InternalNodeBuilderFlags | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading IntrinsicTypeKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading IterationTypeKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading JSDocParsingMode | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading JsxEmit | 2 | 3 | 0 | 0 | 2 | 0 |
| NotYet: reading ListFormat | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading ModifierFlags | 22 | 22 | 0 | 0 | 22 | 0 |
| NotYet: reading ModuleDetectionKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading ModuleInstanceState | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading ModuleKind | 11 | 13 | 0 | 0 | 11 | 0 |
| NotYet: reading ModuleResolutionKind | 6 | 7 | 0 | 0 | 4 | 0 |
| NotYet: reading ModuleSpecifierEnding | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading NewLineKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading NodeBuilderFlags | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading NodeFactoryFlags | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading NodeFlags | 20 | 23 | 0 | 0 | 20 | 0 |
| NotYet: reading NodeResolutionFeatures | 7 | 8 | 0 | 0 | 7 | 0 |
| NotYet: reading OuterExpressionKinds | 3 | 4 | 0 | 0 | 3 | 0 |
| NotYet: reading PragmaKindFlags | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading RegularExpressionFlags | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading ScriptKind | 4 | 7 | 0 | 0 | 4 | 0 |
| NotYet: reading ScriptTarget | 5 | 5 | 0 | 0 | 5 | 0 |
| NotYet: reading SignatureFlags | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading SnippetKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading StatisticType | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading SymbolFlags | 5 | 5 | 0 | 0 | 5 | 0 |
| NotYet: reading SyntaxKind | 481 | 503 | 0 | 0 | 482 | 0 |
| NotYet: reading TokenFlags | 0 | 1 | 0 | 0 | 0 | 0 |
| NotYet: reading TransformFlags | 5 | 7 | 0 | 0 | 5 | 0 |
| NotYet: reading TypeFacts | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading TypeFlags | 3 | 3 | 0 | 0 | 3 | 0 |
| NotYet: reading TypePredicateKind | 2 | 2 | 0 | 0 | 2 | 0 |
| NotYet: reading UpToDateStatusType | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading UsingKind | 3 | 3 | 0 | 0 | 3 | 0 |
| NotYet: reading WatchFileKind | 1 | 1 | 0 | 0 | 1 | 0 |
| NotYet: reading addAggregateStatistic | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading addOutput | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading captureMapping | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading convertToFunctionBlock | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading createBaseSourceFileNode | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading createIntlCollatorStringComparer | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading createPollingIntervalQueue | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading enter | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading getAccessorNameVisibilityError | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading getOptionsNameMap | 2 | 2 | 2 | 2 | 2 | 2 |
| NotYet: reading getPackageJsonInfo | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading getSymbolWalker | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading inferPreference | 0 | 0 | 1 | 1 | 0 | 0 |
| NotYet: reading locationInfo | 0 | 1 | 0 | 0 | 0 | 0 |
| NotYet: reading lookupFromPackageJson | 0 | 0 | 1 | 1 | 0 | 0 |
| NotYet: reading nextPollIndex | 0 | 1 | 0 | 0 | 0 | 0 |
| NotYet: reading optionDependsOnRecursive | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading transformSourceFile | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: reading transformSourceFileOrBundle | 1 | 1 | 1 | 1 | 0 | 0 |
| NotYet: regex replacement other than a string | 4 | 4 | 6 | 6 | 4 | 6 |
| NotYet: spreading an array of other elements | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: storing any in a field | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: storing boolean in a field | 0 | 0 | 0 | 0 | 1 | 1 |
| NotYet: storing string \| number in a field | 1 | 1 | 1 | 1 | 1 | 1 |
| NotYet: storing true \| Node \| undefined in a field | 1 | 1 | 1 | 1 | 1 | 2 |
| NotYet: this outside a method | 7 | 7 | 7 | 7 | 7 | 7 |
| Refused: Object.defineProperty | 0 | 1 | 0 | 0 | 0 | 1 |
| Refused: a boolean \| undefined as a condition | 11 | 0 | 12 | 12 | 11 | 0 |
| Refused: a cast the runtime can't check | 23 | 25 | 57 | 57 | 22 | 69 |
| Refused: a definite assignment assertion ! | 5 | 5 | 5 | 5 | 5 | 5 |
| Refused: a flag initializer outside the non-negative int32 bound | 0 | 0 | 3 | 0 | 0 | 3 |
| Refused: a function taking (node: Node) => T \| undefined seen as one taking (node: Node) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking (value: never, key: never, map: ReadonlyMap<never, never>) => void seen as one taking <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey], key: TKey, map: ReadonlyPragmaMap) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 6 | 6 | 6 | 6 | 6 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking OrdinalParentheizerRuleSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking ParenthesizerRule<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking ParenthesizerRuleOrSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [data: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... \| undefined] seen as one taking readonly StringLiteralLike[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [moduleNames: string[], containingFile: string, reusedNames: string[] \| undefined, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile?: SourceFile \| undefined] seen as one taking string[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [name: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [path: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 7 | 7 | 7 | 7 | 7 | 7 |
| Refused: a function taking [s: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [typeDirectiveReferences: readonly (string \| FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile \| undefined, reusedNames: ... \| undefined] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [typeReferenceDirectiveNames: string[] \| readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingFileMode?: ResolutionMode] seen as one taking string[] \| readonly FileReference[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a function taking readonly string[] seen as one taking readonly string[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking string seen as one taking [data: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking string seen as one taking [name: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking string seen as one taking [path: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a generator function | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a label | 3 | 0 | 3 | 3 | 3 | 0 |
| Refused: a method in object destructuring | 4 | 4 | 4 | 4 | 3 | 3 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createHash would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLink would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 6 | 6 | 6 | 6 | 6 | 6 |
| Refused: a method read as a value (getCommonSourceDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getDefaultLibLocation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (getModuleResolutionCache would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 5 | 5 | 5 | 5 | 5 | 5 |
| Refused: a method read as a value (now would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a method read as a value (readDirectory would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (resolveLibrary would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (resolveTypeReferenceDirectiveReferences would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (resolveTypeReferenceDirectives would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a namespace | 11 | 11 | 11 | 0 | 11 | 0 |
| Refused: a non-exhaustive enum switch; missing ModuleKind.CommonJS | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: a non-exhaustive enum switch; missing ModuleResolutionKind.Classic | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown | 0 | 0 | 30 | 30 | 0 | 33 |
| Refused: a number as a condition | 1 | 0 | 10 | 10 | 1 | 0 |
| Refused: a number outside the proven flag domain assigned to CheckFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 2 | 0 | 0 | 2 |
| Refused: a number outside the proven flag domain assigned to Connection.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 3 | 0 | 0 | 3 |
| Refused: a number outside the proven flag domain assigned to EmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 1 | 0 | 0 | 1 |
| Refused: a number outside the proven flag domain assigned to Extensions.TypeScript; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 6 | 0 | 0 | 6 |
| Refused: a number outside the proven flag domain assigned to Extensions; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 1 | 0 | 0 | 1 |
| Refused: a number outside the proven flag domain assigned to InternalEmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 2 | 0 | 0 | 2 |
| Refused: a number outside the proven flag domain assigned to LexicalEnvironmentFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 2 | 0 | 0 | 2 |
| Refused: a number outside the proven flag domain assigned to ObjectFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 0 | 3 | 0 | 0 | 3 |
| Refused: a number \| undefined as a condition | 2 | 0 | 2 | 2 | 1 | 0 |
| Refused: a parameter property | 0 | 0 | 6 | 0 | 0 | 0 |
| Refused: a string as a condition | 14 | 0 | 14 | 14 | 15 | 0 |
| Refused: a type predicate | 581 | 581 | 581 | 581 | 581 | 581 |
| Refused: a union of differently held members as a condition | 3 | 0 | 3 | 3 | 3 | 0 |
| Refused: a value as a condition | 50 | 0 | 54 | 54 | 49 | 0 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type AccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ArrayBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type BigIntLiteral \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type BindingElement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 5 | 5 | 5 | 5 |
| Refused: a value of type BuilderProgram seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what BuilderProgram can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type BuilderProgram \| Program seen as BuilderProgram, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 5 | 5 | 5 | 5 | 5 | 5 |
| Refused: a value of type CallExpression seen as RequireOrImportCall, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & Identifier would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfListType, which can write "list" \| "listOrElement" where "boolean" is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption \| undefined seen as TsConfigOnlyOption, which can write "object" where "boolean" is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CommandLineOptionOfBooleanType \| CommandLineOptionOfCustomType \| CommandLineOptionOfNumberType \| CommandLineOptionOfStringType \| TsConfigOnlyOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Declaration seen as T, a type parameter whose constraint Declaration can be written, so it can write what Declaration can't hold | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type DiagnosticMessageChain seen as T, a type parameter whose constraint DiagnosticMessageChain \| ReusableDiagnosticMessageChain can be written, so it can write what DiagnosticMessageChain can't hold | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 10 | 10 | 10 | 10 | 10 | 10 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 10 | 10 | 10 | 10 | 10 | 10 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type ElementAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type EvaluatorResult<number> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where number is read | 16 | 16 | 16 | 16 | 16 | 16 |
| Refused: a value of type EvaluatorResult<string \| undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 6 | 6 | 6 | 6 | 6 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Expression \| Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ExpressionWithTypeArguments seen as ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; }, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & (Identifier \| PropertyAccessEntityNameExpression) would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type FlowNode seen as FlowLabel, which can write undefined where BinaryExpression \| CallExpression is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Identifier seen as T, a type parameter whose constraint Node can be written, so it can write what Identifier can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type ImportTypeNode \| undefined seen as ValidImportTypeNode \| undefined, whose readonly field argument becomes writable: a readonly field may hold something narrower than TypeNode, which a write of LiteralTypeNode & { literal: StringLiteral; } would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type JSDocNullableType seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type JsxOpeningFragment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type JsxTagNameExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type LeftHandSideExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type LeftHandSideExpression \| UnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type LiteralLikeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<CharacterCodes, RegularExpressionFlags> seen as Map<CharacterCodes, number>, which can write number where RegularExpressionFlags is read | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, SyntaxKind> seen as Map<string, number>, which can write number where SyntaxKind is read | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 | 10 | 10 | 10 | 10 | 10 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type MethodDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type Node seen as SyntaxList, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.SyntaxList would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node can be written, so it can write what Node can't hold | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what Node can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node seen as TemplateLiteralTypeSpan, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.TemplateLiteralType would replace | 1 | 1 | 0 | 0 | 1 | 0 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type Node \| NodeArray<Node> seen as NodeArray<Node>, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type NodeArray<ClassElement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type NodeArray<T> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type ObjectBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ParameterDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type PostfixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type PostfixUnaryExpression \| PrefixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type PrivateIdentifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type PropertyAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type PropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type PropertyName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type QualifiedName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, which can write { resolved: Resolved; isExternalLibraryImport: true; } \| undefined where Resolved \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SearchResult<undefined> seen as SearchResult<Resolved>, which can write Resolved \| undefined where undefined is read | 6 | 6 | 6 | 6 | 6 | 6 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SolutionBuilderHost<T> seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where ((path: string, data: string, writeByteOrderMark?: boolean \| undefined) => void) \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 24 | 24 | 24 | 24 | 24 | 24 |
| Refused: a value of type StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 6 | 6 | 6 | 6 | 6 | 6 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type T seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type TemplateLiteralLikeNode seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Token<SyntaxKind> seen as T, a type parameter whose constraint Node can be written, so it can write what Token<SyntaxKind> can't hold | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as CompilerOptionsValue, which can write string \| number where string is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type TypeCheckerHost seen as Program, whose readonly field redirectTargetsMap becomes writable: a readonly field may hold something narrower than RedirectTargetsMap, which a write of MultiMap<Path, string> would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type VariableDeclarationList seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type VisitEachChildTable seen as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>, which can write VisitEachChildFunction<any> \| undefined where VisitEachChildFunction<QualifiedName> is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type any seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what any can't hold | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 7 | 7 | 7 | 7 | 7 | 7 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 | 3 | 3 | 3 | 3 | 3 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { host: ModuleResolutionHost; compilerOptions: CompilerOptions; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: ... \| undefined; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write DiagnosticReporter where (_?: unknown) => void is read | 1 | 1 | 0 | 0 | 1 | 0 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { kind: "ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } \| undefined seen as ModuleSpecifierResult \| undefined, which can write boolean where false is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { matchableStringSet: Set<string> \| undefined; patterns: Pattern[] \| undefined; } seen as ParsedPatterns, which can write ReadonlySet<string> \| undefined where Set<string> \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined seen as Resolved \| undefined, which can write string \| true \| undefined where string \| true is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { resolved: Resolved; isExternalLibraryImport: true; } \| undefined seen as { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined, which can write boolean where true is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 14 | 14 | 14 | 14 | 14 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: false; }; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: false; } is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: true; } \| undefined is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 11 | 11 | 11 | 11 | 11 | 11 |
| Refused: an ExportDeclaration | 77 | 0 | 77 | 77 | 77 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to AccessKind.Read; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to BuilderFileEmit.None; its members are a closed union | 0 | 0 | 0 | 12 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.plus; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.slash; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CheckFlags.None; its members are a closed union | 0 | 0 | 0 | 2 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to Connection.None; its members are a closed union | 0 | 0 | 0 | 3 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to EmitFlags.None; its members are a closed union | 0 | 0 | 0 | 4 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to Extensions.TypeScript; its members are a closed union | 0 | 0 | 0 | 13 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to FlowFlags.Unreachable; its members are a closed union | 0 | 0 | 0 | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to GeneratedIdentifierFlags.None; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to InternalEmitFlags.None; its members are a closed union | 0 | 0 | 0 | 3 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to InvalidPosition; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to LexicalEnvironmentFlags.None; its members are a closed union | 0 | 0 | 0 | 2 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to ModifierFlags.None; its members are a closed union | 0 | 0 | 15 | 15 | 0 | 15 |
| Refused: an arbitrary number or a value from another enum assigned to ModuleKind.ESNext; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to NodeFactoryFlags.None; its members are a closed union | 0 | 0 | 0 | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to NodeFlags.NestedNamespace; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to NodeFlags.None; its members are a closed union | 0 | 0 | 17 | 17 | 0 | 17 |
| Refused: an arbitrary number or a value from another enum assigned to NodeResolutionFeatures.None; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to ObjectFlags.None; its members are a closed union | 0 | 0 | 0 | 3 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to OperatorPrecedence.Invalid; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to OuterExpressionKinds.Parentheses; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ParsingContext.SourceElements; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to Phase.Parse; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to ProgramUpdateLevel.Update; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to PropertyLikeParse.Property; its members are a closed union | 0 | 0 | 0 | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to SignatureFlags.None; its members are a closed union | 0 | 0 | 0 | 8 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to SymbolFlags.None; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayLiteralExpression; its members are a closed union | 0 | 0 | 6 | 6 | 0 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayType; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BinaryExpression; its members are a closed union | 0 | 0 | 20 | 20 | 0 | 20 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BindingElement; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CallExpression; its members are a closed union | 0 | 0 | 7 | 7 | 0 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CaseBlock; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CatchClause; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassDeclaration; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassExpression; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CommaToken; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ComputedPropertyName; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Constructor; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Decorator; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ElementAccessExpression; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.EnumDeclaration; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportAssignment; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportDeclaration; its members are a closed union | 0 | 0 | 13 | 13 | 0 | 13 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportSpecifier; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionStatement; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionWithTypeArguments; its members are a closed union | 0 | 0 | 10 | 10 | 0 | 10 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExternalModuleReference; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ForStatement; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionDeclaration; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionExpression; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.HeritageClause; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Identifier; its members are a closed union | 0 | 0 | 15 | 15 | 0 | 15 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportClause; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportDeclaration; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportEqualsDeclaration; its members are a closed union | 0 | 0 | 5 | 5 | 0 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportSpecifier; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDoc; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocEnumTag; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocReturnTag; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypeExpression; its members are a closed union | 0 | 0 | 7 | 7 | 0 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypedefTag; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxElement; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxNamespacedName; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.LabeledStatement; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MetaProperty; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MethodDeclaration; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MinusToken; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ModuleDeclaration; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceExport; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceImport; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NewExpression; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NoSubstitutionTemplateLiteral; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NumericLiteral; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ObjectLiteralExpression; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Parameter; its members are a closed union | 0 | 0 | 8 | 8 | 0 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedExpression; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedType; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PrefixUnaryExpression; its members are a closed union | 0 | 0 | 7 | 7 | 0 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAccessExpression; its members are a closed union | 0 | 0 | 6 | 6 | 0 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAssignment; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyDeclaration; its members are a closed union | 0 | 0 | 5 | 5 | 0 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.QualifiedName; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SatisfiesExpression; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ShorthandPropertyAssignment; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SourceFile; its members are a closed union | 0 | 0 | 4 | 4 | 0 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.StringLiteral; its members are a closed union | 0 | 0 | 8 | 8 | 0 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateExpression; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateHead; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateLiteralTypeSpan; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateSpan; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeParameter; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeReference; its members are a closed union | 0 | 0 | 3 | 3 | 0 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableDeclaration; its members are a closed union | 0 | 0 | 6 | 6 | 0 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableStatement; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to TokenFlags.None; its members are a closed union | 0 | 0 | 0 | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to TransformFlags.None; its members are a closed union | 0 | 0 | 2 | 2 | 0 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to TypeReferenceSerializationKind.Unknown; its members are a closed union | 0 | 0 | 1 | 1 | 0 | 1 |
| Refused: an import cycle | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: an index signature | 7 | 7 | 7 | 7 | 7 | 7 |
| Refused: arithmetic assigned back into an enum; the result need not be one of its members | 0 | 0 | 42 | 50 | 0 | 42 |
| Refused: debugger | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: delete | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: enum | 162 | 162 | 0 | 0 | 162 | 0 |
| Refused: export * | 0 | 75 | 0 | 0 | 0 | 75 |
| Refused: in | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: inherited library member compare read as an own field | 0 | 0 | 0 | 0 | 0 | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: inherited library member prototype read as an own field | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: inherited library member replace read as an own field | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 10 more ... \| undefined written where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type ClassElement \| undefined written where ClassElement is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type CommandLineOption \| undefined written where CommandLineOption is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type CommentRange \| undefined written where CommentRange is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type Diagnostic \| undefined written where Diagnostic is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type EmitHelper \| undefined written where EmitHelper is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type JsxChild \| undefined written where JsxChild is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type T \| undefined written where T is read | 2 | 2 | 2 | 2 | 2 | 2 |
| Refused: instantiating a generic function makes a value of type VariableDeclaration \| undefined written where VariableDeclaration is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type string \| undefined written where string is read | 1 | 1 | 1 | 1 | 1 | 1 |
| Refused: the comma operator | 4 | 0 | 4 | 4 | 4 | 0 |
| Refused: the non-null assertion ! | 180 | 180 | 180 | 180 | 180 | 180 |
| Refused: the void operator | 3 | 0 | 3 | 3 | 3 | 0 |
| Refused: this in a namespace function; a qualified call and a detached call have different receivers | 0 | 0 | 0 | 1 | 0 | 1 |
| Refused: var | 2 | 2 | 2 | 2 | 1 | 1 |
| Refused: yield (generators) | 4 | 4 | 4 | 4 | 4 | 4 |
| Refused: \|\|= | 15 | 0 | 15 | 15 | 15 | 0 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 23 | 25 | 23 | 23 | 23 | 25 |
| error: lower: src/compiler/parser.ts:1472:9: the checker gave a declaration no symbol | 0 | 0 | 0 | 1 | 0 | 1 |

# Mutants and validation

`audit.py` checks continuation through two failing functions, two refusal sites in
one function, an imported sibling call, diagnosed-body skips, signature-only
eligibility, same-line function ranges, and measurement labels. An overlay-only
NotYet at `one.a:3:1` adds one finding in one.a and leaves two.a identical. A
misattribution mutant fails that equality. Expanding body scope to include the
signature wrongly skips `signatureOnly`; the eligibility check catches it.
`data/audit.log` records these checks. Synthetic probes are scratch .a programs,
not TypeScript-derived runtime fixtures or Node/native comparisons.

`audit_corpus.py` plants the extra NotYet in the real eligible tsc function
`getModuleInstanceState`, `src/compiler/binder.ts:330:1`. Binder's unique-site count
rises by exactly one; all other 77 files' entire records and every unit's eligibility
remain unchanged. `data/corpus-mutant-audit.json` records every file's delta.
`data/corpus-mutant-run.log` preserves the run. The baseline contains no mutant.

`audit_output_guards.py` builds two scratch compiler mutants. Non-nil IR from Lower
is caught by `measurement returned usable IR`; exposing the permissive loader through
ordinary Load is caught by `measurement loader exposed an output program`. Both
audit runs fail as intended, with exact guard panic logs preserved in data/.

`audit_report.py` independently verifies adapted source hashes, root coverage,
checker spans, body eligibility, measurement labels, unique findings, attempt events,
per-file reasons and feature deltas. Mutants alter a source hash, drop a file, inflate
checker/finding totals, change a reason/per-file/unit count, remove the label, change the outside-root bucket, alter
a feature delta/common-unit count, and flip a raw skipped-body status. Every mutant is caught.
`data/report-audit.log` records their failures. The actual overlaid scratch compiler
passes `go vet`; all Python scripts pass syntax compilation. All test outputs go to logs.

Initial setup: Go 1.27.1 ready in 0s; clang 20.1.8, Node 24.19.0 and submodules ready
by 1s; cache warming 119s; total 119s. `nproc` is 5; cgroup quota is four CPUs.
`data/setup.log` preserves timing lines.

Commands used after the scratch merges and adaptation:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent-adapted /tmp/latent-complete-runs /tmp/latent-rejected-worktrees.json > /tmp/latent-complete-comparisons.log 2>&1
python3 stage3/census/latent/audit.py /tmp/latent-complete-runs/main-census > /tmp/latent-complete-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent-complete-runs /tmp/tsc-latent-adapted > /tmp/latent-complete-corpus-audit.log 2>&1
python3 stage3/census/latent/audit_output_guards.py /tmp/latent-comparisons/main /tmp/latent-complete-runs/main-overlay /tmp/latent-output-guards > /tmp/latent-output-guards.log 2>&1
python3 stage3/census/latent/summarize.py /tmp/latent-complete-runs /tmp/tsc-latent-adapted > /tmp/latent-complete-summary.log 2>&1
python3 stage3/census/latent/write_report.py > /tmp/latent-complete-report.log 2>&1
python3 stage3/census/latent/audit_report.py /tmp/tsc-latent-adapted > /tmp/latent-complete-report-audit.log 2>&1
cd /tmp/latent-comparisons/main
go vet -overlay=/tmp/latent-complete-runs/main-overlay/overlay.json ./stage3/census/latent/tool > /tmp/latent-complete-vet.log 2>&1
```

`prepare_scratch.py REPOSITORY NEW_TMP_DIRECTORY UNIQUE_BRANCH_PREFIX` reproduces
individual/cumulative scratch configurations and writes worktrees.json for the runner.
Main, feature, and preparation commits are pinned by REPORT.json. Apply adaptations from the new main scratch tree, then measure every
configuration on that one tree. Observed resolution recipes were run; a second full
scratch preparation cycle was not repeated.

# Limits and unmeasured work

Lowering stops at its first returned error within each unit. The refusal visitor
can expose multiple syntax refusals, while type-dependent helpers can still return
their first error. Generic declarations are attempted without invented substitutions.
Symbol registration is best effort; some reads may reflect isolated context rather
than successful whole-program lowering. Diagnosed dependency bodies remain skipped
measurement boundaries. The final module-order, ownership, and backend passes are
omitted; local operations may still produce their own cycle/readiness refusals.
Panics/ordinary errors are explicitly separated from NotYet/Refused.

I did not run TypeScript's upstream suite, native-vs-Node comparisons, or the full
Adamic compiler gate. This binary cannot produce runnable native output, and every
adapted configuration remains checker-rejected. The ledger records latent lowering
observations under the stated rules, not native tsc readiness or correctness.
