- Ranked 175 failing callees for the frozen ed6e2975 any-return reason.
- Reconciled 2,060 distinct boundaries, 2,343 observations and 89,627 credited bytes.
- Stock TypeScript disagrees with latent any for 167 callees and 82,337 bytes; these are not labelled source adaptations.
- The three-callee stock-checker fixture passes; misattributing JSON.parse to tsc is caught.
- Exact latent instantiation provenance and production fixes remain outside this report.

[RESULT.json](RESULT.json) contains every callee and every boundary observation.
The top 20 by credited hidden bytes are below. The diagnostic's declaration
location identifies the callee; the boundary location identifies where recovery
stopped. They often differ because lowering enters an imported helper or prepares
a signature. Repetitions across attempts are retained as observations, while
boundary counts deduplicate (file, start, end) per callee. All 2,060 distinct
spans resolve to exactly one callee. Existing outermost-cause byte credit is
preserved from ranking `6c4fc1af`: nested or already examined spans have zero
credit. This is not a sum of full boundary lengths.

The unit branches from main `3ffb1a835184713998a34874e86326cd21db971f` and imports
only frozen measurement evidence, without merging the previous topic branch.
Compiler measurement remains `ed6e29751ee47d86fad450cd1674139883bc0f70`.
All adapted compiler file hashes are verified before inspecting them. Stock
TypeScript 6.0.3 resolves declarations and signatures independently. Its options
match the census's strict ES2024/Bundler settings, with Node typings explicitly
loaded to identify their provenance. This inspection reports 172 stock diagnostics;
these are observations on a rejected program, not a successful compiler build.
External library/Node declaration hashes are retained in inspection.json.

Declaration ownership and any provenance are separate. All failing declarations
are tsc's own. An explicit tsc return annotation is an adaptation candidate;
`tryParseJson` also exposes the library producer JSON.parse. The local ambient
setTimeout declaration is tsc-owned, with no body; it is not relabelled as Node's
typings. No recorded failing signature belongs directly to lib.d.ts or Node.
The fixture nevertheless verifies both external declaration classes.

167 declarations have non-any stock return contracts, such as memoize's `() => T`
and forEach's `U | undefined`. The census signature path calls `l.concrete` on
the checker's return type before printing this reason. The original ledger stores
neither the pre-substitution type nor the mapper nor the transitive call stack.
The source of those any results therefore remains **latent checker/substitution/
context unresolved**, rather than an invented explicit any in tsc. Matching calls
within each failed span carry independently resolved stock return types and any
argument sources. They are syntactic evidence, not a recorded failing invocation;
some boundaries have no direct matching call because preparation or transitive
lowering failed. There are 2,322 observations with matching calls in their spans; only two matching
calls have an any stock result (258 credited bytes). The other 21 observations
have no direct matching call. Fixing this discrepancy requires a targeted
lowering probe.

Totals by declaration owner:

| Owner | Callees | Boundaries | Observations | Credited bytes |
|---|---:|---:|---:|---:|
| tsc | 175 | 2,060 | 2,343 | 89,627 |
| lib.d.ts | 0 | 0 | 0 | 0 |
| Node typings | 0 | 0 | 0 | 0 |

Totals by any origin (library wrapper provenance included):

| Origin | Callees | Boundaries | Observations | Credited bytes |
|---|---:|---:|---:|---:|
| latent context unresolved: stock return is not any | 167 | 2,050 | 2,333 | 82,337 |
| tsc | 7 | 9 | 9 | 7,186 |
| lib.d.ts | 1 | 1 | 1 | 104 |
| Node typings | 0 | 0 | 0 | 0 |

Top 20 callees. Examples name boundaries; the callee declaration is a separate column.

| Callee and declaration | Boundaries | Credited bytes | Boundary file:line examples | Any origin | Existing or proposed return contract |
|---|---:|---:|---|---|---|
| memoize<br>src/compiler/core.ts:1891:1 | 8 | 16,876 | utilities.ts:1384<br>emitter.ts:1209<br>emitter.ts:1206 | latent context unresolved: stock return is not any | () => T |
| forEach<br>src/compiler/core.ts:33:1 | 108 | 11,252 | moduleSpecifiers.ts:464<br>program.ts:344<br>moduleNameResolver.ts:3177 | latent context unresolved: stock return is not any | U &#124; undefined |
| convertToJson<br>src/compiler/commandLineParser.ts:2478:1 | 2 | 4,862 | commandLineParser.ts:2478<br>parser.ts:1616 | tsc | JsonRecoveryValue &#124; undefined |
| filter<br>src/compiler/core.ts:277:1 | 46 | 4,474 | utilities.ts:2774<br>executeCommandLine.ts:517<br>executeCommandLine.ts:525 | latent context unresolved: stock return is not any | readonly T[] &#124; undefined |
| setTextRange<br>src/compiler/factory/utilitiesPublic.ts:10:1 | 136 | 4,215 | transformers/destructuring.ts:451<br>transformers/es2020.ts:250<br>transformers/legacyDecorators.ts:188 | latent context unresolved: stock return is not any | T |
| append<br>src/compiler/core.ts:926:1 | 80 | 4,114 | transformers/module/system.ts:836<br>transformers/module/system.ts:801<br>moduleSpecifiers.ts:532 | latent context unresolved: stock return is not any | T[] &#124; undefined |
| setOriginalNode<br>src/compiler/factory/nodeFactory.ts:7420:1 | 87 | 3,487 | factory/utilities.ts:438<br>factory/utilities.ts:425<br>factory/nodeConverters.ts:105 | latent context unresolved: stock return is not any | T |
| forEachAncestorDirectoryStoppingAtGlobalCache<br>src/compiler/moduleNameResolver.ts:3051:1 | 5 | 2,512 | moduleNameResolver.ts:3029<br>moduleNameResolver.ts:3305<br>moduleNameResolver.ts:796 | latent context unresolved: stock return is not any | T &#124; undefined |
| Debug.checkDefined<br>src/compiler/debug.ts:255:5 | 50 | 2,201 | transformers/declarations/diagnostics.ts:595<br>transformers/destructuring.ts:559<br>emitter.ts:584 | latent context unresolved: stock return is not any | T |
| addRange<br>src/compiler/core.ts:985:1 | 38 | 2,138 | utilities.ts:4676<br>transformers/legacyDecorators.ts:555<br>utilities.ts:4672 | latent context unresolved: stock return is not any | T[] &#124; undefined |
| mapDefined<br>src/compiler/core.ts:494:1 | 13 | 1,747 | moduleSpecifiers.ts:897<br>moduleSpecifiers.ts:1358<br>checker.ts:10263 | latent context unresolved: stock return is not any | U[] |
| cast<br>src/compiler/core.ts:1783:1 | 19 | 1,735 | transformers/utilities.ts:304<br>transformers/es2020.ts:92<br>commandLineParser.ts:2207 | latent context unresolved: stock return is not any | TOut |
| toSorted<br>src/compiler/core.ts:1038:1 | 6 | 1,670 | utilities.ts:9853<br>utilities.ts:9876<br>executeCommandLine.ts:179 | latent context unresolved: stock return is not any | SortedReadonlyArray<T> |
| convertConfigFileToObject<br>src/compiler/commandLineParser.ts:2437:1 | 2 | 1,545 | commandLineParser.ts:2437<br>commandLineParser.ts:2279 | tsc | JsonRecoveryObject |
| getFirstJSDocTag<br>src/compiler/utilitiesPublic.ts:1279:1 | 21 | 1,431 | utilitiesPublic.ts:1206<br>utilitiesPublic.ts:1225<br>utilitiesPublic.ts:1176 | latent context unresolved: stock return is not any | T &#124; undefined |
| visitEachChild<br>src/compiler/visitorPublic.ts:598:1 | 170 | 1,411 | transformers/module/system.ts:1543<br>transformers/module/system.ts:1503<br>transformers/module/system.ts:1512 | latent context unresolved: stock return is not any | not evident |
| setEmitFlags<br>src/compiler/factory/emitNode.ts:92:1 | 69 | 1,207 | transformers/module/system.ts:235<br>transformers/module/system.ts:1218<br>transformers/legacyDecorators.ts:727 | latent context unresolved: stock return is not any | T |
| setParent<br>src/compiler/utilities.ts:10703:1 | 23 | 1,206 | transformers/typeSerializer.ts:600<br>factory/utilities.ts:354<br>factory/utilities.ts:211 | latent context unresolved: stock return is not any | T &#124; undefined |
| flatMap<br>src/compiler/core.ts:399:1 | 10 | 1,008 | utilities.ts:4686<br>moduleSpecifiers.ts:1176<br>utilities.ts:9675 | latent context unresolved: stock return is not any | readonly U[] |
| memoizeOne<br>src/compiler/core.ts:1907:1 | 1 | 992 | moduleSpecifiers.ts:133 | latent context unresolved: stock return is not any | (arg: A) => T |

The eight actual-any tsc declarations follow. These candidates come from inspected
bodies and are separate from recorded checker types. They require contract and
call-site validation before implementation; no source annotation was changed.
JSON recovery objects can contain undefined-valued properties on invalid input.
Use `JsonRecoveryValue = string | number | boolean | null | JsonRecoveryValue[] |
JsonRecoveryObject` and `JsonRecoveryObject = { [key: string]: JsonRecoveryValue |
undefined }`. For strict JSON.parse without a reviver, `JsonValue` recursively
contains only JSON primitives, arrays and objects; it excludes undefined.

| Callee | Bytes | Candidate and body evidence |
|---|---:|---|
| convertToJson<br>src/compiler/commandLineParser.ts:2478:1 | 4862 | JsonRecoveryValue &#124; undefined: Absent root yields {} or undefined; other roots delegate to the recursive JSON syntax visitor. |
| convertConfigFileToObject<br>src/compiler/commandLineParser.ts:2437:1 | 1545 | JsonRecoveryObject: Invalid non-object roots return {}; array recovery selects an object; absent roots yield {}; valid object roots recursively construct an object. returnValue is true. |
| objectAllocator.getNodeConstructor<br>src/compiler/utilities.ts:8553:25 | 540 | typeof Node: Expression body returns the local Node constructor through an explicit as any assertion. Preserve the constructor contract instead of that assertion. |
| convertToObject<br>src/compiler/commandLineParser.ts:2467:1 | 150 | JsonRecoveryValue &#124; undefined: Delegates to convertToJson with returnValue true; arbitrary recovered root expressions can be invalid. |
| tryParseJson<br>src/compiler/utilities.ts:7809:1 | 104 | JsonValue &#124; undefined: JSON.parse(text) has no reviver and returns a JSON value; the catch returns undefined. External any producer is lib.es5.d.ts JSON.parse. |
| setTimeout<br>src/compiler/sys.ts:51:1 | 89 | not evident: Ambient tsc-local host shim has no body. A universal opaque timer handle or a host-specific handle needs a matching clearTimeout contract; Node's declaration was not the failing declaration. |
| convertToJson.convertObjectLiteralExpressionToJson<br>src/compiler/commandLineParser.ts:2491:5 | 0 | JsonRecoveryObject &#124; undefined: Returns the {} or undefined accumulator. Object fields receive recursively converted values, including undefined for invalid property values. |
| convertToJson.convertPropertyValueToJson<br>src/compiler/commandLineParser.ts:2538:5 | 0 | JsonRecoveryValue &#124; undefined: Switch returns booleans, null, strings, numbers, recursively constructed objects or arrays; invalid expressions return undefined. Array conversion filters undefined entries. |

Reproduce using the frozen adapted tree from the hidden-source measurement:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/any-returns-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
python3 stage3/census/any-returns/test_fixture.py > /tmp/any-returns-fixture.log 2>&1
node stage3/census/any-returns/inspect.cjs corpus /tmp/hidden-adapted stage3/census/any-returns/evidence/boundaries.json stage3/census/any-returns/evidence/inspection.json > /tmp/any-returns-inspection.log 2>&1
python3 stage3/census/any-returns/report.py > /tmp/any-returns-report.log 2>&1
```

The three-callee fixture is committed as `.a.txt` and copied to a scratch `.a`.
Stock TypeScript accepts it with zero diagnostics. Its known answer is own -> tsc,
JSON.parse -> lib.d.ts and require -> Node typings; own has an explicit any return
but its literal body supports number. Counts are one boundary per callee, with
5 own-call bytes, 16 JSON.parse-call bytes and 20 require-call bytes, total 41.
These are synthetic full call-span credits, not a new latent census run.
The lib-to-tsc mutant changes the real
classifier's result for JSON.parse. It still has zero stock diagnostics and is
caught by `three-callee declaration provenance`, not compilation failure.
Fixture and mutant artifacts and logs are in evidence/.

No native/oracle fixture or counts.md changed. No whole packages, gate or compiler
semantic oracle were run. We did not identify an exact source of any for the
167 stock-signature discrepancies, or prove proposed contracts against all callers.
Setup cumulative timing: Go 0.016s, Node 0.018s, markdown 0.048s, submodules 0.059s,
clang 0.117s, Go build 32.855s, test binaries deferred 32.928s, cache warm 32.929s,
done 32.951s. `nproc` is 5; CPU quota is 4.
