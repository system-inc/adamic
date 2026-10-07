Built: numeric symbol-description listener; atomic transfer/meet kernel; require-await syntax kernel; three manifests.
Commits: React parking bd93c67d3; fourth-batch claim fcfd9568c; implementation/evidence follows.
Commands and outputs: symbol listener 17 controls/6 findings and both corpora byte-identical; 16 atomic and 8 await kernel controls agree with production Go.
Mutants: symbol argument-count inversion, atomic refresh deletion and await-using mask mutation caught by byte comparison; new handle-retention mutant caught by required panic.
Not covered: native shared-driver integration, atomic full rule and require-await full rule; these three claims remain incomplete.

# Fourth batch partial implementation

The prior six ports are green on origin/main f8013f0b and pushed. The three React
claims are explicitly parked: globals awaits JSX integration; immutability and
no-deriving-state-in-effects await React HIR/SSA/capture passes plus JSX. Ahra
identified #dnv6f2c and area/stage1-lint as the dependency owners. Those branches
were not modified or pushed by this unit.

After fetching 529 origin refs, the volume ranking had 197 checker-dependent
entries, 154 named in origin claims and 25 already ported on main/bridge. React
and structure/react-hook-no-any-type entries were skipped while React is parked.
The next three non-React candidates were require-atomic-updates, require-await
and symbol-description, each with zero compiler/repository volume. An immediate
pre-claim refresh found none reserved; the claim was pushed before code.

## Symbol listener

symbol_description/rule.a takes a supplied HandedNode with numeric SyntaxKind
and UTF-8 ranges. It neither reads a string kind nor retrieves the current node
from a parser. rule.json listens to CallExpression (214). The raw bridge question
call-symbol-shape, in its own Go and .a files, supplies argument count, the
parenthesis-unwrapped callee's numeric kind/text, symbol presence and each
resolved declaration's declaration-file flag. The Adamic handler chooses the
finding and exact production message. No fix or suggestion is emitted.

The isolated test driver receives numeric syntax captures from the independent
Go oracle. It is deliberately a driver replacement for these tests: it is not
proof that the shared native parser supplies the right nodes or spans. Production
Go rules remain unchanged and import no bridge code. The listener matches all
17 controls (6 findings, 2812 bytes), 77 compiler roots (0 findings, 5318 bytes)
and 287 frozen repository roots (0 findings, 18485 bytes). All three comparisons
also match under ASan/UBSan and leak checks. Tests include shadowing, parentheses,
optional calls, new/member calls, explicit undefined, spread and Unicode.

The argument-count mutant compiles, exits 0 with empty stderr and disagrees only
at byte comparison. Released call-symbol-shape inspection panics with exit 70 and
`adamic: panic: invalid or released checker handle`. A registry-retention overlay
mutant exits 0 with empty stderr, proving that check can fail. Checker package
tests PASS 0.123s. No shared registration file or harness was changed.

Three interleaved process timings per corpus, with captured syntax generation
excluded: compiler Go median 0.292923093s, native 0.821113525s (2.803x); repository
Go 0.126508479s, native 0.168059858s (1.328x). Both processes load a Go checker;
Go traverses its AST, while native consumes the capture. These observations are
not a full native-parser benchmark or a claimed speed improvement.

## Atomic kernel

require_atomic_updates/state.a ports the production fresh/outdated state,
clone, union meet and per-event transfer. Both sets travel across joins; a read
refreshes its symbol, a suspension stales fresh symbols and a write tests the
outdated set. Numeric event kinds and checker symbol IDs carry identity.

Sixteen transfer/join scenarios agree byte for byte with the unchanged production
lattice through a test-only Go overlay. Original and refresh-deletion mutant
compile and exit cleanly under sanitizers; Go bytes catch the mutant. Full rule
replay still needs numeric syntax/event delivery, deferred suspension/store
placement, binding resolution, escape filtering, fixed-point solving and final
finding rendering. The kernel is not presented as a complete rule.

## Await syntax kernel

require_await/syntax.a ports the empty-body and await search predicates using
numeric node kinds, declaration flags and a supplied indexed node graph. It
recognizes await, for-await and the composite await-using mask, and skips nested
functions/classes. Eight source controls parsed independently by typescript-go
agree with the unchanged production predicates. Two additional native-only
controls exercise an empty block and concise await. The mask mutant compiles
and exits cleanly under sanitizers, then fails the eight-case Go comparison.

Full require-await replay still needs shared numeric function/parent delivery,
checker type/signature facts for contextual and declared-generic promise demand,
heritage demand, exact head/name rendering, async-token repair and suggestion
serialization. No full-rule oracle, fixes, suggestions or timing is claimed for
this rule. The manifest lists the seven production function kinds.

## Current blocker and scope

The shared parser on this branch still has ParseNode.kind:string. There is no
shared numeric node-delivery API here. New listeners must consume a supplied
node, so the new code declares numeric kinds and avoids adding old string-based
per-rule dispatch. Numeric delivery and the shared Diagnostic integration are
pending with their owners. Full native integration cannot be validated against
the shared driver until that API arrives. The additional atomic/await work
listed above is unfinished implementation, not a claim that their entire
semantics are blocked by the harness alone.

No new batch may be claimed until these three are completed or explicitly
parked. Evidence, timings, independent captures, exact streams and mutants are
in validation. verify_symbol.py documents the runnable listener checks; the
kernel .a entries and Go overlay helpers document the smaller comparisons.
All new Adamic source files use .a. The preceding six-port oracle gate was not
repeated for this additive partial batch; only the new checker package and the
new listener/kernel checks were run.

## Landing refresh onto c01907a7036a

Rebased onto origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 without
conflicts. Tested rebased code at 52df32d42123d8471019ccd5726aafd2667f798d.
No shared leak-helper changes were reverted.

The original TestWave01AgreementAndMutants gate passed in 162.650 seconds.
The continuation verify.py gate passed all three completed continuation rules,
including compiler and repository comparison, sanitizers, logic mutants,
released handles and retention mutants. The symbol verify_symbol.py gate
passed controls, compiler and repository comparison through captured syntax,
sanitizers, its argument mutant and released-handle check. Atomic and await
kernels again matched unchanged production Go; their refresh-deletion and
await-using-mask mutants exited cleanly under sanitizers and failed byte
comparison. Exact logs and records are in validation/landing-c019.

This refresh does not complete the three latest claims: symbol native frontend
integration and full atomic/await rules remain unfinished as described above.
No additional rules were claimed. The production Go sources of these three
rules contain no regex usage, so this batch requires no regex translation.

## Atomic fixed-point solver continuation

On unchanged origin/main c01907a7036a, added require_atomic_updates/solver.a.
The driver supplies indexed successor edges and numeric event/symbol pairs.
The solver discovers entry reachability, carries both sets across predecessor
joins, iterates to a fixed point and refuses nonconvergence after 10000 rounds.
It does not interpret syntax kinds or refetch parser nodes.

The independent Go overlay constructs test graphs, then invokes the unchanged
production control_flow_graph.Solve and outdatedReadLattice. Sixteen graphs
exercise branch joins, suspension, refresh, loop back edges and an unreachable
predecessor. All 96 block-boundary observations match byte for byte. Native
and the one-sweep mutant both exit zero with empty sanitizer stderr; the mutant
fails only the Go byte comparison. Exact streams are in validation/atomic-solver.

Command after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/typeaware/wave_01_fourth/testdata/verify_atomic_solver.py /workspace/wave-01-solver-validation /workspace/wave-01-fourth-adamic > /tmp/wave-01-solver.log 2>&1
```

Output: PASS 16 graphs, 96 block boundaries; sanitizers clean; one-sweep mutant
caught by Go bytes. gofmt and git diff --check pass. Initial builds exposed two
stage-0 limitations: prefix increment in a condition and early constructor return.
Both were avoided in the final source; no compiler files were edited.

Setup succeeded: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s,
build cache warm 127s, done 127s. nproc is 5; cgroup quota is 4 CPUs.

This adds the solver only. Full atomic event collection, deferred positioning,
escape filtering and rendering remain unfinished; full await implementation and
symbol native integration remain unfinished too. The shared supplied numeric-node
API is still absent from current main. No new claim was taken. The six completed
ports were unchanged and their previously recorded oracle gates were not repeated.
No full-rule timing, released-handle coverage or corpus agreement is claimed for
this pure solver addition; it creates no checker handles.

## Atomic deferred-event placement continuation

Current origin/main remains c01907a7036a; branch base is unchanged. Added
require_atomic_updates/deferrals.a, porting production placeDeferrals over
supplied raw byte ranges and original event-order floors. Kept events preserve
their order; deferred events sharing an insertion slot preserve their order.
No regex or diagnostic offset conversion is involved.

An independent test-only Go overlay invokes unchanged placeDeferrals on 32
fixtures: four floors, contained and noncontained valid ranges, missing targets
and shared insertion slots. All output bytes match native under ASan/UBSan/LSan.
The mutant that drops the original floor compiles, exits zero with empty stderr,
and fails only the production-Go byte comparison. Exact streams are recorded
in validation/atomic-deferrals.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_01_fourth/testdata/verify_atomic_deferrals.py /workspace/wave-01-deferrals-validation /workspace/wave-01-fourth-adamic > /tmp/wave-01-deferrals.log 2>&1
```

Output: PASS 32 placements; sanitizers clean; lost-floor mutant caught by Go bytes.
gofmt and git diff --check pass. Existing completed-rule gates were not repeated
for this isolated helper, which is not wired into their suites. No checker handle
is created by this helper; no new handle check or full-rule timing is claimed.

Remaining atomic work includes syntax/CFG event collection, symbol resolution,
escape filtering, construction of the two deferral phases and full findings.
Await and symbol integration remain incomplete as previously recorded. The
shared numeric supplied-node interface remains absent on main. No new claims
were taken, and no shared files were edited.

## Named-kind contract correction

Ahra superseded the numeric rule.json contract: registry kinds are typescript-go
ast.Kind names. Corrected all nine owned rule manifests to names and updated
the previous listener check accordingly. Raw numeric AST captures in standalone
tests remain an internal representation, not a required shared registry API.
The earlier assertion that lack of numeric delivery blocks integration is
superseded; named supplied nodes are the requested integration contract.

verify_named_kinds.py compares every manifest with the unchanged production Go
listener registration. All nine pass. For each manifest, replacing its first
kind with a different valid kind is rejected by that same comparison. Output
is preserved in validation/named-kinds/comparison.log.

```sh
python3 stage1/cohere/typeaware/wave_01_fourth/testdata/verify_named_kinds.py > /tmp/wave01-named-kinds.log 2>&1
```

Current main remains c01907a7036a. The executable lint context/harness announced
at ab70f38d4 is not in this checkout's main base. The inspected RuleContext there
adds reportNode/reportRange but exposes no checker program handle for the
type-aware listener. Native symbol/shared-driver integration remains unvalidated.
Atomic and await have unfinished implementation beyond this integration gap.
No shared files were changed and no additional claims were taken. This metadata
change does not alter rule execution, so full corpus/sanitizer/handle gates were
not repeated; their prior observations remain recorded separately.

## Landing refresh onto b8fb957aa839

Fetched origin and rebased without conflicts onto origin/main
b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Tested rebased code at
b2d36db8b4c89973ab024cd9bc3d7497f9534640. No shared changes were reverted.

Original TestWave01AgreementAndMutants passed in 148.649 seconds: compiler
691 findings / 276439 bytes, repository zero findings, all controls, normal
and sanitized comparison, three logic mutants, released handles and retention
mutants. Continuation verify.py passed its three rules, all corpora, sanitized
comparison, three logic mutants, four released questions/retention mutants
and package tests. Symbol verify_symbol.py passed captured-syntax corpora,
sanitation, argument mutant and released question; full native-driver integration
is still not claimed.

Atomic and await kernel comparisons passed with sanitizer-clean refresh and mask
mutants caught by Go bytes. Atomic solver passed 16 graphs / 96 block boundaries
and its sanitizer-clean one-sweep mutant. Deferred placement passed 32 cases
and its sanitizer-clean lost-floor mutant. All nine named-kind manifests passed
production registration comparison and valid-but-wrong-kind mutants.

Commands were the recorded original go test gate with -count=1 and -timeout 30m,
continuation verify.py with compiler corpus /workspace/wave-01-typescript,
verify_symbol.py, verify_atomic_solver.py, verify_atomic_deferrals.py and
verify_named_kinds.py, after sourcing /workspace/adamic-tools/env.sh. Kernel
Go overlays and sanitized native entries were rebuilt and compared. Exact logs
and streams are in validation/landing-b8fb.

This turn handles landing readiness. The three latest claims remain incomplete
as described above. No new claim was taken and no main or area branch was pushed.

## Atomic syntax predicate continuation

On unchanged current main b8fb957aa839, added require_atomic_updates/syntax.a.
The driver supplies indexed parent and structural expression/name fields and
operator text. Adamic decides declaration-name exclusion, plain assignment
targets through member chains and property assignment anchoring. There is no
string-kind relevance dispatch and no parser refetch. Manifest kinds stay named.

The independent Go overlay parses 24 controls, captures raw structural fields
and invokes unchanged isDeclarationName, isPlainAssignmentTarget and
propertyAssignmentHeadedBy. All 74 identifier observations agree byte for byte.
Controls include property names versus objects, computed keys, nested chains,
binding elements, parameters, ordinary reads and all 16 assignment operators.
The member-object guard mutant compiles and exits zero with empty sanitizer
stderr, then fails only the Go predicate comparison. Raw captures, expected
results and exact native streams are in validation/atomic-syntax.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_01_fourth/testdata/verify_atomic_syntax.py /workspace/wave01-atomic-syntax /workspace/wave-01-fourth-adamic > /tmp/wave01-atomic-syntax.log 2>&1
```

Output: PASS 24 parsed controls, 74 identifier predicates; sanitizer-clean
member-name mutant caught by Go bytes. gofmt and git diff --check pass. No
checker handles are created; no handle check or full-rule timing is claimed.
This comparison uses captured Go syntax and does not prove native parser facts
match it. Full atomic event/CFG construction, symbol resolution, escape filtering,
two-phase deferral integration and final finding rendering remain unfinished.
Await and symbol integration remain incomplete too. No new claims were taken;
existing completed-rule corpus gates were unchanged and were not repeated.

## Requested harness integration rebase

Fetched and rebased onto origin/area/stage1-lint at
7481e0324e34a2537aafa9db7eeacda50405611b, which includes the announced
harness and newer main changes. Rebase completed without conflicts. Tested
rebased code at b40208758ef2b24820edb1387613d2628ffffe8f. No shared
changes were reverted. Only codex/typeaware-wave-01 is pushed.

Original TestWave01AgreementAndMutants passed in 146.783 seconds, with 691
compiler findings / 276439 identical bytes and zero repository findings. All
controls, sanitized corpora, three logic mutants, released handles and retained
registry mutants passed. Continuation verify.py passed all three completed
rules, all corpus/sanitizer comparisons, logic mutants, four released questions
and retained-registry mutants, and checker/bridge package tests.

Symbol captured-syntax checks passed all corpora, sanitizers, argument mutant
and released question. Atomic/await kernels passed Go comparison and clean
sanitation with refresh-deletion and await-mask mutants caught by bytes. Solver
passed 16 graphs / 96 boundaries and one-sweep mutant; deferrals passed 32
placements and lost-floor mutant; syntax passed 24 parsed controls / 74
identifiers and member-object mutant. All nine named-kind manifests and their
wrong-kind mutants passed. Exact commands, records and logs are under
validation/landing-area-7481; commands use the same recorded corpus and
verification scripts after sourcing /workspace/adamic-tools/env.sh.

The executable shared lint harness is now present. Its RuleContext still lacks
a checker program handle for these type-aware listeners. Full symbol shared
driver integration remains unvalidated; captured Go syntax is not proof of
native syntax delivery. Full atomic and await rules also have unfinished
implementation, not merely a shared-harness gap. No new claims were taken.
No new throughput measurements were made for this rebase; previous timings
remain observations of their recorded versions.
