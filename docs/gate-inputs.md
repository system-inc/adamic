# Inputs the correctness gate must provide

This is the setup worker's inventory for the AST census at area/developer-tools
`10e709cef5b6190c42fe8249f83d2cde20f5a402`, with the post-opt-in rule applied. A missing correctness input is
`required-input`, including debug bypasses and harness request files. Benchmark
and profile switches belong to `measurement` and do not satisfy verification.
The machine-readable declaration is
[skips.json](../internal/skipcensus/testdata/skips.json).

Run `bash cloud/setup.sh --gate-inputs`, then source the env.sh it creates.
This opt-in installs the pinned shallow TypeScript checkout and locked npm
projects under `$ADAMIC_TOOLS/gate-inputs`, builds the checker archive, and
exports their paths plus `ADAMIC_GITIGNORE_LARGEST=1`. npm installs use `npm ci`
with checked-in manifests and integrity hashes in `cloud/gate-inputs/`. A second
run validates installed bytes and pins locally, skips downloads, and reports
each cached input as skipped. Without the flag, setup retains its ordinary
behavior. This worker writes `/workspace/adamic-tools/env.sh`. The observed tools are Go 1.27.1, clang
20.1.8 and Node 24.19.0. Setup selects Go from go.mod, sanitizer-capable clang,
and Node 24; it does not enforce exact patch versions for clang or Node. For a
reproducible fleet, provision these observed versions explicitly. Keep TMPDIR
world-traversable (`/tmp/adamic-gate`, mode 1777), and Node outside /root.

The repository pins cohere to `715ba94f3608a6500086b1076ce5cb7e51b836db` and
its typescript-go submodule to `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
`git submodule update --init --recursive` provides both. This checker source is
separate from the upstream TypeScript v6.0.3 corpus below.

| Variable or lookup | Exact required input and how the test finds it |
| --- | --- |
| `ADAMIC_TYPESCRIPT_SOURCE` | Absolute root of microsoft/TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Must be a Git checkout with `src/compiler/*.ts`. Parser `compilerManifest` and lint `TestCompilerAndStage1Agree` check `git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD` and walk `src/compiler`. |
| `ADAMIC_CSS_LIBRARY` | npm project root containing package.json and node_modules with `postcss@8.5.16` and `postcss-scss@4.0.9`. CSS testdata/library.mjs resolves packages with createRequire at this root and checks both versions. |
| `ADAMIC_CSS_PRINTER_LIBRARY` | npm project root with `node_modules/prettier`, version `3.9.6`. CSS testdata/print_library.mjs and gaps/printer_library.mjs load from this path. The embedded fork comes from pinned cohere/internal/format/prettier/bundles, provided by the submodule. |
| `ADAMIC_JSON_PRETTIER` | npm project root containing package.json and `node_modules/prettier@3.9.6`. JSON testdata/library.mjs resolves from that package.json and rejects version drift. |
| `ADAMIC_GRAPHQL_LIBRARY` | npm project root containing package.json and `node_modules/graphql@17.0.2`, resolved by GraphQL testdata/library.mjs. |
| `ADAMIC_MEDIA_QUERY_LIBRARY` | npm project root containing package.json and `node_modules/postcss-media-query-parser@0.2.3`, resolved by mediaquery testdata/library.mjs. |
| `ADAMIC_SELECTOR_LIBRARY` | npm project root containing package.json and `node_modules/postcss-selector-parser@2.2.3`, resolved by selector testdata/library.mjs and gaps/nontermination.mjs. Both parity and nontermination proofs need it. |
| `ADAMIC_VALUES_LIBRARY` | npm project root containing package.json and `node_modules/postcss-values-parser@2.0.1`, resolved by values testdata/library.mjs. |
| `ADAMIC_GITIGNORE_LARGEST=1` | Enables the test-generated ignore file of exactly 104857600 bytes. This is a switch, not a path to a downloaded corpus. Gitignore askedCases passes largest=true to its overlay generator; the boundary mutant must run. Allow scratch space for cases over 100 MB. |
| `ADAMIC_CLANG_TSGO_ARCHIVE` | Absolute file path to the C archive built from this tree's bridge/tsgo/archive and the pinned checker above. Build with `go build -buildmode=c-archive -o /absolute/path/tsgo.a ./bridge/tsgo/archive`; set the variable to that tsgo.a. TestSplitTSGoAgrees links it with both whole and split generated C and compares outputs. No independent archive release version exists: the pin is this tree plus both submodule commits. |
| `PATH`: node and clang | TestMiniRunner uses exec.LookPath for both. nodePrototypeMembers executes Node to enumerate Object.prototype. Supply Node 24.19.0 and sanitizer-capable clang 20.1.8 as above; absence is required-input. |
| `ADAMIC_C_COVERAGE` | Must be unset (or not equal to `1`) in the normal correctness gate. CoverageRequested in internal/native/native.go reads it; forcing coverage on every build skips TestRuntimeCacheKeepsCoverageApart, losing the proof that covered and ordinary runtimes use separate cache entries. A coverage shard must have a separate normal run of this check. |
| `ADAMIC_JSON_NODE_ONLY` | Must be unset. A nonempty value bypasses native JSON parity and triggers required-input skips in TestPortMatchesGoCohere and TestAdditionalJSONBoundaries. |
| Native regex dependency | TestRegexCycleFixtureHasItsNativeDependency reflects on ir.Program for Regexps. The gate tree must include native regex lowering/emission, not just a fixture manifest. There is no environment path or independent version: this is a required source dependency. |
| `ADAMIC_STAGE3_FIXTURE`, `ADAMIC_STAGE3_RESULT` | stage3/fixtures TestFixtures builds the oracle test binary and invokes TestStage3FixtureHook with an absolute fixture source path and writable JSON result path. Inputs are the checked-in fixture manifest and source from the current tree, not an external package. A plain hook skip is still required-input under the conservative census rule. |
| `ADAMIC_PORT_REQUEST` | Parent slice tests generate a JSON request file in t.TempDir, set this absolute path, and run the corresponding testdata/cohere_side_test.go through a Go overlay against pinned cohere. CSS also overlays compose_side_test.go and print_side_test.go. The request carries Cases/Answers, or Scratch/Output/Mode and generation parameters as appropriate to that slice. Use the parent harness, which supplies the exact schema and writable destinations; do not run these overlay tests standalone without a request. This covers CSS, formatfiles, gitignore, GraphQL, mediaquery, selector, suppression and values. |
| `ADAMIC_JSON_CASES`, `ADAMIC_JSON_ANSWERS` | JSON audit_test.go writes a JSON array of `{name,text}` cases, sets CASES to its absolute path and ANSWERS to a writable output JSON path, then overlays testdata/cohere_side_test.go into pinned cohere. This is generated by the parent audit harness. |

For the supplied npm inputs, setup uses the checked-in package.json and
package-lock.json in each `cloud/gate-inputs/<name>/` project. Shared formatter
exports reuse the `css-printer` project. The installer verifies the npm bootstrap
integrity, runs `npm ci`, and validates installed bytes on subsequent runs.
TypeScript is fetched shallowly by its exact commit and verified with
`git rev-parse HEAD` and `git fsck`. The exported largest-gitignore switch enables
the parent harness's generated 100 MiB case.

The census's `reads` field is a conservative list of calls in the lexical
function and same-package test helpers it can reach. It includes optional
artifact destinations and corpus overrides, not only the input causing a skip.
Those optional reads are not newly mandated gate inputs. In particular
COHERE_GIT_SOURCE and generation seed/count overrides do not guard a Skip site.
This census does not detect omitted checks expressed only as logging or an
optional comparison with no Skip call.

## Required shards added by the latest area merge

The cohere shard must set `ADAMIC_GATE_COHERE=1` and run
`TestRepositoryPassesCohereBaseline` in `./cloud`. Its Python runner builds the
pinned cohere submodule and reads `cloud/cohere-baseline.json`; it needs Go,
Python 3 and pinned Node v24.19.0. Its skip is required-input, even though
ordinary development loops leave the switch off.

The required WASI shard runs `bash cloud/setup.sh --wasi-sdk`, sources the
printed env.sh and sets `ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1`. Setup pins
wasi-sdk 27.0. `WASI_SYSROOT` must name its `share/wasi-sysroot` directory,
including headers and wasm32-wasi libraries. The SDK's sibling `bin/clang`,
linker and compiler builtins must be usable; the integration test also looks up
clang, node and go on PATH and probes a real WASI build. Node v24.19.0 must
provide node:wasi and node:module registerHooks. Missing switches, tools,
sysroot or failed prerequisite probes are required inputs. Once opted in,
TestWASI reports missing sysroot, PATH tools, failed compiler probes and Node
probes with Fatalf and preserves their diagnostic text. Put
`/path/to/wasi-sdk/bin` before the native clang directory on PATH: having only
the WASI sysroot is insufficient when native clang lacks WASI builtins. The gate's
WASI plan assigns these to its required shard rather than treating opt-out as
successful verification. After opting in, emission fixtures marked as not
lowering fail rather than skip; the shard must satisfy its selected scope.

The optional release lane similarly fails when a selected fixture does not
finish on Node. Parser performance comparisons opted in with
`ADAMIC_PARSER_BENCH=1` fail if `ADAMIC_TYPESCRIPT_SOURCE` is absent, using the
same pinned v6.0.3 checkout described above. Opting into a measurement does not
make its missing inputs an acceptable skip.

Node v24.19.0 is now enforced by internal/nodepin and setup verifies its archive
against `cloud/node-pin.json`; the earlier observation that Node patch versions
were not enforced describes the initial base, not the current merged tree.
