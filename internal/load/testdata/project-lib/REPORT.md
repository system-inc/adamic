Built collection iterator result typing and project-aware TypeScript checking; JSON.stringify is unchanged.
Iterator commit: d5b7660fa2254e6dce58bf594a6a40aea77f916a; project work is the commit containing this report.
Loader, lower, flow, whole oracle, vet, native/JavaScript/WASI witnesses and test262 passed; zero disagreements.
Seven checker mutants failed their regression tests; omitted-done-as-true ran cleanly and was caught only by Node stdout.
Not covered: rerunning the entire 259-row ledger or compiling all of tsc; ambient witnesses need executable counterparts.

## Behavior and assumptions

MapIterator and SetIterator next return IteratorResult<T, undefined>, using TypeScript's own done discrimination. Yield inference through Iterable<T> stays T. The truth correction now lives in the selected lib.es2015.iterable declaration, so an ES5 project does not acquire iterator declarations. An optional done test uses false for omission; observing done still preserves undefined. Bound collection next calls retain their internal runtime closure; detached reads and copying inherited next remain refused.

For TypeScript roots, Load locates the nearest tsconfig.json and uses TypeScript's parser, including extends, project options, selected libs, type packages and project declaration roots. Declaration diagnostics are collected when the project emits declarations, including isolatedDeclarations. Inherited Node console signatures are not shadowed by the standalone one-string signature. JSON.stringify remains string | undefined; no overload was added.

Assumptions recorded in the commits:

- A compiler program has one project configuration, including imports. Explicit roots from different projects are refused; ambient .d.ts roots share the implementation project's options. Load still selects the requested implementation roots and their imports, plus project declaration roots, rather than compiling every implementation file named by the project.
- Standalone .a inputs retain Adamic defaults. Inputs without a project retain the existing default options. A TypeScript project gets its own lib and options.
- Standalone .a Set extensions live in prelude_set.d.ts. Project .ts checking does not inject them; ES2025's own declarations remain available when selected. The repository tsconfig explicitly roots the moved file, preserving the declarations that this project previously declared through its prelude.
- The five original ledger files have ambient collect/visit/transform declarations and cannot all execute on Node. They are preserved verbatim as checking witnesses. Populated executable counterparts use concrete MapIterator bodies supported by lowering. Generic Iterable Array.from bodies still refuse; this unit does not add that compiler capability.

The custom Set fixture is authored as .a and copied to a temporary .ts project root by the loader test. It checks clean under ES2024 and fails under ES2025, matching stock TypeScript 6.0.3. Node runs that custom object and prints 0. It is a checking fixture, not a claim that lowering supports arbitrary custom objects through native Set representation.

## Diagnostic counts

Baseline is f4efdd2369311d1420aa53fdf5c1a55bdda811d4. Ledger witnesses are copied from 3f0926c0a55a7b5f64f037b1745e0e984e08c8be.

| Witness | Before | After |
| --- | ---: | ---: |
| iterable.a | 1 | 0 |
| callback.a | 1 | 0 |
| entries.a | 1 | 0 |
| set-copy.a | 1 | 0 |
| nested-entries.a | 2 | 0 |
| custom Set in ES2024 project | 1 | 0 |
| Total | 7 | 0 |

All five exact witnesses also check clean as generated .ts roots under an ES2024 project. Stock TypeScript 6.0.3 reports zero on those five and the ES2024 custom Set, and one on the ES2025 custom Set. A real @types/node 25.3.3 project with ES2020, NodeNext, strict, strictBindCallApply=false, useUnknownInCatchVariables=false and skipLibCheck=false checks clean, including console.log() and console.log(42, 'value').

## Test262

Corpus c8c798898646638cd0c24879f8e0374e847e7d74; Node v24.19.0. Both tables are observations, not inferred totals. Columns below are Pass / Fail / Refused / Crashed / Skipped, following the CLI JSON fields.

| Group | Before | After | Total |
| --- | --- | --- | ---: |
| built-ins/Map | 6 / 0 / 77 / 0 / 121 | 6 / 0 / 77 / 0 / 121 | 204 |
| built-ins/Set | 19 / 0 / 205 / 0 / 159 | 19 / 0 / 205 / 0 / 159 | 383 |
| built-ins/Iterator | 0 / 0 / 240 / 0 / 414 | 0 / 0 / 240 / 0 / 414 | 654 |
| built-ins/JSON | 7 / 0 / 95 / 0 / 63 | 7 / 0 / 95 / 0 / 63 | 165 |
| Total | 32 / 0 / 617 / 0 / 757 | 32 / 0 / 617 / 0 / 757 | 1406 |

## Commands and results

All test output was redirected directly to logs, not piped. Every shell using the toolchain sourced /workspace/adamic-tools/env.sh. Relevant final commands:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/adamic-overlay-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/adamic-overlay-npm.log 2>&1
node -v

# Baseline CLI built before edits; each witness got its own log.
go build -o /tmp/adamic-overlay-before ./cmd/adamic
# For each of iterable, callback, entries, set-copy, nested-entries:
/tmp/adamic-overlay-before types internal/load/testdata/overlay-iterators/iterable.a > /tmp/adamic-overlay-before-iterable.log 2>&1

# Same test262 command before edits, after step 1, and after the final implementation.
ADAMIC_GATE_UNCACHED=1 go run ./cmd/adamic-test262 -adapt -jobs 4 -json \
  -test262 /tmp/adamic-test262-corpus built-ins/Map built-ins/Set built-ins/Iterator built-ins/JSON \
  > /tmp/adamic-overlay-test262-final2.json 2> /tmp/adamic-overlay-test262-final2.log

# Record counts for the six new registered fixtures.
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts \
  > /tmp/adamic-overlay-part1-counts.log 2>&1

# Full requested gate, run at step 1 and again on the final implementation.
ADAMIC_GATE_UNCACHED=1 go test ./internal/load ./internal/lower ./internal/flow ./internal/oracle \
  -count=1 -timeout 30m > /tmp/adamic-overlay-final2-gates.log 2>&1

# The ordinary oracle compares source Node, sanitized native, release native and JavaScript,
# and checks leaks. This also runs the project copies and the WASI leg.
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle \
  -run 'TestProjectIteratorBackends|TestLibraryIteratorDoneMutant|TestWASIAgreesWithNode/internal/oracle/testdata/library_overlay_' \
  -count=1 -v > /tmp/adamic-overlay-final2-backends.log 2>&1

go vet ./... > /tmp/adamic-overlay-final2-vet.log 2>&1
gofmt -l cmd internal > /tmp/adamic-overlay-final2-gofmt.log

# Seven isolated Go overlay mutants, no checkout edits.
python3 internal/load/testdata/project-lib/run_mutants.py /tmp/adamic-overlay-repro-mutants \
  > /tmp/adamic-overlay-repro-mutants.log 2>&1
```

Final package results: load 6.058s, lower 37.893s, flow 139.671s, whole oracle 227.999s, all exit 0. The focused backend/WASI/mutant run passed in 4.103s. Vet exited 0 with an empty log; gofmt listed no files. Final CLI checking reports exit 0 for every exact witness, the custom Set project and the real Node-types project.

The independent stock check used stage3/api's TypeScript 6.0.3 createProgram/getPreEmitDiagnostics, strict=true, target=ES2024, types=[], noEmit=true, with each .a witness copied to a temporary .ts file and a console declaration. lib.es2024 was used for all six; the custom Set was repeated with lib.es2025. Results are in /tmp/adamic-overlay-stock-checks.log. The custom object's Node execution was:

```sh
node oracle/node.mjs internal/load/testdata/project-lib/custom_set.a > /tmp/adamic-overlay-custom-set-node.log 2>&1
```

It exited 0 and printed 0. Node's experimental stripping warning is not a program disagreement; oracle invocations disable that warning.

## Mutants

| Mutant | What caught it |
| --- | --- |
| Anonymous next union folds completion undefined into yield inference | All five exact witnesses and all five project copies fail checking; 1/1/1/1/2 diagnostics recur |
| Inject standalone Set extensions into an ES2024 project | Custom Set checking fails TS2740, missing the seven methods |
| Override project options with compilerOptions() | ES5, ES2025, DOM/default libs, strict=false, optional-property and unchecked-index tests fail |
| Use narrow standalone console methods in the project prelude | Inherited project console calls fail checking |
| Drop project declaration roots | Project globals and its declared Set method fail checking |
| Skip declaration diagnostics | Unannotated isolated declaration is incorrectly accepted; the regression test fails |
| Accept roots from different projects | Mixed-project refusal test fails |
| Treat omitted done as true | Builds with -Werror, exits 0 under ASan/UBSan/LSan without stderr; only comparison with source Node stdout catches it |

The seven checking mutants all exited 1 with failing regression tests and no Go build failure. The last mutant is permanently tested by TestLibraryIteratorDoneMutant. No runtime C file changed, and no new libc call was introduced.

## Setup timings

Setup succeeded. Cumulative ready lines: node 0.034s, Go 0.035s, submodules 0.079s, markdown 0.094s (validated installed dependencies, 0.011s step), clang 0.182s, WASI SDK 0.247s, Go build 39.409s, test binaries deferred 39.593s, cache warm 39.594s, done 39.631s. nproc=5; cpu.max=400000 100000; 17.6 GB. Go 1.27.1, clang 20.1.8, WASI SDK 27, Node v24.19.0. npm ci added three packages in 380ms.
