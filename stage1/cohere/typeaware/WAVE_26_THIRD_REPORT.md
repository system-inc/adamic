Built: native .a ports of no-eval, no-extend-native and no-func-assign, with isolated raw questions.
Commits: claim 492c03bf; implementation e903d658; evidence is committed separately.
Commands and outputs: rule suite PASS 172.997s, bridge PASS 91.531s, Node oracle PASS 23.547s.
Mutants: all three rules and both questions caught only by finding bytes; registry, ABI and Node mutants caught too.
Not covered: full gate, all upstream/options fixtures, emitted JavaScript rule runner; pinned CLI refuses .a.

# Wave 26 third batch

Both prior batches were complete and pushed at a21c17fb before claiming this
batch. There are now nine completed rule ports on this branch. This batch's claim
was pushed before implementation, and no additional rules are claimed here.

## Selection

Fetched all origin branches with `git fetch --prune --no-recurse-submodules origin
'+refs/heads/*:refs/remotes/origin/*'`. The selection inspected 330 remote refs,
excluded ports on origin/codex/tsgo-c-library at 5afbdb83 and origin/main at
ef3d907e, and conservatively excluded 108 ranked names mentioned in Markdown
claim files anywhere on origin. The full ref and claim-blob audit is in
[validation-wave-26-third/selection.json](validation-wave-26-third/selection.json).
There were 64 available rules before this claim. The first three by combined
compiler/repository volume descending and lexical ties were no-eval,
no-extend-native and no-func-assign, each with zero findings in both ranked corpora.
A separate git grep on base and main found these names only in skipped-rule
frequency evidence and the checker-dependent counter's eligibility list, not ports.

## Implementation

Each rule has its own `.a` file. `core_access.a` performs the production structural
write climb, including the rest-element workaround, and ports the compiler's
access-kind semantics in Adamic rather than asking Go for a lint judgment.
It also distinguishes static member names from dynamic subscripts and recognizes
all assignment operators. The raw questions and their separate Go/Adamic files
are documented in [WAVE_26_THIRD_FACTS.md](WAVE_26_THIRD_FACTS.md).

The eval rule reports direct callees syntactically even when a local shadows eval,
uses actual resolution for bare value references, excludes declaration/property
name positions, and handles the production global-object chains and wrappers.
The prototype rule carries the unchanged production set of 49 builtin spellings,
checks declaration origin for the receiver and Object.defineProperty callee,
accepts all assignments, and reports the full assignment or call span.
The function assignment rule anchors the first local declaration resolved from
each function name, then matches writes by the same declaration identity. It
handles hoisting, overload/merged declaration order, nested shadows and shorthand
value symbols without double counting a same-kind shadow.

`binding-origin` exposes ordinary and shorthand symbols' ordered declarations
and actual SourceFile.IsDeclarationFile flags. This is the exact Go global helper's
criterion, not the stricter default-library criterion from the previous batch.
`binding-structure` supplies complete raw syntax with initializer, element-access,
loop and unary roles. Like the second batch, this runner consumes the bridge's raw
AST projection instead of the Adamic parser. All structural judgments and findings
remain native Adamic. The Go oracle independently loads and walks the program and
calls unchanged production registry rules; it imports no bridge implementation.

Only two single-line switch registrations were added to shared facts.go. No shared
registration generator, generic lint harness, profile compilation, serializer or
parser was changed, and no protected compiler implementation was touched.
The new runner and batch test file use the existing type-aware test helpers.
No shared-harness gap blocked this batch's native comparison. The work on
codex/lint-harness-dot-a was inspected, but its shared files were left untouched.

## Agreement and controls

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Generated controls | 44 | 45 | 28705 |
| Frozen repository | 287 | 0 | 18485 |
| TypeScript src/compiler | 77 | 0 | 5318 |

All three populations match complete production Go output under ordinary native
execution and ASan/UBSan/LeakSanitizer. Each canonical row serializes findings,
fixes and suggestions, preserving duplicate findings and edit order. These rules
emit no fixes or suggestions: all tested rows have zero of both. Positive controls
produce 15 no-eval findings, 9 no-extend-native findings and 21 no-func-assign
findings, rather than relying on the zero-finding corpora as sole evidence.

Controls cover direct, indirect, optional and shadowed eval; parentheses,
assertions and non-null wrappers; computed/template names; global-object chains;
value references and excluded name positions; direct and defineProperty prototype
edits; optional access; logical assignments; local builtin shadows and benign
updates; function declarations, expressions, variables and arrows; writes before,
after and inside declarations; nested same-kind shadows; property writes versus
binding writes; explicit and shorthand destructuring; default-value reads;
array/object rest, parenthesized rest and loop targets; unary/compound writes;
Unicode and CRLF byte spans. Text and config are preserved in controls.json.

The raw contract test compares both symbol accessors' declaration order, source,
kind and spans directly with the compiler, including actual declaration-file flags.
It checks the raw syntax population and malformed request rejection. The full
checker package passed in 0.350s separately and 0.292s in the bridge regression.
All output hashes and compressed canonical streams are preserved beside the
portable frozen manifests and source hashes. Original absolute path headers are
retained; moving the corpus changes those headers. The frozen corpus excludes
these new implementation files, as it did for selection and both previous batches.

## Mutants

Every rule and question mutant compiles, exits 0 and has empty stderr. Only the
independent production finding comparison detects it:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| no-eval | Remove window from recognized global-object names | 3817 |
| no-extend-native | Accept only equals, losing compound/logical assignments | 11247 |
| no-func-assign | Resolve shorthand through the ordinary property symbol | 18228 |
| binding-structure | Return logical-or for binary operator metadata | 9277 |
| binding-origin | Clear declaration-file flags | 1290 |

A binding-origin request after program release must panic 70 with
`invalid or released checker handle`. Keeping the handle in the registry makes
it exit 0; the required refusal catches that separate mutant.

The full bridge regression again caught input and output lengths off by one with
ASan heap-buffer-overflow; retained released handles with the stale-handle
assertion; a source-file type instead of the queried type with byte comparison at
byte 6; removed link opt-in with refusal; omitted C-buffer frees with LeakSanitizer;
and a region result allocated on the heap with LeakSanitizer. It checked 100 C ABI
queries, output strings surviving release, zero/stale and distinct handles, and
162 independently queried positions / 3261 identical bytes under sanitizers.
The Node oracle caught its one-byte mutant and passed eight selected native,
JavaScript and Node cases. These are compiler regressions, not a JavaScript run
of this checker-linked rule runner.

## Native time against Go

After all builds and regression tests finished, three isolated alternating rounds
ran the complete default rule set with count-only output. All rounds agreed on
zero corpus findings. Medians below are seconds; separate phase medians need not
sum to the process median.

| Corpus / implementation | Load | Run | Whole process |
| --- | ---: | ---: | ---: |
| Compiler native | 0.357589 | 6.817689 | 7.224246 |
| Compiler Go | 0.339482 | 0.152471 | 0.507919 |
| Repository native | 0.101241 | 0.661501 | 0.772037 |
| Repository Go | 0.090350 | 0.073796 | 0.182675 |

Native is about 14.22 times Go's compiler process time and 4.23 times its
repository process time. Native issues 9267 compiler queries and 920 repository
queries, including one complete syntax projection per root and declaration-origin
requests. Complete AST transport and decoding add substantial cost. The zero
corpora exercise traversal and declaration anchoring, while positive controls
exercise the rule judgments. No Go-speed parity claim is made. Raw rounds and
phase stderr are in measurements.json and benchmark.txt.

## Commands and environment

The same toolchain from the first batch is used, sourced from
`/workspace/adamic-tools/env.sh`: Go 1.27.1, clang 20.1.8, Node 24.19.0 and
`nproc` 5. Original setup timing lines: Go ready 0s, clang/Node ready 1s,
submodules 1s, build cache warm 135s, done 135s. The original setup log is
validation-wave-26/setup.txt. No submodule pins were changed. All test output
went directly to logs, without pipes.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE26_THIRD_ARTIFACTS=/workspace/wave-26-third-validation \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave26ThirdAgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave-26-third-test.log 2>&1
# PASS 172.997s.
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave-26-third-checker.log 2>&1
# PASS 0.350s.
TMPDIR=/workspace/wave-26-bridge-scratch go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /tmp/wave-26-third-bridge.log 2>&1
# PASS bridge 91.531s, checker 0.292s.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout 10m > /tmp/wave-26-third-node.log 2>&1
# PASS 23.547s, eight selected cases and one-byte mutant.
go vet ./... > /tmp/wave-26-third-vet.log 2>&1
# Exit 0, empty output.
# gofmt -l cmd internal and all new Go files: exit 0, empty output.
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-26-third-validation/wave26 /workspace/wave-26-third-validation/wave26-oracle /workspace/wave-26-third-bench --corpus compiler /workspace/wave-26-typescript/src/compiler/tsconfig.json /workspace/wave-26-compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-26-repository.manifest > /tmp/wave-26-third-bench.log 2>&1
```

## Limits

The full repository gate and all earlier type-aware suites were not rerun. The
touched bridge packages, this complete default-rule differential suite and filtered
Node oracle were run. Nondefault allowIndirect/exceptions options are exposed by
constructors but are not compared by this fixed default runner. All upstream
fixture permutations, arbitrary JavaScript/JSX populations, suppression and edit
application are not covered. These rules have no edits to apply.

The raw AST projection means this batch does not prove Adamic's independent parser
on the corpus. The JavaScript backend has no checker bridge, so emitted JavaScript
for this rule runner was not compared. No shared files were changed to work around
that limit. The pinned `/workspace/wave-26-cohere --no-cache --no-fix` command over
the seven new `.a` files exits 1 and refuses them as not TypeScript or JavaScript
inputs, as recorded in cohere-refusal.txt. That is not a lint pass; the native
rule comparison and sanitizer checks above are complete despite that CLI limit.
