Built: reporting-only portions of three newly claimed React rules; the earlier six rule ports remain complete and pushed.
Commits: completed correctness ports `99c9e8d3` / evidence `8673be2b`; React claim `f3b2e6f9`; partial reporters `758b9c60`.
Commands and outputs: Go controls PASS with 2 memo, 3 purity and 1 refs findings; 11 reporting records match 5948 bytes across Go, native, ASan/UBSan/leaks, source Node and emitted JavaScript.
Mutants: three renderer end+1 mutants exit 0 and fail only byte comparison; three removed-refusal mutants exit 0 and fail the expected NotYet panic check.
Not covered: React analysis verdicts, corpus parity, full-rule mutants, checker-handle checks or timings for these three; missing native React HIR and JSX parsing block them.

## Claimed rules and honest status

- `react-hooks/preserve-manual-memoization`
- `react-hooks/purity`
- `react-hooks/refs`

These were the first three unported/unclaimed names after the previous three
were pushed. Selection fetched all 358 origin references, used the 197-rule
combined volume ranking, excluded the 25 ranked baseline ports and 139 names
in remote claims, and preserved the search proof in
[wave-04-selection-2.json.gz](../claims/wave-04-selection-2.json.gz).
The claim was pushed in `f3b2e6f9` before writing any implementation.

These three are **unfinished and remain claimed**. Their separate `.a` files
implement message selection, formatting, span transport and zero fix/suggestion
fields. They do not implement a lint analysis. Calling `analyze()` gives an
explicit `NotYet` panic rather than silently returning an empty finding list.
The reporter is deliberately separate from the verdict so it can be integrated
when the prerequisite native substrate exists.

## Exact blockers

The purity control is the production rule's original vendored fixture. Go reports
three findings, one for each of `Date.now`, `performance.now` and `Math.random`.
The unchanged shared native parser exits 70 before a rule can run:

```text
adamic: panic: parser slice expected GreaterThanToken, got Identifier at 166 in /workspace/typeaware-wave-04-react/final-2/purity.tsx
```

The memoization and ref controls parse natively, so JSX is not the only blocker.
All three production analyses consume
`cohere/internal/lint/ecmascript/high_level_intermediate_representation`:

- Memoization calls `CloneFunction(ForFunction(...))` and
  `AnalyzePreservedManualMemoization`. The latter outlines and inlines functions,
  infers reactivity and mutable ranges, builds SSA/reactive scopes, aligns and
  merges scopes, collects dependencies, rebuilds the reactive tree, prunes scopes
  and validates whether source memoization survived. A syntax-only approximation
  cannot reproduce that comparison.
- Purity walks HIR instructions and phis, propagates builtin aliases, carries
  captured values across closures and re-raises nested calls at their true spans.
- Refs consumes `ForFunctionWithoutManualMemoization`, SSA/phi values, captures,
  nominal checker types and a six-element lattice with an exact ten-round bound.

There is no native React HIR implementation under stage 1 on this branch, and
no checker bridge question supplies it. Scoped source searches return exit 1
with zero stdout and stderr; their commands and outputs are preserved in
[evidence](evidence). The existing syntax CFG from the completed correctness
ports provides control-flow edges, not these typed SSA values, capture spaces,
reactive scopes or memoization passes.

Completing these analyses requires a native port of that shared substrate, plus
JSX support for the corpus. Running the Go validators in a new bridge question
would pass Go lint verdicts across the bridge and would not be a native rule
port. No such shortcut was added.

Ahra's correction says "Keep your changes inside your own rule directories" and
"If anything else blocks you, say exactly what it is and stop, rather than
editing shared files." The shared parser, lowering, harness and registration
files were left untouched. Work stops here with the missing dependencies
recorded; this is not an automatic approval rejection or a request for permission.
No further rules were claimed.

## What the partial checks establish

An isolated Go test overlay imports the production React rule package and runs
its unmodified rules through the typed test API. It exports positive controls:
`reassignedContextCapture`, `purityImpureFunctionsInRender` and the ref-read
fixture. They report 2, 3 and 1 findings respectively. The overlay also exports
all five `refsMessageFor` arms, including the non-convergence condition.

The 11 records are replayed through the native reporters, producing identical
complete canonical diagnostic lines: 5948 bytes. Normal and sanitized native,
source on Node with types stripped by the existing oracle loader, and emitted
JavaScript on Node all match the independent Go text. This verifies reporting
only; replaying Go's test records does not establish a native analysis verdict.

Each renderer's end+1 mutant compiles, exits 0, produces 11 records and has empty
stderr; only byte comparison kills it. Each analysis-refusal mutant removes its
panic, compiles and exits 0 with empty stderr; the expected exit-70 assertion
kills it. These are partial reporting/refusal mutants, **not full-rule mutants**.
The parser probe's memo and refs inputs exit 0; its purity input exits 70 as above.
The three analysis probes exit 70 with their distinct `NotYet` reasons.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_partial.py /workspace/typeaware-wave-04-react/final-2 > /workspace/typeaware-wave-04-react/final-2-validation.log 2>&1
go vet ./... > /workspace/typeaware-wave-04-react/vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_04_react/testdata/export_test.go > /workspace/typeaware-wave-04-react/gofmt.log 2>&1
```

Validation ends `PASS partial reporting only`; vet and formatting logs are empty.
All test output went to files. The full repository gate and full-rule corpus
comparisons were not run for these unfinished rules. Native/Go timing comparisons
would be misleading while native analysis cannot run, so none are claimed.

The completed earlier correctness trio's normal/sanitized corpus results, three
comparison-only rule mutants, released-handle registry mutant and timings remain
in [its complete report](../wave_04_next/REPORT.md): compiler medians native
1.815s versus Go 0.335s; repository native 0.266s versus Go 0.118s.

## Dependency refresh on 2026-10-07

Fetched every origin head again and froze all 399 reference SHAs in
[evidence/refresh/origin-inventory.json](evidence/refresh/origin-inventory.json).
The inventory searches native source filenames for HIR, React, memoization and
JSX candidates; it is not an exhaustive semantic search of all source bodies.
Main is `e011f8f6`, the bridge branch `5afbdb83`, and the harness branch
`f4d98cab`. None supplies a native React HIR implementation in this inventory.

There is now a reusable isolated JSX parser on `origin/codex/typeaware-wave-16`,
SHA `58d204d3`, in four `wave16_jsx*.a` dependency files. Its follow-up report
records positive JSX comparison controls. The narrow JSX adapter helpers on the
helper branches describe raw JSX nodes; they do not implement React HIR. Thus
JSX is a potential integration dependency, rather than evidence that no JSX
implementation exists anywhere. This branch still uses the unchanged shared
parser, which rejects the original purity control. No copied parser or shared
parser edits were added while the separate HIR prerequisite remains missing.

The React worker on wave 08 independently records missing HIR lowering, SSA and
instruction transfer. Its partial lattice/derivation kernels do not supply the
three pipelines needed here. Worker reports are frozen beside the inventory.

Rebuilt and reran the owned partial validation:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_partial.py /workspace/typeaware-wave-04-react/refresh > /workspace/typeaware-wave-04-react/refresh-validation.log 2>&1
```

PASS: the 11 reporting records still match Go at 5948 bytes in native, sanitized
native, source Node and emitted JavaScript. All three end+1 renderer mutants
exit zero with empty stderr and are caught by byte comparison; all three
removed-refusal mutants are caught by the expected-panic assertion. Memo and
refs parse controls pass; the purity JSX parser probe and all three explicit
analysis refusals still exit 70. These remain partial reporting/refusal checks,
not completed rule ports, full-rule mutants, corpus analysis or rule timings.

No further rules were claimed. The exact remaining prerequisite is native
React HIR lowering and SSA/capture propagation, plus memoization's reactive-scope
passes. The user instruction to stop on prerequisites outside the owned rule
directories still applies. The prior completed six rule ports remain pushed.

## Refs convergence kernel continuation

`refs_value.a` now ports `refsTypeEqual`, `refsFunctionEqual`, `refsIsHookName`
and `refsIsRefLikeName` from production Go. Values and functions use arena
indices rather than owning recursive graphs. This is a partial analysis kernel,
not a source lint entry point: `Refs.analyze()` still refuses the missing HIR
pipeline. No new claims, shared harness edits or bridge verdict queries were added.

The six lattice kinds retain Go's exact convergence equality: none/nullable
ignore metadata; guards compare ID and presence; refs ignore identity; ref values
compare access span and presence; structures recursively compare held values
and function effect/return types. The ref-name predicate preserves Go's actual
implementation, including accepting `a-Ref`; it does not substitute the stricter
regex described in the upstream comment. Hook names require an ASCII uppercase
letter after `use`, excluding `use9`.

An owned Go test overlay calls the unmodified private production functions.
Its 23 value descriptors include distinct ref IDs, zero/present guard identity,
span presence, nested structures, nil links, function effects and function return
types. All pairs, plus 16 naming controls including Unicode and punctuation,
produce 592 records / 10,151 bytes. Native, ASan/UBSan/leaks, source Node and
emitted JavaScript match that stream byte for byte with empty stderr. Four
sanitized kernel mutants compile and exit zero with empty stderr and are caught
only by the Go byte comparison:

- Comparing ref identity during convergence.
- Dropping the guard ID presence bit.
- Dropping the access-span presence bit.
- Accepting a bare `Ref` binding name.

The first attempt omitted the guard ID-zero/presence witness, so its mutant
survived. That failure is retained in evidence. Adding the witness killed the
mutant in the final run; the initial run is not counted as a pass.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_refs_kernel.py /workspace/typeaware-wave-04-react/refs-kernel-final > /workspace/typeaware-wave-04-react/refs-kernel-final.log 2>&1
```

The final command prints `PASS partial refs kernel: 592 records, 10151 bytes`.
Inputs, outputs, build logs, source hashes, initial failure and fetched reference
SHAs are in [evidence/refs-kernel](evidence/refs-kernel). The bridge branch's new
`eb6df00e` tip changes build-profile evidence, not the missing React HIR.

Uncovered: refs joins and transfer/sweep, source lowering, SSA, checker ref-type
questions, manual-memoization reactive scopes and purity alias/capture analysis.
No full React rule mutant, corpus verdict comparison, released-checker-handle
check or native/Go rule timing is claimed for this kernel. The earlier six full
ports and their measurements remain unchanged. The full repository gate was
not rerun; this continuation checks the owned kernel through its Go overlay
and both Adamic backends.

## Refs joins and destructuring continuation

Landing cap verified before this work: main remained `e8ba3d5d`, wave 04 was
rebased and fully re-greened there, and its pushed SHA was `e945c91d`. A remote
read after validation confirms that same main tip. No additional rules were
claimed and no shared files were edited.

The owned `refs_value.a` now also ports production Go `refsJoin`,
`refsJoinRefCarrying`, `refsJoinMany` and `refsDestructure`. The arena allocates
fresh ref IDs in the same order as Go; none is identity, conflicting guards
and guard/nullable joins collapse, ref values dominate carrying joins and erase
provenance unless both operands identify the same ref, structures recursively
merge values/functions, function effects combine with OR, and origin spans prefer
the left operand when present. Folding starts at none; destructuring follows
structure values until a non-structure or an empty structure.

The expanded Go overlay calls those private production functions directly.
Twenty-four value descriptors plus nil cover all ordered pairs, nested
structures, distinct and shared ref identities, zero/present guard identity,
access/origin spans, function effects, function returns and function origin
preference. Six folds and sixteen name controls complete 1,297 records / 55,984
bytes. Native, ASan/UBSan/leaks, source Node and emitted JavaScript match exactly
with empty stderr. No Go lint verdict crosses a bridge.

Nine sanitized native kernel mutants compile, exit zero with empty stderr and
are caught only by the independent Go byte stream. The four prior equality/name
mutants still fail. Five new mutants preserve a guard/nullable join incorrectly,
shift fresh ref IDs, use AND for function effects, stop destructuring too early,
and select the right origin for a same-ref join. The last mutant depends on the
new shared-identity/different-origin witness rather than an incidental failure.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_refs_kernel.py /workspace/typeaware-wave-04-react/refs-joins-release --adamic /workspace/typeaware-wave-04-landing-current/adamic > /workspace/typeaware-wave-04-react/refs-joins-release.log 2>&1
```

The command prints `PASS partial refs kernel: 1297 records, 55984 bytes`. Inputs,
source hashes, Go results, exact backend streams, mutant build/run logs and
validation are in [evidence/refs-joins](evidence/refs-joins). The six earlier
complete rules and their unchanged main-base validation remain in the
[landing report](../wave_04_next/LANDING_REPORT.md); these new kernel files are
not imported by those rule runners. Their latest native/Go medians are compiler
1.905/0.318s and repository 0.296/0.130s for the correctness continuation.

This is still a partial refs analysis kernel. The source entry point refuses
missing native React HIR/SSA, instruction transfer, captures, nominal checker
ref-type questions and the bounded sweep. Purity alias/capture propagation and
manual-memoization reactive scopes also remain missing. Full React findings,
fixes/suggestions, full-rule mutants, checker-handle tests and full-rule timings
remain unverified. The full repository gate was not rerun; the owned kernel
oracle and both backends cover this change. No further claims were made.

## Refs environment continuation

Landing checked first: wave 04 was already rebased, green and pushed on current
main `e8ba3d5d`, with remote SHA `a8c569a2`. No branch was pushed except
`codex/typeaware-wave-04`. No new rules were claimed.

`refs_environment.a` ports the production `refsEnvironment` helpers: ref identity
allocation, one-step temporary resolution, definition/access-node transport,
source/cached/property names, declaration lookup and monotone widening. Native
values, nodes, identifiers and declarations use arena indices; maps do not own
recursive graphs. Declaration ID zero and node index zero are valid values.
`set` preserves Go's argument order, stores initial none without marking a change,
uses convergence equality rather than ref identity, and caches non-none results
against both the original and resolved declarations. `get` prefers direct
resolved data, then the original declaration, then the resolved declaration.

The owned Go overlay calls the unmodified production helpers on four operation
traces. Forty-one steps produce 451 records / 9,943 bytes. Controls exercise
initial none, reset/change state, repeated distinct refs, widening with none,
shared declarations, aliases onto unnamed temporaries, access-node precedence,
source-name precedence, empty-name/property preservation, missing bindings and
identity allocation. Go, native, ASan/UBSan/leaks, source Node and emitted
JavaScript agree byte for byte with empty native/Node stderr.

Six sanitized environment mutants compile and exit zero with empty stderr and
the same record count; only the Go byte comparison catches them: ignore alias
resolution, query the wrong declaration, suppress change state, mark initial
none as changed, discard the incoming widening value, and lose access-node
transport. These are partial environment mutants, not a full refs rule mutant.

The first normal build refused a string `||` expression. The owned implementation
now uses Go's explicit name-selection branch; no compiler files changed. The
first declaration mutant used `false`, which prevented TypeScript narrowing and
failed compilation. It was replaced with a type-correct wrong-key lookup and
rerun; the compiler failure is not counted as a killed oracle mutant. Both initial
failures are retained in evidence.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_refs_environment.py /workspace/typeaware-wave-04-react/refs-environment-release > /workspace/typeaware-wave-04-react/refs-environment-release.log 2>&1
```

The command prints `PASS partial refs environment: 451 records, 9943 bytes`.
Inputs, exact output streams, build logs, source hashes, mutants and initial
failures are in [evidence/refs-environment](evidence/refs-environment). This uses
the fresh stage 0 built for main `e8ba3d5d`; the six finished rule runners do not
import this new environment. No shared harness, generator or bridge dispatch
was edited.

A fresh inventory of 466 origin refs finds wave 21's `wave21_react/hir.a`, whose
header explicitly says graph production from source is a separate adapter. Its
input model is preserved in evidence. This updates the earlier dependency claim:
a native HIR input model exists, but it does not supply source lowering, SSA or
the three complete pipelines needed here. The inventory is a filename search,
not exhaustive semantic inspection.

Uncovered: connecting source/HIR to the environment, checker ref-type facts,
instruction transfer and the ten-round sweep, purity aliases/captures, and
manual-memoization reactive scopes. These three claimed React rules remain
incomplete, with explicit source-analysis refusals. No full React findings,
fixes/suggestions, full-rule mutants, checker-handle tests or full-rule timings
are claimed. The full gate was not rerun for this owned kernel. Earlier complete
ports, released-handle checks and native/Go timing remain in the landing report.
No further claims were made.

## Refs finding predicates continuation

Landing verified first: main was still `e8ba3d5d`, already an ancestor of the
green pushed wave 04 branch at `5df6d34b`. No rebase was needed and no additional
rules were claimed. This work stays in the owned React directory.

`refs_checks.a` ports `refsCheckDirectValueAccess`, `refsCheckValueAccess`,
`refsCheckPassedToFunction` and `refsCheckUpdate`. These operate on the already
classified environment. Direct access reports ref values, including values
inside structures. The broader value check also reports ref-reading functions,
using the value-access finding kind. Passing checks both refs and ref values,
plus ref-reading functions; it deliberately omits an access-node override.
Updates check refs and ref values and preserve the supplied update node. All
helpers append to existing findings rather than replace or clear them.

The Go overlay calls the unmodified private production helpers on eleven
classified values plus absent data, four checks and both seeded/unseeded finding
lists. Controls include none, nullable, guards, bare refs, ref values, nested
structures, functions with and without read effects, a function returning a ref
value without a read effect, and a structure carrying none plus a read function.
Cached access-node and explicit update-node identities are distinct, so a span
source swap changes the comparison.

All 96 cases produce the same 180 records / 3,136 bytes on Go, native,
ASan/UBSan/leaks, source Node and emitted JavaScript, with empty backend stderr.
This stream compares complete helper finding metadata (kind, value, presence
and node override), not source lint verdicts or final source spans. The earlier
reporting oracle independently covers diagnostic text and zero repairs.

Four sanitized predicate mutants compile and exit zero with empty stderr and
are caught only by the independent Go bytes: direct reads classify refs instead
of ref values, the function-read effect is inverted, passing emits the wrong
finding kind, and an update loses its explicit node. These remain kernel mutants,
not the required full refs rule mutant.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_refs_checks.py /workspace/typeaware-wave-04-react/refs-checks > /workspace/typeaware-wave-04-react/refs-checks.log 2>&1
```

The command prints `PASS partial refs predicates: 96 cases, 180 records, 3136
bytes`. Inputs, source hashes, exact backend streams and mutant build/run logs
are in [evidence/refs-checks](evidence/refs-checks). It uses the fresh stage 0
built for main `e8ba3d5d`. No shared harness, generator or bridge was edited.
The six earlier complete rule runners do not import these new predicates.

Uncovered: source/HIR lowering, SSA and capture propagation, type-based ref
classification, instruction transfer, guarded initialization, bounded sweep and
final AST-to-diagnostic placement. Purity and manual-memoization also still need
their HIR/reactive pipelines. All three source analyses remain explicitly refused.
No full React corpus parity, fixes/suggestions, full-rule mutants, checker-handle
checks, timings or full repository gate are claimed for this kernel. Earlier
complete ports retain their landing evidence and native/Go measurements. No
further rules were claimed.

## Numeric listener declarations

All nine claimed rule modules now export `listenerKinds: readonly number[]`.
These are pinned TypeScript-go numeric SyntaxKinds, independently extracted
from the actual production Go `rule.Listeners` maps by
[testdata/listeners.go](testdata/listeners.go), rather than inferred from enum
line numbers. The three React rules listen to SourceFile (307), matching Go.
The strict-void rule declares all twelve listeners, including CallExpression
(214) and NewExpression (215).

`validate_listeners.py` imports the actual nine modules and compares complete
listener output with Go in native, ASan/UBSan/leak and source Node runs. Each
module also has a compiling, successful-exit mutant incrementing its first
numeric listener; only the byte comparison catches these nine metadata mutants.
These supplement, and do not replace, prior rule and kernel mutants.

Emitted JavaScript validation is blocked: `adamic js` refuses the imported
`rules.ts` checker call as an unlinked typescript-go library call. Its CLI has
no `--tsgo` option; attempting that option prints usage and exits 2. Both
failures are retained in listener evidence. No shared emitter was changed.

This adds declarations, not a dispatch speedup. The current `ParseNode.kind`
is a string and the six earlier complete rule runners still use legacy
whole-file scans and string-kind comparisons. Converting those runners to act
only on a handed numeric-kind node requires the shared numeric parser/driver
API, outside this unit's allowed territory. The three React analyses still
explicitly refuse source execution because their native source-to-HIR/SSA and
reactive pipelines are incomplete. No new rules were claimed.

Main remains `e8ba3d5d`; prior landing oracle results remain in force for
unchanged rule behavior. Latest measured complete continuation corpus timing:
compiler native 1.904769 s / Go 0.318496 s; repository native 0.296343 s / Go
0.130137 s. Listener declarations are currently ignored by the legacy driver.
No full repository gate or complete React corpus agreement is claimed.
