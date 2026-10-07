# TypeScript compiler namespace fixtures

Twelve small programs cover every one of the census's **11 namespace declarations
in 7 files**: ten runtime declarations, including two nested declarations, and
one type-only namespace. The census reason is exactly `a namespace`.
Source is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
All selected executable functions and statements are copied from that source.
Drivers and reduced external type contracts make the slices runnable; these are
not claims that complete upstream functions or compiler files compile.

Read references: `stage3/census/REPORT.md` and `data/sites.json` from
`origin/codex/tsc-census` (`429c1177`), `stage3/README.md` from
`origin/codex/stage3-base` (`87284051`), and `docs/namespaces.md` on both
namespace branches. The latter distinguishes qualified singleton storage from
callable merges and escaped runtime containers.

| Fixture | Original declaration and exercised form |
| --- | --- |
| 01_builder_state | `BuilderState`, builderState.ts:100; interface merge, original optional-state reuse decision with both absent/present maps |
| 02_jsx_names | `JsxNames`, checker.ts:54223; all ten exported branded string constants |
| 03_react_names | `ReactNames`, checker.ts:54236; single exported branded string constant |
| 04_debug_state | `Debug`, debug.ts:113; mutable exported enum/logging state, qualified writes, original `shouldLog` |
| 05_debug_log_merge | `Debug.log`, debug.ts:137; hardest retained case: callable function/namespace merge, nested functions calling parent `logMessage` and `shouldLog`, live host replacement, function identity |
| 06_binary_expression_state | `BinaryExpressionState`, factory/utilities.ts:1273; generic callable type alias and namespace coexistence, original `done`, self-reference and detached function identity |
| 07_parser_singleton | `Parser`, parser.ts:1437; uninitialized private var, original `countNode`, singleton reset across two calls |
| 08_parser_jsdoc_nested | `Parser.JSDocParser`, parser.ts:8790; two nested const enums, parent singleton token access through original `token` and `nextTokenJSDoc`, original `parseOptionalJsdoc` body |
| 09_incremental_parser | `IncrementalParser`, parser.ts:9946; private original `shouldCheckNode` switch and const enum `InvalidPosition` |
| 10_tracing_escape | `tracingEnabled`, tracing.ts:37; private mode/catalog, original `recordType`, const string enum, `tracing = tracingEnabled` escape and calls through its alias |
| 11_status_type_only | `Status`, tsbuild.ts:58; all fourteen original interfaces and their enum discriminants; no runtime namespace |
| 12_builder_release_cache | `BuilderState`, builderState.ts:100; second interface-merge case, original `releaseCache` writes, isolated namespace admission control |

## Every reopening and merge

`declarations.json` records a stock TypeScript 6.0.3 AST and symbol audit of all
compiler sources, checked against all eleven census locations.
**No namespace is reopened** in the pinned `src/compiler` tree.

| Namespace declaration | Other declaration on the same symbol |
| --- | --- |
| BuilderState, builderState.ts:100 | InterfaceDeclaration, builderState.ts:62 |
| Debug.log, debug.ts:137 | FunctionDeclaration, debug.ts:133 |
| BinaryExpressionState, factory/utilities.ts:1273 | TypeAliasDeclaration, factory/utilities.ts:1271 |

There are no other namespace merges. Interface/type-alias coexistence is admitted
by both feature branches; callable `Debug.log` remains a possible source adaptation.
This bucket does not implement that adaptation or invent a reopened namespace
fixture when upstream has none.

## Scope of the reductions

The tiny BuilderState contracts retain the fields read or written by each selected
function. Fixture 12 explicitly represents present undefined rather than applying
the optional-declarations adapter to upstream. The escaped-name support retains
the original string-brand arm from types.ts:6196; unused void and internal-name
arms are omitted. BinaryExpressionState's machine and node types are reduced,
and its assertion helper is driver support, not a port of Debug's assertion cache.

The parser scanner in 08 supplies a fixed comma-to-EOF JSDoc token stream;
it does not implement scanning. `parseOptionalJsdoc`, originally local to
`parseJSDocCommentWorker`, is isolated directly in the nested namespace with its
body unchanged. Fixture drivers expose otherwise private functions and set the
same state slots upstream initializes. Tracing runs catalog logic without the
filesystem, timestamps or JSON host. Date and Path in 11 are unused/reduced type
support, and every Status interface itself is retained unchanged.

Not covered: all 437 Parser functions, all 75 Debug functions, scanner/factory
initializers, namespace classes, assertion-cache reflection, overloads, complete
JSDoc parsing, and tracing host I/O. They need their own feature/host fixtures.
No compiler production file, adaptation, other fixture bucket, shared fixture
Go test, or native implementation was edited. The complete adapted-tree upstream
suite and full uncached integration gate were not run for this fixture-only unit.

## Recorded experiments

`status.json` uses exactly the shared fixture schema and records **current main**.
`validation.json` additionally records both feature branches, exact diagnostics,
and native execution for every successful build. `mutant.json` contains the
flattened source and its complete main/Node observations.
The Node runner is `oracle/node.mjs` from **codex/namespaces-tsc**, using Node's
independent `stripTypeScriptTypes(..., { mode: 'transform' })` change; no Adamic
JavaScript output serves as the source oracle.

Pinned compilers:

- main: `ef3d907ecdc4c771b016f7d9c52372def057a340`
- codex/namespaces-tsc: `ce8a2acf14e420a9c82345236845a377cd4c7a50`
- codex/parameter-properties-namespaces: `adc45ca417482834336c08510be0db5b91974680`

All three use cohere `715ba94f3608a6500086b1076ce5cb7e51b836db` and its pinned
TypeScript Go submodule `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.

Commands, with shell environment loaded from `/workspace/adamic-tools/env.sh`:

```sh
bash cloud/setup.sh > /tmp/namespaces-setup.log 2>&1
node --disable-warning=ExperimentalWarning /tmp/namespaces-tsc/oracle/node.mjs <fixture.a> > /tmp/fixture-node.log 2>&1
go run ./cmd/adamic build <fixture.a> -o /tmp/fixture-native > /tmp/fixture-build.log 2>&1
/tmp/fixture-native > /tmp/fixture-native.log 2>&1
go test ./internal/load ./internal/lower -count=1 -timeout 30m > /tmp/namespaces-main-packages.log 2>&1
```

Build/run commands were repeated in the main checkout and both detached feature
worktrees. Every test/experiment output was captured to a log, never piped.
Setup: Go 1.27.1, clang 20.1.8, Node 24.19.0; go/clang/node/submodules 0s each,
build cache 99s, total 99s; `nproc` 5, cgroup CPU quota 4.
Main package tests: load 1.659s, lower 23.281s, both pass.
All twelve inputs also pass independent stock TypeScript 6.0.3 with strict,
noUncheckedIndexedAccess, exactOptionalPropertyTypes, noImplicitReturns and
noFallthroughCasesInSwitch. `.a` files are explicitly parsed as TypeScript and
forced modules, as Adamic loads them. Selected original executable spans were
checked against the pinned source after newline normalization.

The flattening mutant removes only the BuilderState wrapper from fixture 12,
keeps its releaseCache statements intact, and changes the driver to call the
module function. Main's outcome changes from Checker (TS1294 on the namespace)
to Compiles. Node and the mutant native binary both print `true true\n`, exit 0,
with empty stderr. Thus the namespace admission/status check can fail; this
mutant does not claim to demonstrate a runtime namespace ownership fault.


## Final outcomes

| Compiler | Checker | NotYet | Refused | Compiles |
| --- | ---: | ---: | ---: | ---: |
| main | 12 | 0 | 0 | 0 |
| codex/namespaces-tsc | 0 | 6 | 2 | 4 |
| codex/parameter-properties-namespaces | 0 | 5 | 6 | 1 |

All 12 Node executions exit 0 with empty stderr. **No silent miscompile was
observed**: all five successful feature-branch builds match stdout, stderr and
exit status byte for byte. All five also pass ASan/UBSan and leak detection,
recorded in `sanitized.json`. These historical feature CLIs reject `--sanitize`
with usage/exit 2 (go run exits 1). A scratch Go overlay sets only
`native.Options.Sanitize: true` in their CLI invocation of native.Build; no source
file or language option is changed. This uses their existing native sanitizer
implementation and leaves their checker/lowering untouched.

The filtered feature oracle also passes on codex/namespaces-tsc:

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/namespaces_(debug_state|parser_state)\.a$' -count=1 -timeout 30m -v > /tmp/namespaces-feature-oracle.log 2>&1
```

Both cases ran; total 6.576s, four native and four Node cache misses, two native
hits for the leak phase. The recorded log is in `logs/feature-oracle.log`.
`logs/observations.log` records the complete 36-build matrix.

Reproduce the matrix with prepared main and feature worktrees and their pinned
submodules (the script writes the observations in this bucket):

```sh
python3 stage3/fixtures/namespaces/observe.py --main /workspace/adamic --namespaces-tsc /tmp/namespaces-tsc --parameter-properties-namespaces /tmp/namespaces-pp --logs /tmp/namespaces-runs > /tmp/namespaces-check.log 2>&1
```

The script never classifies an unexpected build error as NotYet or success, and
returns nonzero on a native/Node mismatch. The shared Go fixture test remains the
other worker's responsibility.
