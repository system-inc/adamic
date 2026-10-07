Built: three default-option native Adamic rule ports and two raw checker questions.
Commits: claim `dda5edca`; implementation is the commit containing this report.
Commands: wave PASS 121.092s, bridge PASS 95.871s, Node oracle PASS 24.213s; vet and gofmt clean.
Mutants: delete predicate, stringification certainty, class constructor-slot equality caught only by finding bytes; released-registry mutant caught by required panic.
Not covered: nondefault options, full repository gate, all upstream fixtures, JSX populations, or the pinned cohere CLI's unsupported `.a` source gate.

# Wave 14

Branch `codex/typeaware-wave-14` starts at
`0d540f413625f016f20fea39761c7b184f335de6`, origin/codex/tsgo-c-library.
The claim was committed and pushed before implementation. Combined compiler and
repository counts linked by VOLUME_REPORT.md were sorted by descending volume,
then lexical name, excluding the 26 ports on that branch. Positions 40, 41 and
42 are `@typescript-eslint/no-array-delete`, `@typescript-eslint/no-base-to-string`
and `@typescript-eslint/no-extraneous-class`, each compiler 1 / repository 0.
No selected rule was skipped. All fetched origin branch claim trees and Adamic
source rule-name occurrences were searched; the only other source match was a
configuration set listing, not a port. The initial checkout fetched main alone;
an explicit all-branch fetch retrieved the requested base.

Each rule owns its `.a` file. `wave_14.a` loads one checker program, parses each
root once, builds parent links, runs the three rules, sorts complete diagnostic
lines and releases the program. It reuses the existing protocol and parser.
Array deletion preserves the three surgical suggestion edits, including trivia.
Stringification uses the production three-valued union/intersection/array/tuple
recursion and default ignored types. Empty classes follow symbol identities,
type positions, exports, superclass references and contextual constructor slots.
No lint verdict, diagnostic text or proposed edit comes from Go.

`stringification-type` supplies type flags, display and generic declaration names,
constraint/number-index/union/tuple/base links, computed method names and origins,
and coercion methods' declaration parents. `reference-context` supplies symbol
and type identities, type-position classification and contextual construct
signature counts. Both have separate Go and Adamic files; the only existing
source edit adds their two switch dispatch entries in checker/facts.go.
No protected compiler files, old rules, submodule pins or shared native files
were edited. The existing tsgo_inspect ownership and handle contract is unchanged.

The independent oracle in `testdata/oracle_wave_14.go` is built through a cohere
module overlay. It loads its own program and walks its own AST, invoking the
three unchanged production rule implementations through the registry and
production program views. It imports no bridge code. The frozen populations
retain configured declaration roots and `.a` loading. TypeScript is v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`; cohere and typescript-go retain their
base-branch pins. Manifests, source hashes and complete diagnostic stream hashes
are in `validation-wave-14`.

| Population | Files | Findings | Identical bytes, Go / native / sanitized native |
| --- | ---: | ---: | ---: |
| Controls | 17 | 33 | 5,308 |
| Compiler | 77 | 3 | 5,794 |
| Frozen repository | 287 | 0 | 18,485 |

Headers and the summary explain the nonzero repository bytes. Controls require
at least one positive finding from each rule, so zero findings on the repository
are not treated as evidence that the instruments ran by themselves. Controls
cover constrained and composite array receivers, holes versus object deletion,
comments, parentheses, comma indexes, Unicode/CRLF, useful and base coercions,
recursive arrays, join certainty, ignored generic error bases, Symbol.toPrimitive,
String shadowing, tagged templates, empty/constructor/static/instance classes,
parameter properties, constructors as tokens, exports, reassignment, inheritance,
implements, static blocks and named/unnamed class expressions.

All three rule mutants compiled and exited 0 with empty stderr. Only the
independent Go diagnostic comparison caught them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| delete | Reverse the underlying-array predicate | 48 |
| stringify | Render Object fallback certainty as `may` instead of `will` | 1,501 |
| class | Reverse equality between constructor-slot count and value-use count | 4,250 |

The released-handle probe exits 70 with `invalid or released checker handle`.
Retaining the released program in the registry makes that probe exit 0, and the
required-panic expectation catches it. The bridge regression checks 100 C ABI
queries, output buffers surviving release, distinct handles, stale/zero handles,
and 162 independently answered positions / 3,261 identical bytes under
ASan/UBSan/LeakSanitizer. Its input/output length mutants trigger ASan; missing
output frees and a region result allocated on the heap trigger LeakSanitizer;
wrong-position, retained-handle and removed-link-guard mutants fail their
independent expectations. ASan instruments C/native ownership, not Go's heap.
The direct checker tests also verify the new metadata and reject malformed IDs
and unexpected question suffixes.

## Timing observations

Three alternating count-only rounds, after builds and tests finished. Each run
creates and releases one program. Values are medians in seconds. Native timing
includes the parser, native judgments and bridge; Go invokes production rules
directly. These are whole-process observations, not isolated crossing costs.

| Corpus | Native process | Go process | Native / Go | Native load | Native run | Go load | Go run |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 1.987 | 0.458 | 4.34x | 0.258 | 1.705 | 0.271 | 0.172 |
| Repository | 0.522 | 0.167 | 3.13x | 0.062 | 0.455 | 0.061 | 0.093 |

Native makes 14,912 compiler and 16,144 repository queries. Median aggregate
adapter query intervals are 0.429s and 0.194s. These include checker work,
serialization, C transfer and buffer conversion/freeing, excluding Adamic's
frame decoding. No speed-parity claim is made.

## Commands and limits

`bash cloud/setup.sh` passed: Go, clang, Node and submodules ready at 0s;
build cache warm 82s; done 82s. `nproc`: 5. cpu.max: `400000 100000`.
Go 1.27.1, clang 20.1.8, Node v24.19.0. Every Go command sourced
`/workspace/adamic-tools/env.sh`. Test output went directly to log files.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE14_ARTIFACTS=/workspace/wave-14-artifacts \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript \
ADAMIC_WAVE14_COMPILER_MANIFEST=/workspace/wave-14-artifacts/compiler.manifest \
ADAMIC_WAVE14_REPOSITORY_MANIFEST=/workspace/wave-14-artifacts/repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave14AgreementAndMutants$' \
  -count=1 -v -timeout 30m > /workspace/wave-14-artifacts/test-full-2.log 2>&1
# PASS 121.092s

go test ./bridge/tsgo/... -count=1 -v -timeout 30m \
  > /workspace/wave-14-artifacts/bridge.log 2>&1
# bridge PASS 95.871s; checker PASS 0.300s

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout 30m -v > /workspace/wave-14-artifacts/node-oracle.log 2>&1
# PASS 24.213s; eight selected fixtures including method_closures and generic_functions

go vet ./... > /workspace/wave-14-artifacts/vet.log 2>&1
# exit 0, empty log
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware
# empty output

python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/wave-14-artifacts/wave-14 /workspace/wave-14-artifacts/wave-14-oracle \
  /workspace/wave-14-artifacts/bench \
  --corpus compiler /workspace/wave-14-typescript/src/compiler/tsconfig.json /workspace/wave-14-artifacts/compiler.manifest \
  --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-14-artifacts/repository.manifest \
  > /workspace/wave-14-artifacts/bench.log 2>&1
```

The first controls run exposed a wrong decoder for an existing symbol-origin
question; it was replaced with the documented resolved-name origin frame.
The first complete repository run refused a named-type frame around the
long-literal fixture. Concatenation now requests unnamed raw shapes, and the new
question normalizes display strings to valid UTF-8 before framing them. The
unchanged corpus then passed normally and under sanitizers. Earlier failures
are not counted as passing validation.

The pinned cohere command built, but its dry type/lint source gate rejected the
six new `.a` paths as outside its program. Its format-only command returned 0
without demonstrating that those files were formatted, so that is not claimed
as source-gate evidence. Native compilation checks all new sources under the
Adamic compiler's strict options. All authored Adamic sources and generated
controls/probes are `.a`; no new Adamic `.ts` source was authored.

The full `go test ./...` gate and the old 26-rule corpus suite were not rerun.
The targeted wave suite, all bridge packages, Go vet and the filtered independent
Node oracle are the gate used here. Nondefault rule options, all upstream fixture
matrices, JSX populations, suppression processing and applying suggested edits
are outside this unit's measured scope. Complete proposed fixes/suggestions on
the requested populations are covered. No pull request was opened.
