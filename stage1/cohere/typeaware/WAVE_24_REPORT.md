Built: prefer-return-this-type, return-await and use-unknown-in-catch-callback-variable as native Adamic .a ports.
Commits: claim bd26db5b9c19a500c1fa9a1988ff9ee724f4a680; implementation is the commit containing this report.
Checks: 24 controls, 43 findings; exact diagnostics and sanitizer runs on 77 compiler and 287 repository roots; targeted bridge, checker, refusal and Node gates.
Mutants: three rule predicates, three raw checker questions and released registry; byte oracle or required refusal catches each compiled mutant; existing ownership and frame mutants also pass.
Uncovered: nondefault rule options, arbitrary TypeScript/JSX inputs, the full repository gate and a complete production cohere CLI lint pass.

## Selection and source pins

Branch `codex/typeaware-wave-24` starts at the requested bridge tip
`0d540f413625f016f20fea39761c7b184f335de6`, rather than main. Read CLAUDE.md,
README.md, docs/0.1.md, docs/memory.md, the type-aware README, VOLUME_REPORT and
COVERAGE_REPORT before implementation. The specific requested bridge base takes
precedence over the general instruction to start from main.

Fetched all 268 origin refs and checked competing claims and named ports.
No selected rule was already claimed or ported, so none was skipped. The claim
was committed and pushed before source implementation. Positions 70–72 are the
checker-only combined-volume ranking after excluding the 26 existing ports,
using VOLUME_REPORT's linked compiler-all.counts and repository-all.counts,
descending totals and lexical ties. All three selected rules have zero recorded
findings in both corpora. The claim records the distinction from the separate
inventory ranking, which includes non-checker rules.

Cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
Its typescript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
Compiler corpus: TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`.
The portable root manifests are exact copies of validation-coverage's frozen
77 compiler and 287 repository roots. Configured declaration roots remain in
both checker programs. The repository manifest contains 212 .a and 75 .ts files;
it does not grow to include the ports being built. Source hashes are saved in
[validation-wave-24](validation-wave-24).

## Native decisions and bridge fields

Each rule has its own .a file. `wave_24_suite.a` parses each file once, builds the
existing parent index, runs the three classes, sorts complete diagnostic lines,
and releases one checker program after all files. Findings include UTF-8 ranges,
rule/message IDs, exact text, all fixes and all suggestions with their complete
edit ranges and replacement text. The test does not apply edits or suppressions.

`prefer_return_this_type.a` selects class-name return annotations, walks returned
expressions without descending into nested functions, and decides whether class
instance or receiver identities permit replacing the annotation with `this`.
`return_await.a` uses the production default `in-try-catch`, decides await edits
and suggestions from return contexts, thenability, try/catch/finally and using
scopes, and preserves comments and required parentheses.
`use_unknown_in_catch_callback_variable.a` selects catch/then rejection handlers,
decides parameter safety from raw signature fields, and emits exact unknown
annotation suggestions or destructuring messages.

The three new questions have matching Go and Adamic files:

| Question | Fields after common version/question header |
| --- | --- |
| class-this-types | Class instance identity, class receiver identity (zero if absent) |
| then-callback-signatures\nTYPE_ID | Count followed by call-signature counts of the first callback parameter of each then signature; apparent types, unions and rest element types are queried |
| handler-parameter-types | Signature count; per signature: parameter count, first type present, raw type flags, rest boolean, array/tuple boolean, type-argument count, first argument's raw flags |

These are raw checker answers, not Go lint verdicts. Existing framing and native
Frames validation retain canonical integer, boolean, length and trailing-data
checks. Type identities are borrowed from the live program and are validated
before lookup. `wave_24_questions.go` routes these questions and delegates older
questions unchanged. The only existing shared source edit is one line in
`bridge/tsgo/archive/main.go`, registering InspectWave24. The original facts
switch and its refusal mutation anchors are byte unchanged. New suite dispatches
are one line per rule inside a new file. No protected compiler files, submodule
sources, pins or existing ports were edited. All new Adamic implementation and control files use .a.

The independent oracle `testdata/oracle_wave_24.go` is built by an overlay inside
the pinned cohere module. It constructs its own program, walks its own AST, and
calls the three unchanged production Go rules with nil/default options. It
imports no bridge code. Equality covers full diagnostic lines and file headers,
not just finding counts.

## Setup and commands

`bash cloud/setup.sh > /tmp/wave-24-setup.log 2>&1` succeeded. Its timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (78s)
setup: done in 78s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` prints 5; CPU quota is four cores. Go 1.27.1, clang 20.1.8 and Node
24.19.0. Commands sourced `/workspace/adamic-tools/env.sh`. Test processes write
stdout/stderr directly to files; no test output is piped.

Final agreement command (the absolute manifests resolve the committed portable
manifests against the indicated source roots):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE24_ARTIFACTS=/workspace/wave-24-final-complete \
ADAMIC_WAVE24_REPOSITORY_MANIFEST=/workspace/wave-24-repository.manifest \
ADAMIC_WAVE24_COMPILER_MANIFEST=/workspace/wave-24-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus \
go test ./stage1/cohere/typeaware -run '^TestWave24AgreementAndMutants$' \
  -count=1 -v > /tmp/wave-24-all-final.log 2>&1
```

Final combined gate PASS 74.854s. The generated 24 .a controls have 43 findings and 9,311 identical serialized
bytes, including file headers whose lengths depend on the artifact path. Both
native normal and ASan builds match Go. Controls exercise class returns and
unions, nested functions and arrow properties, try/catch/finally and using,
unknown/any/unconstrained parameters, comments and parentheses, computed method
names, spread calls, structural thenables, rest tuples/arrays, destructuring,
optional/default parameters, conditional/comma callbacks, Unicode and CRLF.

Final corpus result: compiler 77 roots, zero findings, 5,010 identical bytes;
repository 287 roots, zero findings, 18,485 identical bytes. Both normal and
ASan/UBSan/LeakSanitizer builds match independently loaded Go programs, with empty
native stderr. Positive controls prevent mistaking those zero-finding corpora for
proof that a missing rule implementation works.

Additional gates:

```sh
go test ./bridge/tsgo/... -count=1 -v > /tmp/wave-24-bridge-final.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave-24-checker-final.log 2>&1
go test ./stage1/cohere/typeaware \
  -run 'TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags' \
  -count=1 -v > /tmp/wave-24-facts-final.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave-24-vet-final.log 2>&1
```

Bridge PASS 61.800s; checker PASS 0.121s (also 0.190s inside the bridge gate);
frame/refusal/flags PASS 35.960s; vet and gofmt output empty. Direct checker tests
pin the three used flags, verify class identities against the checker, require
positive Promise signatures, check exact unknown parameter fields and reject
malformed or mismatched questions/identities. The bridge compares 1,600 positions
across four files, 54,982 bytes under sanitizers.

The filtered internal/oracle gate runs TestTheOracleCatchesOneByte and native
versus Node subtests for functions, closures, generic_functions, lone_surrogates,
maps_and_text, string_index, method_closures and sorting. PASS 18.473s; saved Node
log names every selected subtest. No full `go test ./...` run is claimed.

## Every mutant and its catcher

All six new rule/question mutants compile and exit 0 with empty stderr. Their
only failure is changed bytes compared with unchanged production Go findings.
Offsets below are zero based in the final 24-control serialization.

| Mutant | Change | Catcher / first difference |
| --- | --- | --- |
| receiver | Accept class returns as well as receiver returns | Go byte oracle, 154 |
| await | Require await outside handling instead of inside | Go byte oracle, 1,758 |
| unknown | Reverse the unknown first-parameter exemption | Go byte oracle, 4,060 |
| class-question | Increment receiver identity | Go byte oracle, 207 |
| then-question | Return an empty signature-count list | Go byte oracle, 1,610 |
| handler-question | OR unknown into raw first-parameter flags | Go byte oracle, 4,060 |
| released-registry | Keep a released program in the live registry | Mutant exits 0; expected panic 70 and exact invalid/released-handle text are absent |

The unmutated released handler query panics 70 with
`adamic: panic: invalid or released checker handle`. Earlier class and handler
question mutation attempts left Go locals unused and failed compilation; those
are saved separately and do not count as caught mutants. An early dispatcher
change broke an existing refusal mutation anchor; restoring the original
switch and registering the new wrapper resolved that regression. A standalone
corpus-check attempt incorrectly rejected Go's intentional timing stderr; the
final harness permits that timing stream and compares diagnostic stdout.

Existing bridge gate mutants also ran against the final registration:

| Mutant | Catcher |
| --- | --- |
| Input view length +1 | ASan heap-buffer-overflow |
| Output string length +1 | ASan heap-buffer-overflow |
| Retain released handle | Stale-handle assertion |
| Query source-file type instead of exact node | Independent byte oracle at byte 6 |
| Remove checker link opt-in guard | Required build/C/JS refusal |
| Omit C output free | LeakSanitizer |
| Allocate region result on heap | LeakSanitizer |
| Remove exact-kind refusal | Required panic 70; mutant exits 0 |
| Permit unknown question | Required panic 70; mutant exits 0 |
| Change one Node-oracle byte | Native/Node output comparison |

The 19 wire mutants are empty, length-negative, length-short, trailing, version,
question, not-integer, rounded-integer, noncanonical-integer, negative-natural,
bad-boolean, zero-id, negative-tuple, missing-root, missing-constraint,
missing-element, duplicate-id, missing-link-17 and missing-link-18. Every one
panics 70 with the specific frame guard, recorded in facts.log. These prove
shared decoder guards; they do not claim exhaustive independent mutations of
every new question field.

## Native time against Go

After build/test activity stopped, used the existing count benchmark for three
alternating rounds. Complete subprocess time includes loading, parsing,
walking, checker questions and process overhead; no build time is included.
Count output is equal each round. Raw phase stderr and all samples are saved.

```sh
python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/wave-24-final-complete/wave24 \
  /workspace/wave-24-final-complete/wave24-oracle /workspace/wave-24-timings \
  --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-24-repository.manifest \
  --corpus compiler /workspace/wave-24-corpus/src/compiler/tsconfig.json /workspace/wave-24-compiler.manifest \
  --rounds 3 > /tmp/wave-24-timing-final.log 2>&1
```

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository, 287 roots | 0.212213s | 0.076679s | 2.77x |
| Compiler, 77 roots | 1.484308s | 0.278337s | 5.33x |

Observation: native is slower for both populations. The repository makes zero
checker queries; compiler makes eleven. These zero-finding workloads measure
load/parse/walk costs as well as bridge calls, and establish no general speedup
or isolated rule cost. Default-rule parity, rather than optimization, is this
unit's result.

## Limits and evidence

Production cohere CLI `--no-cache --no-fix --lint` refuses all seven .a paths as
not TypeScript or JavaScript before linting. A scratch Go loader overlay admits
.a roots but fails to resolve their .a imports; its 43 diagnostics include
unresolved-type warnings and style warnings. Neither attempt is a successful
production CLI lint gate, and neither changes the pinned cohere sources. Both
logs are retained. Native Adamic does resolve, type-check, compile and execute
the .a source imports. No new Adamic .ts source was written to bypass the gate.

Default options only, matching the existing suites. Nondefault return-await
modes are not implemented. No guarantee over arbitrary inputs follows from the
frozen corpora and controls; parser coverage and JSX retain the base limits.
There is no full CLI display/fix-application/suppression parity claim, no whole
repository test gate, and no claim to reproduce all cohere's upstream rule tests.

[validation-wave-24](validation-wave-24) contains logs, portable manifests,
source SHA-256 maps, compressed canonical Go/native/sanitized diagnostics and
timing samples. Binaries, C archives and scratch overlays are not committed.
The final implementation commit is pushed on the requested branch; no PR.
