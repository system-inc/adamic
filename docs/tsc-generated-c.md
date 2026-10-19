# tsc's generated C, measured per file (#0gvtg3y)

Loom sizes its C object cache from the size of the C Adamic emits for TypeScript's own compiler. This page records what that C measures today, what it can't measure yet and why, a projection to tsc's 77 files, and the facts a split policy in the emitter needs.

## Inputs and box

- Adamic at origin/main `dd5009aa`, built with Go 1.27.1. This page is based on the next commit, `1adfd80b`, which changes none of internal/lower, internal/native, internal/load, cmd/adamic, cmd/adamic-meter, stage3/apply.py or stage3/adapt.
- TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, adapted by `stage3/apply.sh` (adaptations 00 through 75 at that main).
- clang 20.1.8, the official LLVM release (`87f0227cb601`), with `native.Flags`' release flags: `-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread`, plus `-O0`, `-O1` or `-O2` and `-I internal/native/runtime`. Each run is `clang -c` of one emitted file, which is all `adamic build` compiles per program: the runtime is a separately cached archive and isn't counted.
- Box: Intel Xeon Processor @ 2.10GHz, 4 cores, 15.7 GiB, no swap, Linux 6.18. One measurement at a time, nothing else running. The one-minute load average was 1.0 to 1.3 throughout, which is the measuring clang itself.
- Seconds are wall clock from `/usr/bin/time -v`, memory is its maximum resident set. Units under 300 KB of C ran three times and the fastest is reported, larger ones once.

## What lowers: nothing of tsc's yet

The meter (`adamic-meter --adapt --json` over the adapted `src/compiler`) examines 79 files: tsc's 77, plus `hostErrors.ts`, which adaptation 47 adds, and `diagnosticInformationMap.generated.ts`, which the upstream build generates. It reports 2 files reaching lowering. Running `adamic c` on every file by itself agrees:

- **None of tsc's 77 files lowers on its own.** 76 stop in the checker and `corePublic.ts` is refused (`corePublic.ts:14:5`, an index signature; the fix is a Map).
- `hostErrors.ts` (adapted tree only, 9 lines) lowers and compiles. It is the only tsc-tree file measured below.
- Every file except `corePublic.ts` and `hostErrors.ts` reaches all 79 files through its imports (71 import `_namespaces/ts`, which re-exports everything). So each one checks the whole compiler, and each first stops at the same place: `builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.` The whole closure has 320 diagnostics: 88 in checker.ts, 62 in sys.ts, 27 in transformers/generators.ts, 17 in builder.ts, and the rest spread out.

### What the builder.ts stop is

The line is `root: arrayFrom(rootFileNames, r => relativeToBuildInfo(r))`, where `rootFileNames` is a `Set<Path>`. It isn't caused by Adamic's stricter checker options (#k881crd):

- Stock tsc 6.0.3 over the adapted tree, with Adamic's options (`strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, and `strictBindCallApply` and `useUnknownInCatchVariables`, which tsc's own base config turns off), reports no diagnostic at builder.ts:1246 in any combination.
- A 12-line repro (`arrayFrom` copied from core.ts, a branded `Path`, a `Set<Path>`) fails under `adamic types` at the same column and passes stock tsc with the same options.
- Adding only the prelude's iterator lines to that stock run (`internal/load/prelude.d.ts`: `AdamicCollectionIteratorResult<T>`, with `done: true; value: undefined`, merged into `SetIterator.next()` and `MapIterator.next()`) reproduces the TS2345 exactly.

So the first stop is the prelude's collection-iterator declaration. Inferring `T` from `Iterable<T>` over a `Set<Path>` picks up the `done: true` arm's `undefined`, which every `for...of` and `arrayFrom` over a Set or Map in tsc will meet. Removing that one stop doesn't make tsc measurable: stock tsc with Adamic's options and no Node types still reports 234 diagnostics on the tree, and 135 with `noUncheckedIndexedAccess` off. Those are #k881crd's. Measuring real tsc needs the iterator declaration fixed, then the remaining option diagnostics (the 87 TS2345, 49 TS18048, 25 TS2412 and 16 TS2375 are, by their codes, most likely `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes`; that's inference) and sys.ts's 46 TS2591 Node-typing gaps adapted or fixed, then whatever lowering finds next.

### Per file

`adamic c` was run on each file alone. The meter's column is the top-ranked reason that names the file in its own-file attribution, with that reason's first site, which can be in another file. "none" means the file has no checker diagnostic of its own and stops only because its imports do.

| File | TS bytes | Lowers alone | First stop (`adamic c`) | Meter's first own-file reason |
|---|---:|---|---|---|
| _namespaces/ts.moduleSpecifiers.ts | 111 | no | builder.ts:1246:69 TS2345 | none |
| _namespaces/ts.performance.ts | 101 | no | builder.ts:1246:69 TS2345 | none |
| _namespaces/ts.ts | 3,288 | no | builder.ts:1246:69 TS2345 | none |
| binder.ts | 195,198 | no | builder.ts:1246:69 TS2345 | none |
| builder.ts | 117,298 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| builderPublic.ts | 11,609 | no | builder.ts:1246:69 TS2345 | none |
| builderState.ts | 26,209 | no | builder.ts:1246:69 TS2345 | none |
| builderStatePublic.ts | 392 | no | builder.ts:1246:69 TS2345 | none |
| checker.ts | 3,154,767 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| commandLineParser.ts | 184,451 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| core.ts | 92,892 | no | builder.ts:1246:69 TS2345 | TS2740, first at core.ts:1637:11 |
| corePublic.ts | 1,320 | no | corePublic.ts:14:5 Refused: index signature | Refused: an index signature, first at corePublic.ts:14:5 |
| debug.ts | 57,783 | no | builder.ts:1246:69 TS2345 | TS2339, first at debug.ts:200:19 |
| diagnosticInformationMap.generated.ts (adapted tree only) | 579,194 | no | builder.ts:1246:69 TS2345 | none |
| emitter.ts | 274,710 | no | builder.ts:1246:69 TS2345 | none |
| executeCommandLine.ts | 54,177 | no | builder.ts:1246:69 TS2345 | none |
| expressionToTypeNode.ts | 70,354 | no | builder.ts:1246:69 TS2345 | none |
| factory/baseNodeFactory.ts | 2,682 | no | builder.ts:1246:69 TS2345 | none |
| factory/emitHelpers.ts | 69,204 | no | builder.ts:1246:69 TS2345 | none |
| factory/emitNode.ts | 12,800 | no | builder.ts:1246:69 TS2345 | none |
| factory/nodeChildren.ts | 2,252 | no | builder.ts:1246:69 TS2345 | none |
| factory/nodeConverters.ts | 7,275 | no | builder.ts:1246:69 TS2345 | none |
| factory/nodeFactory.ts | 336,646 | no | builder.ts:1246:69 TS2345 | none |
| factory/nodeTests.ts | 37,936 | no | builder.ts:1246:69 TS2345 | none |
| factory/parenthesizerRules.ts | 34,072 | no | builder.ts:1246:69 TS2345 | none |
| factory/utilities.ts | 76,635 | no | builder.ts:1246:69 TS2345 | none |
| factory/utilitiesPublic.ts | 2,131 | no | builder.ts:1246:69 TS2345 | none |
| hostErrors.ts (adapted tree only) | 447 | yes | (compiles) | none |
| moduleNameResolver.ts | 183,618 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| moduleSpecifiers.ts | 80,070 | no | builder.ts:1246:69 TS2345 | none |
| parser.ts | 541,260 | no | builder.ts:1246:69 TS2345 | none |
| path.ts | 44,181 | no | builder.ts:1246:69 TS2345 | none |
| performance.ts | 6,161 | no | builder.ts:1246:69 TS2345 | none |
| performanceCore.ts | 3,342 | no | builder.ts:1246:69 TS2345 | none |
| program.ts | 271,357 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| programDiagnostics.ts | 21,401 | no | builder.ts:1246:69 TS2345 | none |
| resolutionCache.ts | 82,223 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| scanner.ts | 219,486 | no | builder.ts:1246:69 TS2345 | none |
| semver.ts | 18,012 | no | builder.ts:1246:69 TS2345 | none |
| sourcemap.ts | 32,210 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| symbolWalker.ts | 8,267 | no | builder.ts:1246:69 TS2345 | none |
| sys.ts | 86,274 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| tracing.ts | 15,679 | no | builder.ts:1246:69 TS2345 | none |
| transformer.ts | 29,875 | no | builder.ts:1246:69 TS2345 | TS2322, first at checker.ts:15145:9 |
| transformers/classFields.ts | 156,676 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/classThis.ts | 5,565 | no | builder.ts:1246:69 TS2345 | none |
| transformers/declarations.ts | 99,818 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/declarations/diagnostics.ts | 46,404 | no | builder.ts:1246:69 TS2345 | none |
| transformers/destructuring.ts | 30,894 | no | builder.ts:1246:69 TS2345 | none |
| transformers/es2015.ts | 231,849 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/es2016.ts | 4,519 | no | builder.ts:1246:69 TS2345 | none |
| transformers/es2017.ts | 50,191 | no | builder.ts:1246:69 TS2345 | TS18048, first at checker.ts:10022:86 |
| transformers/es2018.ts | 69,990 | no | builder.ts:1246:69 TS2345 | none |
| transformers/es2019.ts | 1,563 | no | builder.ts:1246:69 TS2345 | none |
| transformers/es2020.ts | 13,128 | no | builder.ts:1246:69 TS2345 | none |
| transformers/es2021.ts | 4,144 | no | builder.ts:1246:69 TS2345 | none |
| transformers/esDecorators.ts | 127,380 | no | builder.ts:1246:69 TS2345 | TS18048, first at checker.ts:10022:86 |
| transformers/esnext.ts | 32,645 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/generators.ts | 124,156 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/jsx.ts | 36,257 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/legacyDecorators.ts | 36,740 | no | builder.ts:1246:69 TS2345 | none |
| transformers/module/esnextAnd2015.ts | 18,686 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| transformers/module/impliedNodeFormatDependent.ts | 3,750 | no | builder.ts:1246:69 TS2345 | none |
| transformers/module/module.ts | 111,833 | no | builder.ts:1246:69 TS2345 | none |
| transformers/module/system.ts | 85,511 | no | builder.ts:1246:69 TS2345 | none |
| transformers/namedEvaluation.ts | 22,318 | no | builder.ts:1246:69 TS2345 | none |
| transformers/taggedTemplate.ts | 5,489 | no | builder.ts:1246:69 TS2345 | none |
| transformers/ts.ts | 116,761 | no | builder.ts:1246:69 TS2345 | none |
| transformers/typeSerializer.ts | 28,972 | no | builder.ts:1246:69 TS2345 | none |
| transformers/utilities.ts | 35,851 | no | builder.ts:1246:69 TS2345 | none |
| tsbuild.ts | 5,332 | no | builder.ts:1246:69 TS2345 | none |
| tsbuildPublic.ts | 114,671 | no | builder.ts:1246:69 TS2345 | TS2345, first at builder.ts:1246:69 |
| types.ts | 491,975 | no | builder.ts:1246:69 TS2345 | none |
| utilities.ts | 515,367 | no | builder.ts:1246:69 TS2345 | none |
| utilitiesPublic.ts | 104,648 | no | builder.ts:1246:69 TS2345 | none |
| visitorPublic.ts | 88,114 | no | builder.ts:1246:69 TS2345 | none |
| watch.ts | 45,948 | no | builder.ts:1246:69 TS2345 | none |
| watchPublic.ts | 61,201 | no | builder.ts:1246:69 TS2345 | TS2379, first at checker.ts:10941:33 |
| watchUtilities.ts | 35,634 | no | builder.ts:1246:69 TS2345 | none |

## The C, measured

Because no tsc file but a 9-line adaptation compiles, the fit uses the stage1 programs (every `stage1/**/main.ts` outside `gaps/`) as proxies. They are Adamic's own TypeScript, and 22 of 23 compile; `cohere/typeaware` is refused without `--tsgo`. The closest to tsc in shape are **typescript/scanner** and **typescript/parser**, Adamic's ports of tsc's scanner and parser, set in bold. "TS bytes read" is the entry's relative-import closure, which is what the compile reads besides the library. Every program is one translation unit, as the default native build compiles it (`native.go` buildOne). "Split units" is what the opt-in split build (`units.go`) would make of it: `state.c`, one unit per sixteen functions, and `main.c`. The largest function is its name and its lines in the emitted C.

| File | Kind | TS bytes read | C bytes | C lines | Units | Split units | Largest function (lines) | -O0 s | -O1 s | -O2 s | -O0 MB | -O1 MB | -O2 MB |
|---|---|---:|---:|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|
| hostErrors.ts | tsc | 447 | 3,001 | 69 | 1 | 3 | `adamic_function_0_errorCode` (25) | 0.02 | 0.02 | 0.01 | 49 | 53 | 53 |
| cohere/cssstrings | proxy | 4,480 | 24,541 | 448 | 1 | 3 | `adamic_function_0_printString` (124) | 0.03 | 0.10 | 0.10 | 58 | 66 | 67 |
| cohere/cssnumbers | proxy | 11,357 | 58,532 | 1,031 | 1 | 4 | `adamic_function_0_printString` (124) | 0.03 | 0.19 | 0.29 | 59 | 72 | 72 |
| cohere/markdowninline | proxy | 15,210 | 89,103 | 1,474 | 1 | 4 | `main` (373) | 0.04 | 0.23 | 0.27 | 61 | 70 | 72 |
| cohere/mediaquery | proxy | 28,668 | 160,550 | 2,655 | 1 | 4 | `adamic_function_11_parseMediaQuery` (562) | 0.05 | 0.34 | 0.44 | 61 | 77 | 78 |
| cohere/suppression | proxy | 35,645 | 218,767 | 3,904 | 1 | 7 | `main` (323) | 0.06 | 0.45 | 0.50 | 62 | 76 | 76 |
| cohere/lint/helpers | proxy | 28,214 | 310,928 | 5,005 | 1 | 7 | `adamic_function_24_StrictOptions_matches` (857) | 0.09 | 0.73 | 0.86 | 64 | 78 | 77 |
| cohere/selector | proxy | 31,214 | 317,225 | 5,011 | 1 | 6 | `adamic_function_29_Parser_parse` (615) | 0.06 | 0.66 | 0.81 | 65 | 78 | 79 |
| cohere/values | proxy | 49,198 | 359,083 | 5,832 | 1 | 10 | `adamic_function_7_tokenize` (818) | 0.08 | 0.65 | 0.78 | 65 | 77 | 78 |
| cohere/gitignore | proxy | 56,608 | 377,012 | 6,596 | 1 | 10 | `main` (372) | 0.08 | 0.76 | 0.80 | 63 | 80 | 80 |
| cohere/graphql | proxy | 82,340 | 494,252 | 8,701 | 1 | 15 | `adamic_function_59_Parser_parseValueLiteral` (399) | 0.09 | 0.95 | 1.10 | 67 | 81 | 82 |
| cohere/css | proxy | 53,862 | 538,607 | 8,538 | 1 | 9 | `adamic_function_52_Tokenizer_nextToken` (646) | 0.13 | 1.15 | 1.28 | 68 | 86 | 85 |
| cohere/formatfiles | proxy | 78,527 | 545,353 | 9,461 | 1 | 12 | `main` (547) | 0.09 | 1.09 | 1.29 | 66 | 82 | 81 |
| cohere/static_single_assignment | proxy | 90,158 | 583,478 | 8,726 | 1 | 11 | `main` (469) | 0.15 | 1.33 | 1.58 | 68 | 87 | 88 |
| **typescript/scanner** | proxy | 73,768 | 705,330 | 9,709 | 1 | 6 | `main` (4,400) | 0.23 | 1.40 | 1.25 | 83 | 98 | 97 |
| cohere/graphql/printer | proxy | 176,742 | 1,744,917 | 25,134 | 1 | 19 | `main` (12,534) | 0.57 | 13.01 | 13.78 | 121 | 237 | 238 |
| cohere/json | proxy | 132,419 | 1,879,342 | 25,248 | 1 | 7 | `main` (19,857) | 0.83 | 20.16 | 21.39 | 148 | 346 | 345 |
| **typescript/parser** | proxy | 201,428 | 2,005,320 | 27,818 | 1 | 21 | `main` (4,427) | 0.61 | 4.82 | 5.57 | 99 | 136 | 136 |
| cohere/lint/helpers/comments | proxy | 222,220 | 2,219,689 | 31,000 | 1 | 23 | `main` (4,668) | 0.63 | 6.02 | 7.05 | 102 | 140 | 140 |
| cohere/yaml | proxy | 299,803 | 3,371,301 | 47,565 | 1 | 33 | `main` (12,505) | 1.06 | 18.07 | 19.68 | 138 | 261 | 262 |
| cohere/estree | proxy | 429,143 | 4,083,924 | 55,709 | 1 | 39 | `main` (6,405) | 1.00 | 11.91 | 13.13 | 133 | 222 | 222 |
| cohere/tsprinter | proxy | 428,209 | 4,429,107 | 59,096 | 1 | 31 | `main` (19,887) | 1.37 | 32.80 | 35.41 | 177 | 307 | 308 |
| cohere/lint | proxy | 677,284 | 6,958,533 | 96,637 | 1 | 85 | `main` (9,625) | 1.79 | 18.37 | 20.81 | 178 | 274 | 272 |

Totals over the 23 units: 31,477,895 bytes and 445,367 lines of C; 9.1 s at -O0, 135.2 s at -O1 and 148.2 s at -O2, run one after another; peak memory 178 MB at -O0 and 346 MB at -O1 and -O2.

### The slowest unit at -O2

That's `cohere/tsprinter`: 4.4 MB of C, 35.4 s. With `-ftime-trace` (34.5 s in that run; the per-pass split is from a second run at 50 µs granularity, 33.6 s):

| Phase | Seconds |
|---|---:|
| Backend (code generation and the optimizer) | 32.8 |
| Code generation passes | 26.9 |
| Greedy register allocator | 12.1 |
| Live interval analysis | 5.3 |
| Optimizer, middle end | 5.9 |
| Frontend (parsing and semantic analysis) | 0.8 |

By function, 22.4 s of the 26.8 s spent in per-function passes goes to one function: `main`, 19,887 lines, where the emitter puts every module's top-level code. The next largest is `Statements_classDeclaration` at 0.6 s.

## Fits

All fits are least squares in log-log space, y = a·x^b, over the 23 compiled units. The error is the residual standard deviation in natural log: e^s is the typical multiplicative miss, and the ranges below are e^(±2s).

| Fit | a | b (± standard error) | Typical miss | Measured x range |
|---|---:|---:|---:|---|
| C bytes by TypeScript bytes | 2.612 | 1.103 ± 0.031 | ×1.27 | 447 B to 677 KB |
| C lines by TypeScript lines | 5.147 | 0.999 ± 0.037 | ×1.33 | 9 to 16,692 lines |
| -O0 seconds by C bytes | 1.66e-5 | 0.703 ± 0.066 | ×1.75 | 3 KB to 7.0 MB |
| -O1 seconds by C bytes | 2.44e-6 | 1.016 ± 0.063 | ×1.71 | |
| -O2 seconds by C bytes | 1.49e-6 | 1.061 ± 0.056 | ×1.62 | |
| -O2 seconds by TypeScript bytes | 4.72e-6 | 1.158 ± 0.080 | ×1.84 | 447 B to 677 KB |
| -O2 peak memory (KB) by C bytes | 3830 | 0.260 ± 0.039 | ×1.40 | 3 KB to 7.0 MB |

About 10 bytes of C per byte of TypeScript at tsc's file sizes, and C lines track TypeScript lines one for one (about 5 to 1). Clang's time at -O1 and -O2 is close to linear in bytes on average. The miss is wide (×1.6 to ×1.8) because the outliers aren't explained by size at all, see below.

## Projection to tsc's 77 files

tsc's 77 files are 9,433,689 bytes and 192,592 lines of TypeScript; checker.ts alone is 3,154,767 bytes.

| | C bytes | -O0 s | -O1 s | -O2 s | Peak -O2 memory |
|---|---:|---:|---:|---:|---:|
| Each of the 77 files its own unit (seconds summed, serial) | 96 MB (59 to 154) | 17 (5 to 65) | 287 (73 to 1,124) | 359 (106 to 1,218) | 351 MB (179 to 686), for checker.ts |
| All 77 in one unit | 129 MB (80 to 209) | 8 (2 to 28) | 386 (98 to 1,514) | 564 (166 to 1,914) | 480 MB (245 to 940) |

Per file, the projection puts checker.ts at about 39 MB of C, parser.ts, utilities.ts and types.ts at 5.0 to 5.5 MB each, and nodeFactory.ts and emitter.ts at 2.6 to 3.3 MB. 22 of the 77 files project under 100 KB of C (680 KB together), and 46 under 500 KB. The difference between 96 and 129 MB comes from the fit's 1.10 exponent, not from a measured cost of merging.

What the projection can't know:

- It's fitted on Adamic's own programs, not tsc's. The files that don't lower yet may differ in shape: checker.ts is mostly one function (`createTypeChecker`) holding hundreds of nested functions and closures, and nothing measured here is shaped like that. Its C may be nothing like 39 MB, and its largest function nothing like anything below.
- checker.ts is 4.7 times the largest TypeScript closure measured (677 KB), and the whole of tsc is 13.9 times, so both rows are extrapolations.
- Seconds and memory depend most on the largest function (next section), which the fit by size can't see. For tsc it will most likely be module initialization. `diagnosticInformationMap.generated.ts` is a single object literal of 2,131 `diag(...)` calls, and `types.ts` declares 73 enums. If those land in one `main`, it would be larger than any measured here. That's inference; it can't be measured until those files lower.
- Per-file entry builds, as the meter runs them, would each carry the whole 79-file closure. Building 77 entries that way means 77 copies of the 129 MB, not 77 file-sized units. Only per-module emission makes the per-file row real.

## Split-policy facts

1. **Unit size alone doesn't turn clang superlinear, up to 7 MB.** At -O2, units whose largest function is under 10,000 lines cost 1.8 to 3.2 s per MB of C at every size from 0.5 to 7.0 MB: `cohere/lint`, the largest at 7.0 MB, is 3.0 s/MB and 272 MB. Below about 50 KB the per-unit fixed cost shows: about 0.02 s and 50 MB for a 3 KB unit, which is the runtime header (`adamic.h`) parsed again.
2. **The largest function is what turns it superlinear.** Every unit whose `main` is 12,500 lines or more costs 5.4 to 11.4 s/MB at -O2 and uses 238 to 345 MB: graphql/printer (12,534 lines, 7.9 s/MB), yaml (12,505, 5.8), tsprinter (19,887, 8.0), json (19,857, 11.4). At 4,400 to 9,625 lines (scanner, parser, lint/helpers/comments, estree, lint) it stays at 1.8 to 3.2. The cost is register allocation and live-interval analysis on that one function (above). -O0 doesn't show it (0.17 to 0.44 s/MB for every unit over 0.5 MB), so a debug build has no function-size limit to respect.
3. **The largest single function measured is `main`, 19,887 lines** (tsprinter), with json's at 19,857. In every unit over 1.7 MB the largest function is `main`, the module-initialization code of every module the program imports. The largest other function is estree's `Converter_convert`, 4,239 lines, and lint's `RuleSet_new`, 2,794.
4. **Memory follows the largest function, not bytes.** json's 1.9 MB unit takes 345 MB at -O2. lint's 7.0 MB unit takes 272 MB.
5. **Packing tiny modules.** A unit's floor is about 0.02 s and 50 MB, and below about 300 KB the -O2 cost per MB rises from about 2.5 to 4 to 5 s/MB as the fixed part dominates. Grouping to about 0.5 to 1 MB of C per unit keeps that overhead under a few percent: projected, tsc's 22 files under 100 KB of C (680 KB together) would pack into one unit, and its 46 files under 500 KB (7.3 MB together) into about 7.

Recommendation, for the emitter's split policy, not a build:

- Emit each module's top-level initialization as its own function (or several, every few thousand statements), so `main` only calls them. Keep any emitted function under about 8,000 lines of C. That cap is the one with a measured effect: it is the gap between 9,625 lines at 3.0 s/MB and 12,500 lines at 5.4 to 7.9 s/MB.
- Group functions into units of about 1 MB of C, roughly a module or several small ones together, rather than a fixed sixteen functions. Clang's cost per MB is flat from 0.5 to 7 MB, so the size is a choice about parallel compilation and cache granularity, not about clang. At about 1 MB, tsc would make on the order of 100 units at about 3 s each at -O2 on this box.
- Expect checker.ts to need intra-file splitting regardless: at the projected ~39 MB it would be a 2-minute, ~350 MB unit at -O2 even with small functions.

## How to rerun

Linux, from the repository root, with clang 20.1.8 at `/opt/adamic-tools/llvm/bin/clang` (`cloud/setup.sh` installs it) and GNU time at `/usr/bin/time`. Every timing runs one at a time; give the box nothing else to do.

```sh
S=/tmp/tsc-c   # a new scratch directory
git init -q "$S/ts" && git -C "$S/ts" fetch -q --depth 1 https://github.com/microsoft/TypeScript.git 050880ce59e30b356b686bd3144efe24f875ebc8 && git -C "$S/ts" checkout -q FETCH_HEAD
(cd stage3/api && npm ci --ignore-scripts)
ADAMIC_TYPESCRIPT_SOURCE="$S/ts" STAGE3_CACHE="$S/cache" bash stage3/apply.sh "$S/adapted" > "$S/apply.log" 2>&1
go build -o "$S/adamic" ./cmd/adamic && go build -o "$S/adamic-meter" ./cmd/adamic-meter
"$S/adamic-meter" --adapt --json "$S/adapted/src/compiler" > "$S/meter.json"
python3 -I docs/tsc-generated-c/measure.py --adamic "$S/adamic" --clang /opt/adamic-tools/llvm/bin/clang \
    --adapted "$S/adapted" --meter "$S/meter.json" --out "$S/out" > "$S/measure.log" 2>&1
(cd "$S/ts/src/compiler" && find . -name '*.ts' | sed 's|^\./||' | sort) > "$S/out/pristine.txt"
python3 -I docs/tsc-generated-c/report.py "$S/out/rows.json" > "$S/out/report.txt"
```

`measure.py` writes `rows.json` and `files.csv`, and keeps the emitted C under `out/c`. `report.py` writes `summary.json` beside them and prints the tables and fits. The full run took about 14 minutes on this box, after about 3 minutes of apply and 2 of the meter. The raw data from this run (rows.json, summary.json, report.txt, the meter's JSON and the checker probes) is on the branch `evidence/compiler/tsc-generated-c-table`, not in main.
