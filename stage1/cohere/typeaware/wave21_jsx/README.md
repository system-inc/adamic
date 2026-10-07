# Wave 21 JSX rule cores

Three native Adamic cores are implemented in their own directories:
`react/jsx-fragments`, `react/jsx-no-constructed-context-values`, and
`react/jsx-no-undef`. Each exports numeric `syntaxKinds`, has named `rule.json`
listeners, and takes the node already handed to it. The private driver dispatches
by numeric kind and fetches no root node again for another rule. No shared parser,
registration generator, harness, diagnostic model or compiler implementation was
edited by this batch. New Adamic files use `.a`.

These are **prepared-syntax cores, not completed native source ports**. The shared
parser still interprets `<` as a type assertion and cannot construct JSX nodes.
The production shared driver also lacks the prepared-syntax node roles this
batch needs. Source parsing and live checker-fact population are not connected.
`parser_prerequisites_test.go` holds three real Go positives against exact native
parser refusals, under normal and sanitizer builds. It does not count a refusal
as rule agreement. No subsequent reservations were taken.

## Decisions and checker data

The fragment core reproduces both spellings, attribute exemptions, named imports
and aliases, binding patterns, variable initializers and `require('react')`.
The undefined-tag core reproduces intrinsic/custom/namespaced/`this` exemptions,
leftmost member references, local declaration scope, globals and `.cjs` scope.
The context core includes all six cohere messages, not only upstream's four:
provider and component classification, construction/identifier following,
primitive union flags, memo dependency analysis, resolved callee bodies, async
and generator identities, nested helpers, return alternatives, holders, escape
checks and cycle/depth guards. Messages, anchor selection and quoting are native.
Printable-rune tables are generated from the pinned Go Unicode implementation.

`oracle.go.txt` calls the unchanged production Go rules through their registry,
with an independent loader and listener walk. Its separate `--prepare` mode
serializes raw AST roles, parent/child identities, token byte ranges/lines,
lexical Unicode properties, symbols and ordered declarations, type flags and
resolved signature declarations. It calls no lint-analysis helper and emits no
rule verdict. That mode is a test provider outside Adamic; it must not be reported
as native parsing or a production bridge. Both sides retain declaration roots.

Existing bridge questions already expose the needed checker primitives:
`binding-declarations`, `reference-node`, raw type shapes and
`resolved-declaration`. This batch adds no checker question or shared registration
arm. Native parser/role adaptation and production fact wiring remain missing.
The inherited bridge lifetime contract is separately rechecked with real C calls.

## Validation

216 of 219 extracted production-fixture and independent sources parse with Go.
Three rejected inputs are excluded and counted. All 216 run under defaults,
fragment `element` mode, and `allowGlobals`, with normal and ASan/UBSan/LSan
native builds. Complete byte streams include ranges, rule/message IDs and text,
fixes and suggestions, preserving duplicates and sorting full representations.
All repairs and suggestions are empty because that is production Go's behavior.

| Mode | Fragment | Context values | Undefined tag | Total | Bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| Default | 25 | 69 | 200 | 294 | 114617 |
| Element | 7 | 69 | 200 | 276 | 110365 |
| Globals | 25 | 69 | 198 | 292 | 113932 |

All six context message IDs fire. The pinned 77-file TypeScript compiler and
287-file frozen repository populations each produce zero findings for these
three rules. Go/native/sanitizer prepared-input streams match all file headers
and canonical bytes: 5087 compiler bytes and 18485 repository bytes. Those
populations contain no JSX and do not prove native JSX source support.

All 22 default batches also agree on external Node, both the original `.a`
source with types stripped by `oracle/node.mjs` and Adamic's emitted JavaScript.
This checks native/emitted behavior without using the compiler's output as the
source oracle. `compare_node.py` repeats those comparisons and records hashes
and timing observations separately.

Each real native rule mutant compiles, exits 0 and writes empty stderr. Only the
independent complete diagnostic comparison catches it:

| Rule | Mutation | First differing byte |
| --- | --- | ---: |
| JSX fragments | Reverse fragment-name acceptance | 1720 |
| JSX undefined tag | Reverse resolved-symbol acceptance | 49 |
| JSX context values | Reverse the component gate | 2336 |

The released bridge handle exits 70 with exactly
`adamic: panic: invalid or released checker handle`, normal and ASan. Retaining
the released registry entry makes the mutant exit 0 and is caught by that
expectation. Skipping native parsing makes the prerequisite mutant exit 0 and
violates the required parser refusal; this is a prerequisite mutant, not a rule
mutant. Earlier wave-21 mutants are recorded in the inherited landing reports.

## Commands and limits

After the successful setup, source `/workspace/adamic-tools/env.sh`.
Setup reported Go/clang/Node/submodules ready in 0 seconds each, cache warm
119 seconds, total 119 seconds. `nproc` is 5; cgroup quota is 4 cores.

```sh
ADAMIC_WAVE21_JSX_ARTIFACTS=/workspace/wave21-jsx-final \
ADAMIC_WAVE21_JSX_CONTRACT_ARTIFACTS=/workspace/wave21-jsx-contract-final \
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json \
ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest \
ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest \
go test ./stage1/cohere/typeaware/wave21_jsx -count=1 -v -timeout=30m \
  > /workspace/wave21-jsx-c019.log 2>&1

ADAMIC_WAVE21_JSX_ARTIFACTS=/workspace/wave21-jsx-final \
go test ./stage1/cohere/typeaware/wave21_jsx \
  -run '^TestJSXSourcePrerequisites$' -count=1 -v -timeout=10m \
  > /workspace/wave21-jsx-prerequisites.log 2>&1

python3 stage1/cohere/typeaware/wave21_jsx/compare_node.py \
  /workspace/adamic /workspace/wave21-jsx-final /workspace/wave21-jsx-node-c019 \
  > /workspace/wave21-jsx-node-c019.log 2>&1
```

The first complete pre-rebase run passed in 949.114 seconds. Its 66 control
batches took 0.122356 seconds for prepared-input native execution, 1.435371
seconds for full Go execution, and 1.524085 seconds for raw Go fixture
preparation. These are different input pipelines, not an end-to-end native
speed comparison; native source performance remains unavailable. Compilation
and sanitizer timing are excluded from those three execution sums.

The branch was rebased onto current main `c01907a70` after that run. All 27
existing patches are unchanged; `validation/range-diff.log` records the mapping.
Fresh rebase gates and final source/artifact hashes are recorded in validation.
The source parser blockers remain: fragments and undefined tags expect
`GreaterThanToken` and encounter `SlashToken`; the context provider encounters
`Identifier` instead. Shared JSX and node-role integration must land before
these reservations can be called completed native source ports. The prior three
high-level React analysis reservations remain parked, not released.

The complete repository test gate was not run. This unit reruns its owned
oracle suites, checker package, filtered external Node oracle, production Go
JSX tests and vet. Suppression/edit application, production registration, live
source-to-node/checker adaptation and end-to-end throughput are not covered.

Current-main results: seven previous `TestWave21*` suites PASS **859.448 s**;
new JSX suite PASS **1004.383 s**; source prerequisites PASS **11.204 s**;
checker PASS **0.157 s**; filtered external Node oracle PASS **1.756 s**;
`go vet ./...` exits 0 with empty output. External source/emitted Node repeats
all 22 default batches successfully. `validation/landing.json` records final
execution sums; `validation/mutants.md` records every fresh mutant catch,
including inherited rules and prerequisite guards. This is a targeted gate,
not the complete repository gate.

## Named listener correction and current-main landing

Rebased onto main `b8fb957aa` without conflicts, accepting its inherited static
field fix. The three rule.json descriptors now use the AST names validated by
`origin/lint-rules/harness` (`41eb6eab2`): `JsxFragment`, `JsxElement`,
`JsxSelfClosingElement` and `JsxOpeningElement`. Private prepared-syntax tags
remain internal to the test provider and cores. These descriptors are listener
declarations, not completed shared-registry adapters. No shared file changed.

`TestNamedListenerMetadata` compares the descriptor names with the pinned Go AST
constants after exactly the registry's `Kind` prefix normalization. A temporary
valid-but-wrong `CallExpression` listener is rejected by that independent check;
the restored descriptors pass. The existing native listener and real rule
mutants remain separate checks.

Setup: Go ready 0 s, clang/Node/submodules ready 1 s, cache warm 91 s, total 92 s;
`nproc` 5 and cgroup quota 4 cores. The initial checker command used a nonexistent
`internal/checker` package; it was corrected to `bridge/tsgo/checker`, PASS 0.136 s.
Filtered Node oracle PASS 4.846 s; main's inherited-static-field fixture PASS
0.429 s. Vet exits 0. All seven prior wave-21 suites PASS 803.836 s. External Node
source and emitted-JavaScript comparisons pass all 22 default JSX batches.

Fresh evidence is under `validation/landing-b8fb/`. Source parsing still refuses
all three JSX positives under normal and sanitizer builds; skipping the parse is
caught by the prerequisite oracle. Raw node-role and live checker adaptation
remain integration gaps even with the named-listener shared driver. The old
prepared-input timing observations above do not establish source throughput.

Fresh JSX suite PASS **1091.267 s**, including 66 batches (216 controls in three
profiles), both frozen corpora, three normally exiting native rule mutants,
ASan/UBSan/LSan, named metadata, Unicode and released-handle contracts. Execution
sums: prepared-input native **0.137763 s**, full Go **1.702417 s**, raw Go fixture
preparation **1.890935 s**. These are distinct pipelines. The full repository
test gate and native JSX source throughput remain uncovered.
