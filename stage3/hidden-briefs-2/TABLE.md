# Hidden boundary briefs, wave 2

Roadmap step 30, measured by step 05's corrected census. Documentation only: no compiler implementation, new oracle fixture, refreshed oracle counts, or measured revealed bytes.

## Pins and method

Delivery base: `origin/compiler/area-next-fixtures`, `4885cec50290686df487b62aac47c85d871ed40c`. Ranking: `codex/stage3-hidden-ranking`, `ec0b16c04f3bbdfbaf3932ab01307f90b08156fd`, `stage3/census/hidden-ranking/RESULT.json`. First-wave form: `codex/hidden-boundary-briefs`, `3f1b17a270aae45b9f2ae06924bc5bde31bc1bd7`, `stage3/hidden-briefs/TABLE.md`, briefs 12 and 08.

The corrected ranking measured compiler `69501280a81259fb512edbb8dd0e52c6eb0d88c8`: 3,654,880 hidden bytes of 10,615,807 (34.428659%). Its 82 adapted source files match the hash manifest in `stage3/census/hidden/RESULT.json` at the ranking pin. Prepare the adapted source from census pin `388096e6` using its recorded adaptation procedure; verify every manifest hash before replay. The source used here was `/tmp/hidden-adapted/src/compiler`. No cohere source was copied into this branch.

Ranking credits use outermost-cause attribution. `bytes_revealed_if_fixed_alone` is attribution, not an observed counterfactual. These briefs merge adjacent attributed spans with the same reason to name regions. They do not claim that all credited bytes become accepted after removing a stop, or that a whole surrounding function is revealed. Overlapping/enclosing boundaries can still block them.

The top 30 credit 1,477,735 bytes: area-views 824,298; overload results 218,748; other existing units 400,209; this unit 34,480. One new brief remains after ownership exclusions.

## Ownership ledger

| Rank | Historical stop | Credited bytes | Owner / disposition |
| --- | --- | ---: | --- |
| 1 | NotYet: a value of type __String | 168,577 | area-views |
| 2 | NotYet: a function returning CapturedThis | 129,056 | area-views |
| 3 | NotYet: a value of type InitializedVariableDeclaration | 117,867 | area-views |
| 4 | NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 117,289 | codex/census-generic-returns (structural method statics) |
| 5 | NotYet: a computed field name | 113,387 | codex/notyet-destructuring (computed keys) |
| 6 | NotYet: a value of type PrivateIdentifierInExpression | 91,578 | area-views |
| 7 | NotYet: a function returning ImmediatelyInvokedArrowFunction | 87,684 | area-views |
| 8 | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number | 62,495 | codex/overload-results |
| 9 | NotYet: a value of type Path | 61,844 | area-views |
| 10 | NotYet: a value of type ParameterPropertyDeclaration | 59,747 | area-views |
| 11 | Refused: overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 59,499 | codex/overload-results |
| 12 | Refused: overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 41,822 | codex/overload-results |
| 13 | NotYet: a value of type PrimitiveLiteral | 37,568 | area-views |
| 14 | Refused: a method in object destructuring | 34,480 | This unit: [01-method-destructuring.md](01-method-destructuring.md) |
| 15 | NotYet: a value of type any | 29,361 | TypeScript adaptations (any values) |
| 16 | Refused: overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 28,699 | codex/overload-results |
| 17 | Refused: a cast the runtime can't check | 27,462 | area-views |
| 18 | NotYet: a destructured name that isn't plain | 26,845 | codex/notyet-destructuring (binding patterns) |
| 19 | NotYet: a function returning T \| undefined | 24,802 | codex/census-generic-returns and codex/scout-generics (step 16) |
| 20 | NotYet: a value of type ResolvedConfigFileName | 20,619 | area-views |
| 21 | NotYet: a function returning T | 18,006 | codex/census-generic-returns and codex/scout-generics (step 16) |
| 22 | NotYet: for...of over an object | 16,604 | Step 20 object iteration scout |
| 23 | Refused: Object.entries | 15,966 | Step 20 object iteration scout |
| 24 | Refused: var | 15,286 | TypeScript adaptations (var) |
| 25 | Refused: overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 14,206 | codex/overload-results |
| 26 | NotYet: an array of never | 12,123 | codex/parser-generic-empty-array (never-array representation) |
| 27 | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 12,027 | codex/overload-results |
| 28 | NotYet: a function returning Path | 11,989 | area-views |
| 29 | NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10,540 | Step 20 object iteration scout |
| 30 | NotYet: a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 10,307 | area-views |

## Exclusions and conservative ownership assumptions

Checked AST views and phantom brands belong to area-views, including PrimitiveLiteral and intersection views. Stricter compiler options belong to task #k881crd; the separate checker bucket is not allocated wholesale to that task. Optional calls belong to step 18 scout; exceptions to step 21; Maps and Sets to step 19; assignment expressions to #cvhj5fk. None of those four exact reason classes occurs in this top 30. Object.entries is object iteration, not Maps/Sets.

Existing pending units are excluded even when their diagnostics still reproduce on this base. Read-only ownership evidence: `codex/notyet-destructuring` at `903ba7b229bad116798714b05a0ef93164da7316`, `stage3/notyet-destructuring/REPORT.md`; `codex/census-generic-returns` at `c40266051c9292ec36a779226b7ba26764b6060d`, `internal/lower/CENSUS_GENERIC_RETURNS_RESUME.md`; `codex/parser-generic-empty-array` at `4e022aaef7d24085fd0ded0c11b539843022fc47`, `internal/lower/PARSER_EMPTY_ARRAYS.md`; generic scout `e685ca3c09bb5f86a5b54f68582ecf6a70112692`.

The empty-array unit owns the never-array representation even though mutable aliases remain refused. Generic units own concrete specialization and its uninstantiated census boundary. Neither exclusion claims the stop is already fixed here. The destructuring unit's completed lesson list covers computed keys and binding patterns, but does not cover proving receiver-independent method extraction. We conservatively assign that remaining method lesson here; its implementation must coordinate the shared `destructureFrom` function after the pending unit lands. No other worker's commits were merged.

## Replays on this base

Build the inherited full-project replay overlay, preserving ancestor binding context:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-wave-2/overlay > /tmp/hidden-wave-2/overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/hidden-wave-2/overlay/overlay.json -o /tmp/hidden-wave-2/replay ./stage3/census/latent/replay/worker > /tmp/hidden-wave-2/build.log 2>&1
```

Each run used `GOMAXPROCS=1`, `-project /tmp/hidden-adapted/src/compiler`, the actual diagnostic position (not the attempted-unit start), and its exact kind/reason. Exit 0 means exact position/kind/reason reproduced, not successful program compilation. Every replay is marked “measured on a checker-rejected program”; these are latent lowering observations, not accepted whole-project behavior. Logs are `/tmp/hidden-wave-2/replay-NN.log`.

| Rank | Actual diagnostic position | Result |
| --- | --- | --- |
| 4 | factory/emitHelpers.ts:636:99 | Exact NotYet reproduced, exit 0 |
| 5 | visitorPublic.ts:624:5 | Exact NotYet reproduced, exit 0 |
| 14 | sys.ts:597:5 | Exact Refused reproduced, exit 0 |
| 18 | utilities.ts:11499:5 | Exact NotYet reproduced, exit 0 |
| 19 | moduleNameResolver.ts:1189:14 | Exact NotYet reproduced, exit 0 |
| 21 | transformer.ts:358:14 | Exact NotYet reproduced, exit 0 |
| 26 | utilities.ts:9049:23 | Exact NotYet reproduced, exit 0 |

The largest eligible contiguous region was separately replayed at utilities.ts:11316:35; exact Refused reproduced, exit 0, `/tmp/hidden-wave-2/replay-14-largest.log`. Excluded view/overload stops were classified from ownership and source evidence, not all independently replayed.

## Toolchain and validation scope

`GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh` succeeded; env `/workspace/adamic-tools/env.sh`; `nproc` 5, CPU quota 4. Node 24.19.0, Go 1.27.1, clang 20.1.8. Setup timing lines: node 0.022s; go 0.022s; submodules 0.055s; markdown dependency step 0.006s (validated installed bytes), ready 0.062s; clang 0.161s; Go build 35.771s; test binaries deferred 35.918s; cache warm 35.920s; done 35.945s.

Only documentation validation, focused replay and Node witness observations were run. The documentation ledger is checked against the pinned ranking and source hashes; independent corruptions of byte totals, span endpoints and ownership must fail that validator. Compiler mutants and sanitizer/backend oracles are requirements for the future implementation, not claims of this documentation unit. No full package tests or full gate were run, and no counts refresh is needed because no oracle fixture was added.

Validation observed: all 30 rows, one eligible reason, 47 spans totaling 34,480 bytes, all 82 source hashes and both Node outputs passed (`/tmp/hidden-wave-2/validation.log`). Independent documentation mutants (credit +1, endpoint +1, owned computed-key reason reassigned) each failed the ledger validator. A replay-selector mutant appending `MUTANT` to the exact reason exited 1 (`/tmp/hidden-wave-2/replay-mutant.log`). These are evidence-validation mutations; no compiler mutant was run.
