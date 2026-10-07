# Checker 259 ledger

**1. Project errors: 0. 2. Adamic-only errors: 258. 3. Stage-0-only errors: 1.**

| Census column | Rows | Disposition under the October 7, 17:35 ruling |
|---|---:|---|
| 1. Error under tsc's own tsconfig | **0** | No adaptation families or declaration fixes to propose |
| 2. Error under Adamic's checker inputs, absent under tsc's own tsconfig | **258** | Runtime-check candidates, with declaration-driven differences distinguished below |
| 3. Only stage 0 diagnoses the site | **1** | Repeated destructuring diagnostic difference, with a minimal witness |
| Stock-only, with the same Adamic inputs | **0** | [stock-only.csv](stock-only.csv) is an empty list with a header |

These counts use **the census loader's actual options and declaration inputs**. Options alone are not enough to reproduce the census. With unmodified stock declarations, the split is 0 / 171 / 88; 87 of those 88 are reproduced by stock tsc when Adamic's declarations are supplied. They are therefore not typescript-go checker differences. `rows.csv` records both observations, including full messages, source expressions, owners, and option-ablation evidence. Each of the original 259 rows appears exactly once in its original order.

## Inputs and checker runs

Adaptation base and worker branch base: `234ab1aa5f728a5221fb6075c35b94f88a2c6437`, origin/area/stage3. Upstream TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`, v6.0.3. Census: `5fd6445e`, `stage3/interface-downcasts/lazy/census/adapted-diagnostics.csv`. Node: **24.19.0**. Applied all registered adaptations: 78 files, 5,122 lines added and 5,096 removed. **All 79 compiler source SHA-256 hashes equal the census hashes.**

The project run parses `src/compiler/tsconfig.json` and its unchanged `../tsconfig-base`, including NodeNext, ES2020, composite, declaration-only emission, isolatedDeclarations, skipLibCheck, strict mode, `strictBindCallApply: false`, and `useUnknownInCatchVariables: false`. No project option is overridden for this run. The measurement uses stock tsc 6.0.3's `createProgram` and `getPreEmitDiagnostics`, including declaration diagnostics.

The comparable Adamic run uses the effective `compilerOptions()` at the census revision: strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess, verbatimModuleSyntax, allowImportingTsExtensions, noEmit, forced modules, ESNext/Bundler, ES2024, es2024 lib, and `types: []`. `noImplicitReturns` and `noFallthroughCasesInSwitch` are absent. **erasableSyntaxOnly is false in that production loader**, although adaptation 20's script still sets it true. Its old two house-style options are omitted in every stock Adamic run.

The census loader also explicitly roots **@types/node 25.3.3**, merges its prelude's console/process declarations with Node, adds the NodeJS.Require refinement, and overlays regex library signatures. These declarations are presented through a read-only stock compiler host. No compiler source, adaptation, or installed TypeScript library is changed. The Node prelude and Require declaration snapshots are [prelude-node.txt](evidence/prelude-node.txt) and [node-require.txt](evidence/node-require.txt), transcribed from `internal/load/node_library.go` and `node_require.d.ts` at `5fd6445e`. The regex host transformation follows that revision's `regexp_library.go`.

| Stock checker run | Full diagnostics | Census sites matched | Additional stock sites | Full list |
|---|---:|---:|---:|---|
| Exact project tsconfig | **0** | 0 | 0 | [stock-own.csv](evidence/stock-own.csv) |
| Effective Adamic options and census declarations | **258** | **258** | **0** | [stock-census-inputs.csv](evidence/stock-census-inputs.csv) |
| Project inputs plus strictBindCallApply, unknown catches, unchecked indexes and exact optionals | 171 | 171 | 0 | [stock-project-stricter.csv](evidence/stock-project-stricter.csv) |
| Effective Adamic options, unmodified stock declarations | 233 | 171 | 62 | [stock-effective.csv](evidence/stock-effective.csv) |
| Adaptation-script options, unmodified declarations, erasableSyntaxOnly true | 413 | 171 | 242 | [stock-adamic.csv](evidence/stock-adamic.csv) |

[options.json](evidence/options.json) records the complete parsed options and roots, not just selected flags. The unmodified-declaration extras are separately listed in [stock-only-unmodified.csv](stock-only-unmodified.csv): 52 TS2591, 6 TS2304, and one each TS2584, TS2345, TS7006, TS7031. These come from missing ambient Node/console inputs and their downstream typing consequences. The literal adaptation-script run adds 180 TS1294 syntax restrictions. They are not extra sites in the apples-to-apples production comparison, and TS1294 is not an inserted runtime check.

Site matching is by exact relative file, UTF-16 line/column, and diagnostic code. All original census messages are retained; the matching stock message is stored alongside them. Full messages are identical on 228 of 258 matched sites; the other 30 have elaboration, formatting, or type-display differences. Matching does not depend on normalizing away message contents.

## Column 1: adaptation families and owners

**No rows.** The complete adapted compiler checks under its own project configuration. There is consequently no family 30-33, 20/75, 47, 40, or new adaptation family to assign, and no declaration-level fix or hidden-presence check to recommend in this column. `family` and `check_location` remain empty in `rows.csv`. The revised brief limits that review to column 1; presence proofs for runtime-check candidates were not audited.

## Column 2: attribution by option and kind

Attribution comes from disabling one option at a time while retaining all census declarations. Membership at the exact diagnostic site matters: total-count subtraction alone would miscount replacement diagnostics. The primary attribution is indexed read, then exact optional contract, then unknown catch; otherwise it is `other`.

| Primary option/input | Indexed read | Optional write | Optional read | Catch variable | Other | Total |
|---|---:|---:|---:|---:|---:|---:|
| noUncheckedIndexedAccess | 99 | 0 | 0 | 0 | 0 | **99** |
| exactOptionalPropertyTypes | 0 | 66 | 0 | 0 | 1 | **67** |
| useUnknownInCatchVariables | 0 | 0 | 0 | 5 | 0 | **5** |
| other: Adamic declaration contracts | 82 | 0 | 0 | 0 | 5 | **87** |
| **Total** | **181** | **66** | **0** | **5** | **6** | **258** |

“Optional write” includes writing an object value through an exact optional target contract. The one exact-optional `other` row is the `SymbolTrackerImpl` implements relation, TS2420 at checker.ts:54329. The five unknown-catch rows are four member accesses and the returned catch error in sys.ts's `require` implementation. The two TS18046 rows in JSX arise from failed tuple destructuring and downstream generic inference, not catch variables.

checker.ts:43517:20, TS2412, is removed by **both** noUncheckedIndexedAccess and exactOptionalPropertyTypes ablations. Its primary category is indexed read; `removed_by` records both. Thus the exact optional option independently removes 68 sites, although 67 have it as primary cause.

| Ablation | Diagnostics left | Baseline sites removed | Newly reported sites |
|---|---:|---:|---:|
| noUncheckedIndexedAccess false | 159 | 99 | 0 |
| exactOptionalPropertyTypes false | 192 | 68 | 2 |
| useUnknownInCatchVariables false | 254 | 5 | 1 |
| strictBindCallApply false | 258 | 0 | 0 |

Each ablation's full list is saved under `evidence/stock-census-without-*.csv`.

The **87 other-input rows** need care before translating counts into runtime checks:

| Adamic declaration source | Rows | What stock proves when given that declaration | Minimal input-difference witnesses |
|---|---:|---|---|
| MapIterator/SetIterator.next completion-value overload | 84 | Generic Iterable element inference includes undefined; downstream arrays, callbacks and tuple destructuring reflect that type | [iterable.a](witnesses/iterable.a), [callback.a](witnesses/callback.a), [entries.a](witnesses/entries.a), [set-copy.a](witnesses/set-copy.a), [nested-entries.a](witnesses/nested-entries.a) |
| JSON.stringify returns string or undefined | 2 | A string-only return or write is not satisfied | [json.a](witnesses/json.a) |
| Seven ES2025 Set methods added to es2024 Set | 1 | The custom Set-shaped object omits the added methods | [set-shape.a](witnesses/set-shape.a) |

Observation: stock and stage 0 both report the linked declaration witnesses with Adamic's prelude, while stock with its original declarations accepts them. Inference: the collection rows are completion-value pollution of the generic Iterable element type, not proof that a built-in iterator yields its terminal undefined value. The Set row describes a shape contract. These are `other` checker-input differences, not effects of a stricter option flag, and this ledger does not assert that each diagnostic demands its own emitted runtime check.

## Indexed reads, separately

[indexed-reads.csv](indexed-reads.csv) contains all **181 matched indexed-read diagnostics**, with their full row evidence. Of these, 99 originate in array/tuple/index-signature reads and propagated locals under noUncheckedIndexedAccess; 82 originate in the Adamic collection-iterator contract, including Map/Set keys, values, entries, callback values, and copied Sets. The remaining two collection-origin diagnostics are unknown generic callback parameters and are counted as `other` above. The sole column-3 duplicate tuple diagnostic is separate.

Examples: `elementTypes[unionIndex]` at checker.ts:17946; `targets[i]` at checker.ts:20595; `map.sourcesContent[raw.sourceIndex]` at sourcemap.ts:216; `caseBlock.clauses[i]` at generators.ts:1919. Collection-origin examples include `arrayFrom(state.fileInfos.entries(), ...)` in builder.ts and `arrayFrom(props.values())` in checker.ts. Diagnostic counts include propagated uses, not just original reads.

| Owning compiler file | Unchecked index rows | Collection-contract rows | Total |
|---|---:|---:|---:|
| `builder.ts` | 0 | 14 | 14 |
| `checker.ts` | 27 | 49 | 76 |
| `commandLineParser.ts` | 1 | 4 | 5 |
| `moduleNameResolver.ts` | 1 | 0 | 1 |
| `program.ts` | 3 | 2 | 5 |
| `resolutionCache.ts` | 0 | 5 | 5 |
| `sourcemap.ts` | 1 | 0 | 1 |
| `transformer.ts` | 6 | 0 | 6 |
| `transformers/classFields.ts` | 6 | 0 | 6 |
| `transformers/declarations.ts` | 1 | 0 | 1 |
| `transformers/es2015.ts` | 5 | 0 | 5 |
| `transformers/es2017.ts` | 6 | 2 | 8 |
| `transformers/esDecorators.ts` | 3 | 0 | 3 |
| `transformers/esnext.ts` | 5 | 1 | 6 |
| `transformers/generators.ts` | 27 | 0 | 27 |
| `transformers/jsx.ts` | 2 | 1 | 3 |
| `transformers/module/esnextAnd2015.ts` | 5 | 0 | 5 |
| `tsbuildPublic.ts` | 0 | 4 | 4 |
| **Total** | **99** | **82** | **181** |

## Column 3: the one checker difference

| Row | Site | Owner | Cause | Witness |
|---|---|---|---|---|
| D003 | builder.ts:1292:61, TS2488 | `getBuildInfo.fileInfos` | Repeated undefined-tuple iteration failure is reported again by stage 0 and suppressed at the repeated site by stock tsc | [repeated-destructure.a](witnesses/repeated-destructure.a) |

The whole-tree stock run already diagnoses the same Map-entry tuple type at builder.ts:1258:65. At 1292 it still types the destructuring parameter as `[Path, FileInfo] | undefined`, but reports no second TS2488. A minimal witness isolates this without changing any library input:

```ts
declare function transform<T, U>(items: Iterable<T>, callback: (item: T) => U): U[];
declare const values: Iterable<[string, number] | undefined>;
// @ts-expect-error The first destructuring diagnoses the undefined tuple.
transform(values, ([key, value]) => value);
transform(values, ([key, value]) => value);
```

The file is stored as `.a`; stock's read-only host presents it as a virtual TypeScript source. **Stock tsc 6.0.3: zero diagnostics. Stage 0: TS2488 at line 5, column 20.** The expected-error directive suppresses only the first diagnostic, which both checkers would otherwise emit. It is a diagnostic-isolation probe, not a source adaptation or a bypass of the production census. A reduction using only the second call produces TS2488 in both checkers; the repeated-site difference requires the first call.

Observed behavior supports iteration-diagnostic caching as the cause. The stock compiler implementation was not modified or instrumented to prove its internal cache key. A typescript-go semantic rejection is therefore not inferred from a duplicated diagnostic alone.

## Census by diagnostic code

| Code | Column 1 | Column 2 | Column 3 | Census total |
|---|---:|---:|---:|---:|
| TS2345 | 0 | 87 | 0 | 87 |
| TS18048 | 0 | 49 | 0 | 49 |
| TS2322 | 0 | 30 | 0 | 30 |
| TS2412 | 0 | 25 | 0 | 25 |
| TS2375 | 0 | 16 | 0 | 16 |
| TS2532 | 0 | 15 | 0 | 15 |
| TS2379 | 0 | 10 | 0 | 10 |
| TS2488 | 0 | 5 | 1 | 6 |
| TS2769 | 0 | 6 | 0 | 6 |
| TS18046 | 0 | 6 | 0 | 6 |
| TS2339 | 0 | 3 | 0 | 3 |
| TS2538 | 0 | 2 | 0 | 2 |
| TS2722 | 0 | 1 | 0 | 1 |
| TS2556 | 0 | 1 | 0 | 1 |
| TS2420 | 0 | 1 | 0 | 1 |
| TS2740 | 0 | 1 | 0 | 1 |
| **Total** | **0** | **258** | **1** | **259** |

## Reproduction, controls and limits

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/checker-ledger-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/checker-259-adapted > /tmp/checker-ledger-apply.log 2>&1
npm ci --prefix /tmp/checker-259-adapted --ignore-scripts --no-audit --no-fund > /tmp/checker-ledger-npm.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/ledger/checker-259/measure.cjs /tmp/checker-259-adapted /tmp/checker-ledger-reproduction > /tmp/checker-ledger-reproduction.log 2>&1
python3 stage3/ledger/checker-259/validate.py --tree /tmp/checker-259-adapted --mutants > /tmp/checker-ledger-validation.log 2>&1
go build -o /tmp/checker-ledger-load ./stage3/ledger/checker-259/probe-stage0.go > /tmp/checker-ledger-load-build.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/ledger/checker-259/witnesses.cjs > /tmp/checker-ledger-final-stock-witnesses.jsonl 2>&1
/tmp/checker-ledger-load stage3/ledger/checker-259/witnesses/*.a > /tmp/checker-ledger-final-stage0-witnesses.jsonl 2>&1
```

Use any installed stage3/api node_modules directory with the exact pins for NODE_PATH; the absolute path above records this run. `measure.cjs` refuses an existing output directory. The independent reproduction matches **every diagnostic site and full message in all nine retained measurement modes**. Stock witness checks use the same virtual `.a` source view. [final-stock-witnesses.jsonl](evidence/final-stock-witnesses.jsonl) and [final-stage0-witnesses.jsonl](evidence/final-stage0-witnesses.jsonl) retain the final 14-witness suite, including the on/off option controls and single-destructuring reduction. Earlier raw witness logs are also retained.

| Check or mutant | Observed catcher |
|---|---|
| String assigned to number, `type-error.a` | Stock and stage 0 TS2322 |
| Empty-array indexed read, `index-read.a` | Stock and stage 0 TS2322 with unchecked indexes; disappears when that option is off |
| Present undefined optional write, `optional-write.a` | Stock and stage 0 TS2375 with exact optionals; disappears when that option is off |
| Unknown catch member access, `catch.a` | Stock and stage 0 TS18046 with unknown catches; disappears when that option is off |
| Delete a census diagnostic | Coverage-count assertion |
| Put a runtime row into adaptation column 1 | Independent checker-membership assertion |
| Attribute an indexed read to exact optional properties | Independent option-ablation assertion |
| Duplicate one diagnostic site in place of another | Uniqueness assertion |
| Replace one recorded adapted source hash | SHA-256 identity assertion |

The five ledger mutants were all caught. The number-literal green control checks under both checkers; changing its value to a string in `type-error.a` is caught by TS2322 in both. Checker controls change real source input; the stock option controls were run both on and off. The repeated-destructuring control with the first call removed is diagnosed by both checkers. Logs are retained in `evidence/validation.log.gz` and the witness logs.

Setup succeeded: Node ready 0.069s, Go 0.096s, clang 0.506s, markdown dependencies 1.154s, submodules 200.880s, Go build 602.558s, cache warm 602.710s, done **602.796s**. `nproc=5`, cgroup CPU quota 4, Go 1.27.1, clang 20.1.8, Node 24.19.0. [setup.log](evidence/setup.log.gz) contains all timing lines. Upstream npm installed 338 packages in 6 seconds. No setup failure or workaround was needed.

`go vet ./stage3/ledger/checker-259` and `git diff --check` both exit 0 with no output. The probe was compiled with `go build`; no lowerer or emitter was invoked.

The saved 259-row census is the stage-0 reference. A fresh area/stage3 stage-0 whole-tree run reports 319 rows: 258 reference sites plus 61 ambient/dependent errors, missing the reference JSON.stringify-to-fs-write site. This older base does not yet contain the census revision's Node loader integration; [area-stage0-whole.jsonl](evidence/area-stage0-whole.jsonl) records that observation. The portable stock host reproduces the census revision's documented declaration inputs instead of changing this branch's compiler.

Covered: all 259 census rows, all 79 adapted input hashes, exact project checking, production-equivalent stock checking, individual option attribution, source/owner evidence, declaration witnesses, one minimal checker difference, and evidence-corruption mutants. Not covered: emitted runtime checks, whole-program lowering/emission, native execution, reachability, presence-check elision proofs, or a new full repository oracle gate. This unit changes only its ledger directory and adds no runtime implementation.
