# tsc against its declared ES2020 lib

The seven ES2025 Set operations come from Adamic's prelude, not from ES2024 or Node's lib references. Selecting the ES2020 lib and target leaves the meter's diagnostics and all latent sites unchanged. TS2740 at `core.ts:1637:11` remains.

| Configuration | Aggregate diagnostics | Own-file sum | Distinct NotYet | Distinct Refused |
| --- | ---: | ---: | ---: | ---: |
| Main | 144 | 145 | 652 | 2115 |
| ES2020 lib and target | 144 | 145 | 652 | 2115 |
| ES2020, capped Node references | 179 | 144 | 652 | 2115 |

Capping Node's two later-lib references adds 36 declaration errors and masks one source error. It also changes three NotYet reason rows, replacing two sites with two other sites at `sys.ts:51-52`; Refused sites and unit eligibility remain identical. This is not a checker improvement.

Measurement only. No production source or adaptation change is included in this commit.

Pinned main: `7b9c0fe7e52cfe9cdeb774107389bd6870f41423`.
Pinned cohere: `06af2e399c6c1fb93fd570b774027942747400b4`.
Pinned TypeScript-Go: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
Adapted upstream TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`, tag `v6.0.3`.

## Where Set comes from

Adamic's own prelude supplies the seven ES2025 Set methods. This is independent of the selected lib and independent of Node's lib references. Changing the lib to ES2020 leaves those methods in place.

| Declaration | File and line in pinned sources | Contribution |
| --- | --- | --- |
| Set | `cohere/TypeScript/tsc/internal/bundled/libs/lib.es2015.collection.d.ts:90` | add, clear, delete, forEach, has, size |
| Set | `cohere/TypeScript/tsc/internal/bundled/libs/lib.es2015.iterable.d.ts:194` | Symbol.iterator, entries, keys, values |
| Set | `cohere/TypeScript/tsc/internal/bundled/libs/lib.es2015.symbol.wellknown.d.ts:142` | Symbol.toStringTag |
| Set | `internal/load/prelude.d.ts:62` | union at 63, intersection at 64, difference at 65, symmetricDifference at 66, isSubsetOf at 67, isSupersetOf at 68, isDisjointFrom at 69 |
| ReadonlySet | `internal/load/prelude.d.ts:71` | Same seven methods, lines 72 through 78 |
| ReadonlySetLike | `internal/load/prelude.d.ts:95` | Set-operation argument shape |
| SetIterator | `internal/load/prelude.d.ts:91` | next overload with a required done discriminant |

The checker's `lib.es2024.d.ts:17-24` references ES2023 and seven ES2024 fragments. None introduces these Set operations. Its `lib.es2025.collection.d.ts:34` does declare them, but embedding a file does not load it. The actual loaded-file and declaration probe is retained alongside the results.

Node is pinned to `@types/node` 25.3.3 by `internal/load/node_library.go:19`. Its `internal/load/node_types/node_modules/@types/node/index.d.ts:28-30` references `es2020`, `esnext.disposable`, and `esnext.float16`. `cohere/TypeScript/tsc/internal/tsoptions/declarations_generated.go:1281` maps `esnext.float16` to `lib.es2025.float16.d.ts`. That fragment references ES2015 symbol and iterable at lines 17 and 18. `lib.esnext.disposable.d.ts:17-19` references ES2015 symbol, ES2015 iterable, and ES2018 asynciterable. Neither path loads ES2025 collection operations.

## Loader behavior and scratch changes

`internal/load/load.go:57-75` defines one fixed configuration. It does not read a program's tsconfig. At lines 127-138 it adds the prelude, wraps the filesystem with the bundled libraries, and constructs a parsed command line directly from those options and the supplied roots. At lines 143-160 it detects Node imports and rebuilds the program with the embedded Node index as an additional root, again using the same fixed options. `Types: []string{}` does not suppress those explicit roots or triple-slash lib references.

The adapted `src/tsconfig-base.json:7-8` declares `lib: ["es2020"]` and `target: "es2020"`. `es2020.patch` changes only those two loader settings, in a detached scratch worktree. It matches the declared lib and target; it does not import the other tsconfig options. The loader has no caller option for choosing a lib. The prelude, strict settings, Node declarations, filesystem corrections, and adaptations remain in effect.

`es2020-capped-node.patch` is a separate sensitivity measurement. In addition to the two settings, it suppresses only the two post-ES2020 triple-slash references while serving the embedded Node files. It supplies no replacement declarations or stubs. It does not remove the prelude's Set augmentation.

## Reproduction and counting

Setup was invoked with `GOPROXY='https://proxy.golang.org|direct'` and `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`.

The whole `stage3/meter/twice-daily.sh` was read before preparing the tree. One adapted tree was built with `stage3/apply.sh` from the exact `stage3/source.json` pin, after lockfile installation of stage3/api dependencies. Every variant uses that same tree on the same box. The runner uses the meter's ordinary census, latent overlay, and entry overlay, with `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`. The entry census is retained even when checker errors block entry lowering. Compiler-directory latent observations are explicitly measurements on rejected programs, not evidence that tsc can compile.

`run.sh`, `analyze.py`, the probe source and runner, patches, source hashes, box information, setup/apply/build logs, raw census data, full diagnostic sets, and latent changes are retained. Binary hashes identify local executables; executables themselves are not committed. Scratch paths in scripts document this run and can be relocated together. Data use the same adapted-tree absolute path for all variants, avoiding path-based differences.

The Node-adjusted virtual prelude is served as `/adamic-prelude/adamic.d.ts`. Replacing its four-line console declaration with one line shifts subsequent virtual positions by three lines. The actual-program probe therefore reports the Set interface at virtual line 59 and union at virtual line 60; the corresponding checked-in source positions are `internal/load/prelude.d.ts:62` and `:63`. The merged symbol also includes the Set constructor variable at `lib.es2015.collection.d.ts:122:13`; that variable adds no instance methods.

## Prelude and runtime declarations

The prelude exposes ES2025 Set operations, but declares both their signatures and `ReadonlySetLike` itself. It has no unresolved declaration dependency on a post-ES2020 lib. Its `MapIterator` and `SetIterator` overloads use `IteratorReturnResult`, supplied by the checker's ES2015 iterable lib even when the selected aggregate is ES2020. A modern TypeScript declaration name in an ES2015 fragment does not imply that an ES2025 aggregate was loaded. The prelude-only probe checks clean under main and ES2020.

The pinned Node runtime declarations need post-ES2020 library declarations: `Disposable`, `AsyncDisposable`, `Symbol.dispose`, `Symbol.asyncDispose`, and `Float16Array`. Examples include `@types/node/fs.d.ts:2155`, `fs/promises.d.ts:982`, `readline.d.ts:62`, `timers.d.ts:34` and `:86`, and `globals.typedarray.d.ts:15` and `:74`. The ES2020 option-only patch preserves Node's references, so these names remain defined. The minimal Node-import probe checks clean under main and ES2020. The capped-reference measurement below quantifies what happens when those declarations are withheld without stubbing them.

These are checker and latent-lowering measurements. They do not assert that every runtime operation supported by Adamic remains expressible under ES2020, or that all runtime implementations have been tested. Selecting ES2020 removes later standard-library API declarations from user programs even when the runtime implements those APIs.

The capped-reference entry probe adds 36 Node diagnostics: 13 TS2304 and 23 TS2550. It also loses the existing TS2345 at `src/compiler/sys.ts:1708:56`, where the prelude's `JSON.stringify` may return undefined and the result is passed to `fs.writeFileSync`. Node's `globals.typedarray.d.ts:5-20` builds `ArrayBufferView` through a union containing the now-missing `Float16Array`. Losing that declaration damages the parameter type and masks this source error. The resulting net diagnostic increase is 35. This is a declaration failure, not evidence of improved checking. The Set diagnostic remains.

See [the before/after tables and full diagnostic changes](tables.md), [the loaded-library probe tables](probe-tables.md), and [the machine-readable summary](summary.json). Full before/after diagnostic sets are retained per variant, including errors that do not change. Full latent added/removed sites are retained in the `*-latent-changes.json` files, with every changed reason row and examples in the tables.

The ordinary diagnostic comparisons are exact sets, not only matching totals: ES2020 has zero additions and zero removals in both the compiler aggregate and the own-file union. The capped-reference case removes one own-file diagnostic and adds no own-file diagnostics; its 36 additions are external Node declaration diagnostics.

For scheduling, the baseline started first. After its latent pass began, the remaining variants ran concurrently on the same box, preserving the active baseline process. `run.sh` records the same build and measurement commands in a sequential form suitable for reproduction. The adapted source hash manifest is checked again when packaging the results.

The minimal Node probe's 36 capped-reference errors break down as follows. These names account for all its missing post-ES2020 declarations.

| Missing declaration | Code | Sites |
| --- | --- | ---: |
| Disposable | TS2304 | 7 |
| AsyncDisposable | TS2304 | 3 |
| Float16Array | TS2304 | 3 |
| Symbol.dispose | TS2550 | 11 |
| Symbol.asyncDispose | TS2550 | 12 |

While the latent passes ran concurrently, the capped-reference pass was paused with its accumulated state intact to preserve memory headroom. It resumed after baseline completion. Scheduling and elapsed time are not interpreted as a performance comparison.

The diagnostic present in the own-file union but absent from the compiler aggregate is TS2322 at `src/compiler/commandLineParser.ts:2682:9`, involving `originalPath: undefined` versus the exact optional `ProjectReference.originalPath: string`. It is present in every variant's per-file observations and explains the one-diagnostic difference between main's 145 own-file sum and 144 aggregate diagnostics.

The two-line patch is a measurement configuration, not a per-project configuration implementation: it changes the loader's global defaults in the scratch compiler. Landing it verbatim would select ES2020 for all programs. The measurements here cover tsc and the declaration probes; they do not claim that unrelated programs using later library APIs would still check.

Latent totals in the comparison table refer to the meter's compiler-directory census over 79 source files. Every tsc-entry first-error lowering census is blocked by its checker diagnostics, as in the meter. The entry checker census and the compiler aggregate each report 144 diagnostics on main and ES2020, and 179 in the capped-reference case.

Node's disposable lib also changes returned iterator types: `lib.esnext.disposable.d.ts:187` makes `IteratorObject` extend `Disposable`. This affects `SetIterator` indirectly through the ES2015 iterable declaration. It does not add union, intersection, or the other Set operations to Set itself.

The option-only ES2020 latent result is also an exact match: zero distinct sites are added or removed, not only matching NotYet and Refused totals. Main and ES2020 each have 4,017 attempted units, 498 deferred generic-instantiation units, and 28 checker-skipped bodies. No reason row changes between these two configurations.
