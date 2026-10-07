Built: no-new-func, no-new-native-nonconstructor, and no-new-wrappers; nine wave 12 ports are complete.
Commits: claim `2ac7b71c` was pushed before implementation `32f968f3`.
Checks: 94 controls and the frozen 77/287 corpora agree byte for byte, normal and sanitized.
Mutants: three rule mutants, one raw-question mutant, registry retention, and the Node one-byte mutant were caught.
Uncovered: full repository gate, prior 26-rule regression suite, full fixture matrix, and emitted-JavaScript lint comparison.

## Selection and ownership

Six earlier claims were completed and pushed before this batch. Fetching all origin heads inspected 341 refs. Combined VOLUME_REPORT count tables contain 197 checker-dependent rules; 118 ranking names were claimed across origin branches and 25 matched checker-dependent ports on main/the bridge branch. The additional existing inventory port is not checker-dependent. Of 54 remaining entries, the first three with lexical ties were no-new-func, no-new-native-nonconstructor and no-new-wrappers, each with zero compiler and repository counts. Earlier remaining Nexus entries were now claimed and skipped. Exact matching module filename checks across origin branches found no competing ports of these three.

The claim update was pushed before code. Every native implementation module is `.a` inside `wave12_third`. Shared harness, registration generator, dispatcher, compiler, and submodule files were untouched. The only additions outside this directory are the separately named Go raw-question file and its direct-checker test. One initializer line registers the question using the registry already on this branch.

## Implementation and independent oracle

Each rule has its own file. Adamic decides accepted constructor/method names, parenthesis handling, global versus shadow classification, report anchors, and complete messages. Function accepts direct calls/new and static call/apply/bind properties, including template and parenthesized subscripts. Nonconstructor deliberately does not skip parenthesized callees, matching production Go, and reports the identifier. Wrappers skip callee parentheses and report the whole construction.

`symbol-declaration-files` returns raw declaration filenames and IsDeclarationFile bits in binder order. It does not follow aliases. Adamic tests the first declaration, matching production Go exactly. Existing alias-following ancestry would change this behavior. The bridge calls no cohere rule and emits no verdict, finding, or edit. The new question has its own Go and Adamic file and retains ABI version 1 and its ownership contract.

The isolated Go oracle loads its own program, walks its own AST, and invokes the three pinned production Run methods unchanged. It imports no bridge code. The complete serializer compares rule names, message IDs/text, UTF-8 spans, fixes, and suggestions, preserving duplicates and repair ordering. These rules have no fixes/suggestions; both zero fields are compared.

Cohere remains pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, and TypeScript compiler at `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Agreement and controls

The unchanged coverage manifests retain 77 compiler roots and 287 repository roots (212 `.a`, 75 `.ts`). Configured declaration files are retained. Controls contain 11 targeted sources plus 83 literal source rows extracted from pinned Go reference tests: 44 Function, 14 nonconstructor, 25 wrapper rows. Each is reevaluated by production Go rather than copying expected diagnostics. Controls include direct/indirect invocations, optional access, template/string/dynamic subscripts, callee parentheses, nested shadows, named functions, parameters, hoisting, imported source/ambient shadows, multiple findings, Unicode and CRLF. Generated helper.ts is a TypeScript reference input, not an Adamic implementation module; its preserved bytes have a .reference.txt evidence filename.

| Population | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| 94 controls | 56 | 25,839 | PASS |
| 77 compiler roots | 0 | 5,010 | PASS |
| 287 repository roots | 0 | 18,485 | PASS |

Positive controls prove these are not no-op ports. All native comparison stderr is empty. Full compressed Go/native/sanitized/mutant output, stream hashes, source hashes, manifests and controls are in [evidence](evidence/diagnostic-hashes.json). Absolute paths in streams describe this run; portable frozen manifests are preserved separately.

## Commands and outputs

The initial toolchain setup remains in use: bash cloud/setup.sh passed, Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s, build cache warm 115s, total 115s. nproc again printed 5; quota is four cores. Go 1.27.1, clang 20.1.8, Node 24.19.0. Commands source `/workspace/adamic-tools/env.sh`; all test output goes to files.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE12_THIRD_ARTIFACTS=/workspace/wave-12/third/attempt3 \
ADAMIC_WAVE12_THIRD_REPOSITORY_MANIFEST=/workspace/wave-12/repository.manifest \
ADAMIC_WAVE12_THIRD_COMPILER_MANIFEST=/workspace/wave-12/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus \
go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware/wave12_third \
    > /workspace/wave-12-third-attempt3.log 2>&1
```

PASS, **113.028s**, including normal/sanitized archive and native builds, controls, both corpora and every mutant. Direct checker package: `go test -v -count=1 ./bridge/tsgo/checker`, PASS **0.160s**. It compares raw declarations directly against checker APIs and checks alias preservation, unresolved symbols, shadows and invalid requests. `go vet ./bridge/tsgo/... ./stage1/cohere/typeaware/wave12_third`, gofmt over touched Go files, and git diff --check: clean.

```sh
go test -v -count=1 -timeout 30m ./internal/oracle \
  -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' \
  > /workspace/wave-12-third-node.log 2>&1
```

PASS **22.080s**, five native/Node/emitted-JavaScript fixtures with sanitizers and the one-byte mutant. This is the filtered compiler oracle, not emitted-JavaScript execution of the new lint suite.

The first attempt failed building the isolated oracle because the two imported core packages collided. The second byte comparison caught the shared diagnostic helper's default TypeScript namespace on these core rules. Both were corrected inside this directory. Failed logs are preserved; build failures were not counted as killed mutants.

## Mutants and released handles

| Mutant | Execution | Catch |
| --- | --- | --- |
| Function reverses global classification | Exit 0, empty stderr | Go finding bytes differ at 1,411 |
| Nonconstructor reverses global classification | Exit 0, empty stderr | Go finding bytes differ at 61 |
| Wrapper reverses global classification | Exit 0, empty stderr | Go finding bytes differ at 577 |
| Raw declaration-file bits forced false | Exit 0, empty stderr | Go finding bytes differ at 61 |
| Released registry keeps the handle | Exit 0 instead of required panic 70 | Released-handle assertion |
| Filtered Node oracle changes one byte | Oracle rejects differing output | TestTheOracleCatchesOneByte |

Every rule and raw-question mutant is caught solely by complete diagnostic bytes. A valid identifier question on a released program panics 70 with `invalid or released checker handle`; a test-only Go overlay retaining the handle makes the same probe succeed. All mutations use owned scratch copies or overlays, with no shared source edits.

## Quiet timings

Three alternating native/Go rounds per corpus after builds and tests finished; every round compares complete output bytes. The existing benchmark driver was reused without editing it. Raw phases and rounds are in [measurements.json](evidence/measurements.json).

| Corpus | Native process median | Go process median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.492s | 0.336s | 7.43x |
| Repository | 0.382s | 0.092s | 4.14x |

Compiler phase medians: native load 0.309s, run 2.050s; Go load 0.290s, run 0.034s. Repository: native load 0.081s, run 0.271s; Go load 0.078s, run 0.007s. Separate medians need not sum. Native made 79 compiler queries and 287 repository queries, including loader scope queries. Native is slower; these are whole-process observations, not isolated C crossing costs.

## Limits

The full repository gate and prior 26-rule suite were not rerun. The new lint suite was not run through the emitted-JavaScript backend. Shared harness work remains with its owner; no shared generator or harness changes were made. The pinned CLI's .a lint gap is already recorded in the first batch report.

Literal reference rows are replayed under the default checker configuration, not the complete upstream fixture harness and configuration matrix. Arbitrary projects, JSX and JavaScript populations are not established by this evidence. This comparison runner neither applies edits nor implements lint suppression. No pull request is opened.
