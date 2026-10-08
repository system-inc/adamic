# Step 22: JSON.parse and terminal typed boundaries

Base origin/area/library: `ca016bab19c040c44748924cde12976afbf68978`. Branch: codex/library-json-parse.
Source: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.

The src inventory has 32 direct JSON.parse calls and zero revivers.
The callback at jsDocParsing.ts:101 is a JSON.stringify replacer, not a parse
reviver. All parse calls use one argument. sys.ts has no direct JSON.parse call.
The machine-readable inventory includes each exact source line and indirect
readers: [library-json-parse-inventory.json](library-json-parse-inventory.json).

| Site | Shape | Result typing or cast |
|---|---|---|
| src/compiler/sourcemap.ts:424 | Source map v3; file/mappings strings, sources array, optional names/sourceRoot/sourcesContent | any, then isRawSourceMap runtime predicate; returns RawSourceMap \| undefined |
| src/compiler/utilities.ts:7811 | Arbitrary JSON; common package.json and tsbuildinfo entry point | tryParseJson returns any; catches syntax failures and returns undefined |
| src/compiler/utilitiesPublic.ts:749 | Localized diagnostic-message dictionary of strings | argument to setLocalizedDiagnosticMessages(MapLike<string> \| undefined); no cast/check |
| src/harness/client.ts:183 | Server response or event; type discriminator, protocol body | assignment to response: T extends protocol.Response; no boundary check |
| src/harness/documentsUtil.ts:84 | Harness RawSourceMap with numeric version, file/mappings strings and string arrays | as RawSourceMap; assigned to this.raw |
| src/harness/evaluatorImpl.ts:158 | package.json; main field used for module loading | any inferred; main used directly |
| src/harness/fourslashImpl.ts:4633 | Array of TextChange {span:{start:number,length:number},newText:string}; deep copy | assignment to changes: readonly ts.TextChange[]; no cast/check |
| src/harness/fourslashImpl.ts:4864 | Marker object; arbitrary fields and optional name | as { name?: unknown }; name subsequently typeof-narrowed |
| src/harness/harnessIO.ts:898 | Source map; file and sources used to resolve preview inputs | any inferred; no boundary check |
| src/harness/harnessLanguageService.ts:725 | Arbitrary protocol message; parse then pretty stringify | any passed directly to JSON.stringify |
| src/server/editorServices.ts:1571 | typesMap entries with serialized match strings, exclude mixed string/number arrays, types string arrays; simpleMap string dictionary | raw: TypesMapFile; declaration typesMap uses SafeList whose match is RegExp, converted from string after parse |
| src/server/session.ts:3948 | Protocol Request {seq:number,type:"request",command:string,arguments?:any} | input as any as string; result as protocol.Request |
| src/testRunner/parallel/host.ts:206 | Saved test performance dictionary: test hash to number | contextual return { [testHash:string]:number } \| undefined |
| src/testRunner/projectsRunner.ts:199 | Project test {scenario,projectRoot,inputFiles,...} with compiler options | as ProjectRunnerTestCase & ts.CompilerOptions |
| src/testRunner/projectsRunner.ts:241 | Clone project test with compiler options; resolvedInputFiles/emittedFiles added afterwards | resolutionInfo: ProjectRunnerTestCaseResolutionInfo & ts.CompilerOptions; required resolution arrays absent at parse boundary |
| src/testRunner/runner.ts:122 | TestConfig optional flags, timeout/worker/shard numbers, runner/test string arrays, stackTraceLimit number or "full" | as TestConfig |
| src/testRunner/unittests/helpers/tsc.ts:410 | Readable buildinfo: version string, size number, optional incremental tables | as ReadableBuildInfo; subsequently narrowed to incremental variants |
| src/testRunner/unittests/helpers/typingsInstaller.ts:55 | Type registry {entries:dictionary<dictionary<string>>} | as TypesRegistryFile; entries passed to Object.entries |
| src/testRunner/unittests/jsDocParsing.ts:101 | Serialized AST/JSDoc nodes with syntax-specific fields | any inside stringify replacer; this callback is NOT a parse reviver |
| src/testRunner/unittests/publicApi.ts:306 | tsconfig object used by public API tests | any inferred; conditional undefined branch |
| src/testRunner/unittests/reuseProgramStructure.ts:832 | Deep copy of CompilerOptions object or string[] | any inferred return from overloaded duplicate; overloads declare CompilerOptions/string[] |
| src/testRunner/unittests/tsc/incremental.ts:433 | tsconfig object; compilerOptions.declarationMap changed | any inferred; no boundary check |
| src/testRunner/unittests/tscWatch/projectsWithReferences.ts:77 | tsconfig object; compilerOptions.paths replaced | any inferred; no boundary check |
| src/testRunner/unittests/tsserver/autoImportProvider.ts:289 | Package manifest object; name used as computed dictionary key | any immediate property access; no boundary check |
| src/testRunner/unittests/tsserver/projectsWithReferences.ts:110 | tsconfig object; project references edited | any inferred |
| src/testRunner/unittests/tsserver/projectsWithReferences.ts:128 | tsconfig object; project references edited | any inferred |
| src/testRunner/unittests/tsserver/projectsWithReferences.ts:225 | tsconfig object; project references edited | any inferred |
| src/testRunner/unittests/tsserver/projectsWithReferences.ts:243 | tsconfig object; project references edited | any inferred |
| src/testRunner/unittests/tsserver/symLinks.ts:157 | tsconfig object; compiler module-resolution option changed | any inferred |
| src/typingsInstaller/nodeTypingsInstaller.ts:63 | Type registry {entries:dictionary<dictionary<string>>} | as TypesRegistryFile; entries passed to Object.entries |
| src/typingsInstallerCore/typingsInstaller.ts:320 | package.json devDependencies dictionary | as NpmConfig; devDependencies declared MapLike<any> |
| src/typingsInstallerCore/typingsInstaller.ts:321 | package-lock.json optional dependencies/packages dictionaries with version strings | as NpmLock |

## Indirect readers

- compiler/emitter.ts:1143 reads buildinfo through readJsonOrUndefined, then casts
  to BuildInfo or undefined. BuildInfo requires version:string; incremental
  buildinfo has additional arrays, tables, options and flags.
- compiler/moduleNameResolver.ts:831,2470 reads package.json through readJson and
  casts to PackageJson. The optional manifest fields include name/version,
  main/types/typings, imports/exports objects, typesVersions and dependency
  dictionaries. Later reads validate individual fields. At line 831 typings
  can be null in the input although the declaration says optional string.
- compiler/moduleSpecifiers.ts:1154,1256 invokes tryParseJson directly, retaining
  an any result or Record<string,any> or undefined.
- services/utilities.ts:3600 casts tryParseJson to PackageJsonRaw or undefined,
  containing dependencies/devDependencies/optionalDependencies/peerDependencies
  string dictionaries.
- compiler/commandLineParser.ts:2245,2278 uses parseJsonText, the compiler's
  separate syntax parser for tsconfig. It is not a call to JSON.parse.

## Binding boundary ruling

A matching document compares with Node's unchanged source, byte for byte. A
nonmatching document is a fixture whose header says
`boundary check: Adamic stops, Node continues` and records Adamic's exact message.
The inserted check is terminal: exit 70, naming the field and both types. It
cannot be caught by tsc's surrounding catch. JSON syntax errors remain catchable
SyntaxErrors with Node's messages. No second oracle was introduced.

For example, parsing `{"version":1}` into `BuildInfo { version: string }` prints
`adamic: panic: boundary check: $.version expected string, got number` and exits
70, while the unchanged Node source continues. The fixture has a catch which
must not run. Missing fields, array elements, nested fields and literal types
have their own named-difference fixtures. They use the oracle's existing checked
fixture registration and an additional test that enforces each exact header.

## Implementation and supported results

The native parser ports V8's JSON grammar, diagnostic choice, UTF-16 positions,
line/column calculation and ten-character error-context windows from the V8
sources bundled with Node 24.19.0. THIRD_PARTY_NOTICES.md credits both source
files under the existing V8 BSD notice. Number conversion uses the native
binary64 decimal conversion, held to Node. Container parsing and temporary-tree
cleanup are iterative, so deeply nested discarded JSON does not consume the C
call stack. The JavaScript backend uses Node's actual JSON.parse, followed by
the same inserted terminal boundary contract.

Supported forms:

- Discarded raw parses validate arbitrary runtime text, including every JSON
  kind, duplicate keys, NUL keys, lone surrogates and deep containers.
- Declared scalar and homogeneous array results accept runtime text. Scalar
  unions and scalar literal types have checked contracts.
- Object and unknown results require a direct constant document whose complete
  layout can be proved. Matching BuildInfo and the source-map core have fixtures.
  Extra fields are retained, including their original key order, rather than
  erased to the declared structural view.
- An unannotated parse binding travels as tagged unknown, rather than trusting
  the standard declaration's any. Its admitted observations include typeof and
  named property probes. An explicit unknown result uses the same representation.

The small Go JSON decoder proves constant layouts only. It does not supply the
runtime values, syntax diagnostics or Unicode decoding: those all come from the
native V8 port. The compiler refuses any document layout it cannot prove.

## What remains refused

Runtime object/unknown results need owned, length-bearing native key and slot
metadata. Current structural types cannot describe arbitrary extra keys, and
native shape names are C strings. Consequently runtime-text BuildInfo,
package.json, source-map and protocol readers remain blocked; this unit does not
claim that tsc's full readers compile. Constant object layouts containing NUL or
replacement-character keys are also refused because the layout proof cannot
establish their exact native key representation. Discarded parsing admits them.

Other explicit parser refusals cover revivers/non-single argument lists,
ToString coercion beyond strings, any inside a declared schema, index-signature
schemas, nominal/non-JSON fields such as RegExp and Date, optional booleans with
a two-word slot representation, container unions needing recursive variant
selection, generic contracts needing per-instantiation schemas, recursive or
more-than-64-level result schemas, and heterogeneous constant array layouts.
Unknown array indexed reads still hit the base compiler's ElementAccessExpression
refusal. General any returns, non-null assertions, dictionary/index-signature
operations and other source-language gaps also block the original tsc functions.
Each remains a compile-time refusal; no fallback result is fabricated.

Two tsc declarations are wrong at their parse boundary independently of those
compiler gaps: editorServices.ts:1571 declares RegExp match fields which the
serialized document represents as strings and converts afterwards;
projectsRunner.ts:241 declares required resolution arrays which it only adds
later. A sound boundary must never silently accept those intermediate values.

## Fixtures, mutants and verification

Four conforming registered fixtures cover grammar/runtime input, scalars and
arrays, tsc BuildInfo/source-map shapes, and untyped observations. Native under
ASan/UBSan and leak checks, JavaScript, and WASI agree with Node 24.19.0.
Five registered nonconforming fixtures enforce the known difference on all three
backends, with exact terminal messages. The syntax sweep compares 2,126
malformed, mutated and valid documents with unchanged Node; the grammar fixture
also validates 20,000 levels of nesting.

Three conforming-family mutants change the SyntaxError name, add one to parsed
numbers, or replace the parsed BuildInfo version. Each compiles, exits zero,
has empty stderr and clean sanitizers/leak checks. Only stdout comparison with
Node catches them. A fourth mutant removes the boundary and therefore agrees
with Node; the named-difference header contract alone catches that mutant. This
last distinction is required by the ruling, not treated as a Node disagreement.

Test262 is pinned at c8c798898646638cd0c24879f8e0374e847e7d74, adaptation enabled,
stock TypeScript 6.0.3, complete built-ins/JSON/parse directory:

| Outcome | Before | After |
|---|---:|---:|
| Pass | 0 | 32 |
| Disagreement | 0 | 0 |
| Refused | 50 | 18 |
| Not TypeScript | 7 | 7 |
| Crashed | 0 | 0 |
| Skipped | 20 | 20 |
| Total | 77 | 77 |

All 32 admitted adapted source programs were replayed on JavaScript and WASI
against their unchanged adapted source on Node. Their SHA-256 hashes match the
sources measured by the runner: 32/32 passes per backend, zero disagreements.
The remaining test262 refusals are ten `new an Identifier`, five var, one
reviver/non-single parse argument list, one dynamic prototype property value,
and one delete. The seven non-TypeScript outcomes retain matching stock tsc
rejections. Skips name Proxy, Reflect.construct, Symbol, getters and unsupported
harness includes. There is no claim that refused or skipped tests passed.

Focused checks and vet cover internal/lower, internal/flow, internal/ir,
internal/native, internal/javascript and internal/oracle. The parser comparisons,
WASI comparisons, exact boundary assertions, syntax sweep, mutants and the
existing stringify refusal check passed. Run output, downloaded sources, the
pinned test262 tree and replay artifacts stay outside the checkout under
/tmp/library-json-parse/. The rest of the repository is left to the fast gate.

## Linux counts before and after

Recorded fixtures: 635 -> 644. New rows have no previous baseline:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| json_parse_grammar.a | 282 | 282 | 249 | 475 | 7 | 0 |
| json_parse_scalars.a | 58 | 58 | 32 | 84 | 8 | 0 |
| json_parse_shapes.a | 57 | 57 | 34 | 84 | 17 | 0 |
| json_parse_inputs.a | 1 | 1 | 0 | 1 | 1 | 0 |
| json_parse_boundary/array.a | 15 | 11 | 4 | 15 | 6 | 0 |
| json_parse_boundary/buildinfo.a | 14 | 10 | 4 | 14 | 6 | 0 |
| json_parse_boundary/missing.a | 13 | 10 | 4 | 14 | 5 | 0 |
| json_parse_boundary/nested.a | 23 | 15 | 8 | 23 | 10 | 0 |
| json_parse_boundary/literal.a | 14 | 10 | 5 | 15 | 6 | 0 |

Terminal fixtures are counted at the stop, as the counts table specifies;
matching programs free all their values. The checkout-enumerating
node_fs_directory_system.a row changes from 3220/3220/5812/4829/1938/0 to
3240/3240/5853/4859/1953/0 (allocations/frees/retains/releases/peak/regions).
All other existing rows are unchanged. internal/oracle/counts.md is authoritative.
