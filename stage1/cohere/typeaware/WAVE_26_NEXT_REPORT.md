Built: three more native Adamic rules, isolated raw questions and a batch runner.
Commits: claim 815e4759; implementation 9dede6c8; this evidence is committed separately.
Commands and outputs: agreement PASS 149.515s, bridge PASS 87.984s, Node oracle PASS 21.325s.
Mutants: three rules and four questions caught by finding bytes; registry, ABI and Node mutants caught too.
Not covered: full repository gate, all upstream fixtures, rule JavaScript backend; cohere CLI refuses .a.

# Wave 26 next batch

The first batch was complete and pushed as 8035ca2f before the follow-up request.
The claim for this batch was committed and pushed before writing implementation.
No further rules are claimed. Ahra's correction arrived while validating this
already-claimed batch. No shared registration generator or test harness was
edited. No protected native, lowering or oracle implementation was edited.

## Selection and implementation

Fetch used `git fetch --prune --no-recurse-submodules origin
'+refs/heads/*:refs/remotes/origin/*'`. All 321 origin remote refs were inspected.
The linked compiler-all.counts and repository-all.counts tables contain 197
checker-dependent rules, ranked by combined volume descending and lexical ties.
Existing ports on origin/codex/tsgo-c-library at 5afbdb83 and origin/main at
ef3d907e, and rule names reserved in Markdown claim files on every origin branch,
were excluded. The 26 baseline ports include method-signature-style, which is not
in the checker-dependent ranking; 25 ranked names are therefore excluded as ports.
The selection snapshot records every inspected ref and distinct Markdown claim
blob, including reports that conservatively exclude mentioned rule names.
The first three available were:

| Rule | Compiler | Repository |
| --- | ---: | ---: |
| nexus/correctness-no-global-listener-target-assertion | 0 | 0 |
| nexus/correctness-no-leaked-number-render | 0 | 0 |
| nexus/correctness-no-mock-on-module-namespace | 0 | 0 |

Each lives in its own `.a` file. New native sources all use `.a`. The Go oracle
has its own loader and walk, calls the unchanged production registry rules and
imports no bridge implementation. Sorting preserves duplicate diagnostics and
serializes every finding, fix and suggestion in the existing canonical format.
These production rules emit no fixes or suggestions, so all tested fix and
suggestion counts are zero, including the positive controls.

The existing Adamic parser does not support JSX. Rather than modifying shared
parser files, this batch uses the raw `node-structure` syntax projection described
in [WAVE_26_NEXT_FACTS.md](WAVE_26_NEXT_FACTS.md). This is a departure from the
previous wave's Adamic parser walk. All candidate selection, ancestry traversal,
type classification, subtype judgments, message construction and finding selection
remain in native Adamic. The bridge returns complete raw AST structure, symbol
and declaration ancestry, literal Stringer values, and transformed type graphs.
Go and Adamic question files have matching names. The shared facts.go switch has
only four registration lines added, one per question. No verdict, rule message,
edit or rule-specific candidate list is computed on the Go side.

The listener rule follows symbols for the event, resolves inline and same-file
function handlers, requires actual default-library declarations, traverses base
interfaces, and exempts assertions accepted by non-nullable assignability. The
render rule follows rendered and falsy values through logical and conditional
expressions, classifies union parts, numeric literals and branded numbers, and
follows type parameter constraints with the production depth bound. The mock rule
uses namespace import declarations, type-only metadata, resolved signature ancestry,
and the actual emitted module kind rather than matching receiver names alone.

## Agreement and sanitizers

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Generated controls | 43 | 26 | 14759 |
| Frozen repository | 287 | 0 | 18485 |
| TypeScript src/compiler | 77 | 0 | 5318 |

Each row was compared with independent production Go under both ordinary native
execution and ASan/UBSan/LeakSanitizer. Complete output hashes and compressed
canonical streams are in [validation-wave-26-next](validation-wave-26-next).
Original absolute path headers are retained; relocation changes those bytes.
Both corpora are the same frozen populations used in the first batch, not a corpus
expanded to include the new ports.

Controls include local and default-library receivers, direct and nested handlers,
const versus mutable named handlers, checked versus unchecked targets, general
and custom element interfaces, unions, assertion wrappers, nested parameter shadowing,
Unicode and CRLF, number/bigint/zero/nonzero literals, nullable and mixed unions,
never/void, intersections, constrained generics, JSX attributes versus children,
logical chains and conditional branches. JSX inputs are external TypeScript `.tsx`
oracle fixtures generated in scratch, not Adamic implementations. Their text is
preserved in controls.json alongside the `.a` control inputs and ambient declarations.

Additional module comparisons matched Go and sanitized native execution:
CommonJS and NodeNext each have 21 findings / 12098 bytes; Preserve has
26 / 14759. A NodeNext manifest with the same mock call in `.mts` and `.cts`
has exactly one finding / 667 bytes, on the ES module. These are recorded in
module-edges.txt and corresponding canonical hashes.

The checker contract test checks the syntax node population, emitted module kind,
symbol metadata, non-nullable and constraint type identities against the compiler,
malformed selectors and identities, and pinned TypeFlags/SymbolFlags/NodeFlags.
It passed with the full checker package. An initial test assumption that a
primitive union has no base constraint was wrong; the test now compares directly
with GetBaseConstraintOfType rather than guessing that result.

## Mutants

All seven judgment mutants compile and exit 0 with empty stderr. Only comparing
complete finding bytes with production Go catches them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| Listener rule | Treat HTMLElement as a specific element | 1360 |
| Render rule | Stop recognizing zero literals as falsy numbers | 6538 |
| Mock rule | Accept type-only namespace imports | 13974 |
| node-structure | Return logical-or operator kinds for binaries | 5684 |
| declaration-lineage | Clear actual default-library classification | 60 |
| literal-string | Return 1 for a zero literal | 6538 |
| transformed-shape | Return the original parameter instead of its constraint | 10023 |

The new released-handle probe uses declaration-lineage after release and requires
panic 70 with `invalid or released checker handle`. Retaining the released handle
in the registry makes the probe exit 0; the required refusal catches that mutant.

The bridge regression suite independently caught all its existing foundation
mutants: input length +1 and output length +1 through ASan heap-buffer-overflow;
retained released handles through the stale-handle assertion; a source-file type
instead of the queried type through the byte oracle at byte 6; removed link opt-in
through refusal; omitted C-buffer freeing through LeakSanitizer; and a region
allocation moved to the heap through LeakSanitizer. It checked 100 C ABI queries,
zero/stale and nonreused handles, returned strings surviving program release, and
162 independently queried positions / 3261 identical bytes under sanitizers.
The filtered Node suite also caught its one-byte mutant and passed eight selected
native/JavaScript/Node cases, including string indexing, Unicode, closures and sorting.

## Native against Go

Three alternating, isolated count-only rounds ran after builds and tests finished.
Medians are seconds. Both corpora have zero findings for these rules, so these
measurements primarily describe loading, walking and transporting syntax rather
than the positive type judgments exercised by controls. Separate phase medians
need not sum to the process median.

| Corpus / implementation | Load | Run | Whole process |
| --- | ---: | ---: | ---: |
| Compiler native | 0.314165 | 4.282442 | 4.754985 |
| Compiler Go | 0.314270 | 0.044601 | 0.381352 |
| Repository native | 0.092978 | 0.545717 | 0.640604 |
| Repository Go | 0.086900 | 0.063217 | 0.171108 |

Native is about 12.47 times Go's compiler process time and 3.74 times its
repository process time. Native makes one complete syntax question per root here,
77 compiler and 287 repository queries. Transporting and decoding full syntax
adds substantial cost; this work makes no speed-parity claim. Raw alternating
rounds and phase stderr are preserved in measurements.json and benchmark.txt.

## Commands and environment

This is the same cloud toolchain installed for the original batch with
`bash cloud/setup.sh`: Go ready 0s, clang/Node ready 1s, submodules 1s, build-cache
warm 135s, done 135s. Its original timing log is in validation-wave-26/setup.txt.
The environment is sourced from `/workspace/adamic-tools/env.sh`; `nproc` remains
5. Go 1.27.1, clang 20.1.8 and Node 24.19.0. No submodule pins changed.
All test output went directly to log files without pipes.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE26_NEXT_ARTIFACTS=/workspace/wave-26-next-validation-final \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave26NextAgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave-26-next-test-final.log 2>&1
# PASS 149.515s. All rules, raw questions, corpora, sanitizer comparisons and released registry mutant.
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave-26-next-checker-final.log 2>&1
# PASS 0.206s.
TMPDIR=/workspace/wave-26-bridge-scratch go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /tmp/wave-26-next-bridge.log 2>&1
# PASS bridge 87.984s, checker 0.205s.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout 10m > /tmp/wave-26-next-node.log 2>&1
# PASS 21.325s, eight selected cases plus one-byte mutant.
go vet ./... > /tmp/wave-26-next-vet.log 2>&1
# Exit 0, empty output.
# gofmt -l cmd internal and all new Go files: exit 0, empty output.
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-26-next-validation-final/wave26 /workspace/wave-26-next-validation-final/wave26-oracle /workspace/wave-26-next-bench --corpus compiler /workspace/wave-26-typescript/src/compiler/tsconfig.json /workspace/wave-26-compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-26-repository.manifest > /tmp/wave-26-next-bench.log 2>&1
```

## Limits

The full `go test ./...` gate and all earlier type-aware suites were not rerun.
The touched bridge packages, this complete differential suite and filtered Node
oracle were run instead. All upstream fixture permutations, suppression/edit
application and arbitrary JSX/project configurations are not covered. The Go
oracle consumes raw TypeScript; the native runner consumes the bridge's raw syntax
projection. It does not validate Adamic's independent parser on these inputs.

The pinned `/workspace/wave-26-cohere --no-cache --no-fix` run over all eight new
`.a` sources exits 1 and refuses them as not TypeScript or JavaScript files. This
is recorded in cohere-refusal.txt, not represented as a lint pass. Shared `.a`
module support and emitted-JavaScript comparison remain with the worker on
codex/lint-harness-dot-a; no shared module-support files were changed. The native
checker bridge is not available to the JavaScript backend, so this rule runner
was not compared as emitted JavaScript. That limitation is distinct from the
passing filtered compiler Node/JavaScript regression tests.
