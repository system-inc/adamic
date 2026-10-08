# runtime-step37-classes

Current result: readonly class tasks and worker-private mutable class tasks are
admitted by separate proofs. Per-task constructors, field initializers, nested
Scanner state, private arrays and method-returned node mutation agree with Node
and are TSan-clean at 1, 4 and 16 threads. Shared mutable receivers and publication
into captures remain refused. The actual scanner's next blocker is its captured
mutable keyword Map. The readonly report below predates the October 8 ruling;
the worker-private section records the final widened unit.

Base: origin/area/runtime 9bc8201f1c97c4c18a03d263a40ece933ffa752e.
Current origin/main 45487a8 is already an ancestor. This is prework for roadmap
step 37 (#xfpyaj8), not a parallel parser implementation.

## Observed refusal paths

- internal/lower/parallel.go:94-119: shared items must be a readonly array with
  a recursively Shareable element type, followed by a known task-body proof.
- internal/lower/parallel.go:297-299: the body effect visitor refuses this without
  an ownership proof. The old branch refused every this token. Calls use the
  independent argument-expression visitor, now held to the same this rule at
  parallel.go:493-495, so passing this cannot skip the rule.
- internal/lower/parallel.go:428-471: receiver evaluation and supported library
  methods are checked; other method dispatch was refused because its complete
  hierarchy was unproven. The new subset is consulted at 468-470.
- internal/lower/moves.go:15-29: only mutable arrays of object records select the
  move prototype. moves.go:150-176 requires fresh object-literal elements with
  numeric/boolean literal fields. A class instance is not that construction.
- internal/lower/moves.go:182-206: move work is an inline synchronous arrow;
  calls, closures, this and throws are explicitly excluded at 205-207.
- internal/lower/shareable.go:96-100: every reachable data field must be readonly.
  A readonly slot holding Scanner or a mutable array does not make its contents
  Shareable; shareable.go:53-56 refuses mutable arrays.

These are source proofs in internal/lower, before IR emission. This unit does
not use a static method signature as a complete dispatch set. The existing IR
CallTargets reader guard remains in force and this change adds no Call.Function
reads. The narrow AST proof instead excludes every possible source hierarchy
and method replacement. It does not extend the move prototype.

## Parser inventory

Reproduce: `go run ./notes/runtime-step37-classes/inventory` from the repository.
The checked AST inventory covers the seven top-level parser/*.ts modules;
comments, strings, types and testdata are excluded. ClassMethods are property
calls whose resolved declaration is a source MethodDeclaration. InterfaceMethods
in the JSON means other source property-call declarations (including readonly
function properties), not a claim that each is nominal interface dispatch.
LibraryMethods resolve to library declarations. No property call is unresolved.
This is a static occurrence census, not dynamic frequency or a proof that every
body is reachable from a particular task. Full data: inventory.json.

| Module | this tokens | this.method() | this.field.method() | other property calls | source class method calls | other source property calls | library method calls |
|---|---:|---:|---:|---:|---:|---:|---:|
| grammar.ts | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| jsx.ts | 109 | 30 | 50 | 22 | 36 | 50 | 16 |
| lookahead.ts | 49 | 1 | 5 | 24 | 25 | 0 | 5 |
| main.ts | 0 | 0 | 0 | 14 | 2 | 3 | 9 |
| nodes.ts | 4 | 0 | 0 | 5 | 0 | 1 | 4 |
| parser.ts | 1022 | 801 | 13 | 157 | 814 | 0 | 157 |
| statements.ts | 597 | 123 | 425 | 95 | 123 | 425 | 95 |
| Total | 1781 | 955 | 493 | 317 | 1000 | 479 | 286 |

There are also 114 direct calls. The 1,765 property calls include 1,000 resolved
class-method calls, 479 other source property calls and 286 library calls.

Concrete parser shapes and race implications (the race descriptions below are
inferences; only the fixture runs and explicit mutation are observations):

| Shape and evidence | Existing obstacle | What TSan would see if unchecked |
|---|---|---|
| parser.file()/docTypes(), main.ts:10-12, on a freshly constructed Parser | task constructor call is unknown; Parser is mutable | If a distinct reachable graph is exclusive to each worker, no receiver race is expected. Sharing scanner/nodes across workers invalidates that inference. |
| this.kind(), parser.ts:27-28, reads this.scanner.kind | this token and mutable Scanner reachable from readonly field | Concurrent scan writes and kind reads conflict on scanner state. Read-only execution alone does not conflict. |
| this.next() -> this.scanner.scan(), parser.ts:30-31; Scanner.advance scanner.ts:73-74 | method/this proof and mutable graph | Shared scanner position increments conflict with reads/writes in other workers. |
| this.make(), parser.ts:36-38, pushes this.nodes | readonly nodes slot holds mutable array | Concurrent size/storage changes conflict, with possible backing-store lifetime errors beyond races. |
| this.node(id).end assignment, parser.ts:43; ParseNode nodes.ts:7-18 | writes a reachable object | Shared node fields race; readonly kind/pos do not protect end/text/raw/children. |
| this.parser.kind()/next(), statements.ts:95-106 and JSX context | interface/function-property target set plus mutable graph | Dispatch itself need not race. An effectful implementation on the shared parser graph would race; all possible targets must be proven. |
| this.path, parser.ts:18-20; primitive fields of ParseNode nodes.ts:5-6 | blanket this rule before this unit | Immutable field loads do not conflict, but admitting a method still requires the whole receiver graph and its entire body to be proven. |

## Implemented proof subset

A non-static source method must have exactly one declaration and belong to the
receiver's concrete class symbol. All modules must be visible. Any class
heritage (including a class expression), any write to a property with the
method's name (including structural aliases), or any element write anywhere in
the source program prevents this admission. These exclusions are deliberately
conservative implementation limits, not new language promises. Interfaces,
base views, accessors, extracted methods and unknown dispatch remain refused.

The existing recursively Shareable proof checks the whole receiver graph, not
just fields read by that method. Existing receiver/capture evaluation supplies
the runtime's shared roots; no new unmarked receiver is introduced. Only in
that admitted method may this be the direct receiver of a readonly, non-static
data field declared on the same class. A bare this, alias, argument, nested
closure capture or this.method call remains refused. The ordinary task effect
proof still checks all method body statements, arguments, defaults, captures
and transitive calls. Writes through an alias or parameter remain refused.

This is enough for an immutable class containing runtime strings and readonly
arrays, or immutable scalar fields, called on a captured receiver or a task item.
No mutable Parser instance is declared Shareable by this unit.

## Questions for system_adamic

These questions are undecided; this implementation does not answer them.

- May a worker-private mutable graph be returned as a task result through an
  ownership transfer, and what result-transfer proof is required?
- May an immutable class method call another method through this or capture this
  in a nested closure when the complete transitive graph is race-free?
- Is a closed-world all-target effect proof sufficient for inherited/interface
  dispatch, or should a language-level final/sealed restriction be required?
- May a readonly structural view cross the boundary while a mutable alias exists,
  and what whole-program alias proof must establish that no worker can write it?
- May getters, static members or method replacements participate when their
  complete target/effect sets are known, or should those remain excluded?

## Fixtures and mutations

Accepted fixtures are registered automatically by the existing concurrency
fixture glob and are held to source Node, JavaScript, release native, ASan/UBSan,
slab variants, leak checks and TSan. Both use 128 tasks; the item fixture repeats
one receiver, ensuring workers really share the same graph.

- task_class_fields.a: captured immutable Header, runtime name string, deeply
  readonly offsets, method argument and this.offsets.join(). Its source mutant
  makes offsets mutable and pushes through this. The task capture Shareable
  check refuses header.offsets, not a TypeScript error or clang warning.
- task_class_items.a: immutable Position task item, runtime path string and
  scalar start. Its source mutant makes start mutable and increments this.start.
  The Shareable item check refuses Position.start.
- task_parser_shared.a: readonly scanner slot with mutable scanner contents;
  Node prints source|source. The exact task-capture refusal is recorded in .what.
  A this.scanner.pos++ mutation remains refused.
- task_parser_local.a: new Parser per task with this.pos++ in parse; Node prints
  first:1|second:1. The known-constructor/ownership proof is missing, so this
  remains refused with an exact .what. A second write-through-this mutation is
  also refused. No native execution is claimed for refused shapes.

Permanent lower tests also refuse global writes in an otherwise immutable method,
this aliases, this arguments and nested this captures. Declaration/expression
subclasses and source method replacement cannot bypass dispatch exclusion.

The runtime mutation injects `this->slots[0].number += 1` into Position.read's
emitted C, after the proof. Clang successfully builds it with TSan. At four
threads TSan reports a data race in Position_read's shared receiver field.
This is the observed unchecked-admission failure, not a hypothetical race claim.
Healthy explicit TSan tests run each accepted fixture three times at each of
1, 4 and 16 threads and compare byte-for-byte with Node, with perturbation on.
The explicit test passed in 1.160s; the race mutant passed its catcher in 0.282s.

Counts added (no existing numeric row changed): Header fixture 780 allocations,
780 frees, 263 retains, 911 releases, peak 142; Position fixture 394 allocations,
394 frees, 260 retains, 397 releases, peak 137. Both have zero regions/slab
allocations/frees in the ordinary counts table.

## Environment and validation

nproc=5, cgroup quota=4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0.
Initial origin fetch obtained the runtime branch but its recursive TypeScript
submodule fetch stalled; only that optional fetch process was stopped, giving
an early-EOF/signal-15 diagnostic. Setup fetched the required exact submodule
commit successfully. Its first cache warm overlapped helper-file creation and
failed with readonlyThis/readonlyMethod undefined. Setup was rerun on settled
source and passed: node 0.017s, go 0.018s, submodules 0.047s, markdown 0.059s,
clang 0.132s, cache warm 38.898s, total 38.924s. Environment file:
/workspace/adamic-tools/env.sh, sourced in every build/test shell.

Exploratory mistakes were corrected: the first receiver mutant also changed the
outer array and was caught for that unrelated reason; final mutation changes
only the receiver graph. The first explicit TSan test used relative paths and
failed sourceIdentity path normalization before running native; final paths are
absolute. No safety assertion or sanitizer check was weakened.

Commands and logs (all test output redirected to files):

- `go run ./notes/runtime-step37-classes/inventory`: inventory.json; diagnostics
  /tmp/step37-inventory.log (empty).
- `go test ./internal/lower -run '^TestTask' -count=1 -v -timeout 10m`: pass
  0.193s, /tmp/step37-proof-final.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  'TestNativeAgreesWithNode/internal/oracle/testdata/concurrency/accepted/task_class'
  -count=1 -v -timeout 15m`: pass 24.180s, /tmp/step37-fixtures.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestTaskReadonlyClassesThreadSanitizer$' -count=1 -v -timeout 15m`: pass
  1.160s, /tmp/step37-tsan-final.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestTaskClassWriteRaceMutant$' -count=1 -v -timeout 10m`: pass 0.282s,
  /tmp/step37-race-mutant.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  'TestTaskParserShapeNodeWitnesses|TestConcurrencyRefusals/task_parser'
  -count=1 -v -timeout 10m`: pass 0.379s, /tmp/step37-parser-shapes.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1
  -timeout 15m -args -update-counts`: pass 78.053s, /tmp/step37-counts.log.

The first readonly commit stopped at task-local construction/graph ownership.
The ruling and proof below address that blocker. No end-to-end parallel parser,
inherited method admission, move-class admission or full stage1 repository gate
is claimed.

## October 8 ruling and worker-private extension

Readonly commit: 12a6b288fcd84536e39eed8fb311e59f529ed89a. system_adamic's
October 8 step-37 ruling, supplied by the user, is:

> 1. Yes: class constructors and methods are admitted on a receiver proven
> worker-private (allocated inside the task, not reachable from anything the
> task shares). Mutating a private value is fine.

This answers the construction/mutation question, not the undecided result,
closure, hierarchy or accessor questions above. The implemented proof follows
that ruling without making a readonly type out of mutable parser state.

parallel.go now calls taskFunction at the shared task boundary and when gathering
runtime shared roots. The established shared proof runs first. When it cannot
prove the task, parallel_private.go:37 tries a separate worker-private graph
proof; it is used only when a source-class construction is encountered.
Non-class tasks retain their established proof and diagnostics.

- parallel_private.go:112 tracks allocation sites, not counts, liveness or an
  assumption that a parameter is owned. Inputs and captures start shared;
  primitives (including runtime strings) have immutable value semantics.
- parallel_private.go:92 joins every allocation origin and the shared bit.
  Function frames, locals, parameters, this receivers, return values, object
  fields and collection elements keep those origins across branches and loop
  trips. The whole task is revisited to a monotone fixed point. A shared origin
  never becomes private after rebinding. Non-convergence is refused after 32
  passes. Assertions and optional wrappers keep origins.
- parallel_private_calls.go:9 requires a known source-class constructor with no
  unproven heritage. It inspects instance field initializers and the constructor
  body, with this bound to the new allocation. Nested constructors form private
  graph edges. Scalar constructor arguments and shared immutable children are
  allowed; shared children remain shared. Recursive construction is refused
  with a diagnostic rather than risking a compiler recursion panic.
- parallel_private_calls.go:120 follows the actual receiver allocation origins.
  Every possible origin must be private or separately admitted by the readonly
  proof, and every private method target must have a known source body. Private
  interface views can work because the concrete allocations supply the complete
  target set. Heritage, replaced/extracted methods, unknown callees and accessors
  are refused. Recursive private calls currently require ownership summaries
  and remain refused; callback closure values also remain conservative.
- parallel_private.go:419 permits field/element mutation only on private origins.
  Stores to captures, globals or shared receivers are refused even if the stored
  value was freshly constructed. Constructor and method helper calls propagate
  argument/return origins, so a helper cannot hide publication. A private parent
  pointing at a shared readonly array does not authorize mutation of that array.
- Scalar expressions are traversed through syntax wrappers (including template
  spans) to inspect their effects. Library calls use an explicit effect list;
  unknown calls and object coercion dispatch remain refused. Primitive library
  constants used by scanner defaults are allowed. No native runtime safety check
  or sanitizer was disabled.

Only Shareable task results are admitted by this extension. Mutable private
objects may be returned between helpers within the worker, as in Parser.node(),
but this unit does not add an unchecked graph escape at the task boundary.

## Widened fixtures and observed mutants

The previously refused task_parser_local.a is now accepted under the ruling,
with identical source and Node output first:1|second:1. Its publication mutant
stores the private Parser into an outer capture and is refused.

task_private_parser.a is a parser-shaped task with 128 independent input strings.
It includes constructor-to-method calls, new Scanner inside Parser's constructor,
array field initializers, scanner position/kind writes, private ParseNode
construction, this.nodes.push, and this.node(index).end mutation. The node helper
returns a private child; its caller's write keeps that provenance. The output
agrees with original Node, JavaScript, release C, ASan/UBSan, slab builds and TSan.
Both private fixtures run three times at each of 1, 4 and 16 threads in the
permanent TestTaskWorkerPrivateClassesThreadSanitizer.

Permanent source mutants and their observed catchers:

| Mutation | Catcher |
|---|---|
| Use one mutable Parser capture instead of per-task construction | Refused: sharedParser.scanner.pos is not Shareable |
| Store the constructed Parser into an outer capture | Private proof: publishes into a capture or global |
| Publish this from the constructor | Same private publication refusal |
| Publish this.scanner from a method | Same private publication refusal |
| Pass a shared readonly array into a private object, cast its view and push | Private proof: mutates a shared collection |
| Rebind a loop/branch alias from private array to shared items, then push | Joined origins retain the shared bit; mutation refused |
| Extract a method and invoke it without a proven target/receiver | Unknown reference/method dispatch refused |
| Mutate a global inside a template span | Private effect walk refuses publication/write |
| Publish a private argument through a named helper | Helper frame rejects the capture store |
| Recursively construct a class in its field initializer | Recursive construction refusal, no compiler panic |

The C admission mutant replaces every task's Parser_new call with a retained
pointer to one Parser allocated before the pool, then calls adamic_share on that
graph. This ensures the object's counts are correctly shared; the mutation is
sharing mutable parser state, not an unrelated counter race. TSan reports a data
race in adamic_array_push at four threads. The existing readonly receiver-write
C mutant also remains caught. Healthy tasks and the shared-receiver mutation
are separate executions; no mutated output is claimed to match Node.

The initial broad readonly concurrency gate found a stale refusal fixture:
dispatch.a was exactly the newly admitted immutable-class shape. That gate
failed only that expectation. The final dispatch refusal now contains a derived
class, so its hierarchy is outside the known-target subset; the exact prior
method-dispatch diagnostic remains required. Its dedicated check and the final
uncached concurrency gate pass. No output mismatch or race failure was waived.

New private counts: task_parser_local has 14 allocations/14 frees, 10 retains,
17 releases, peak 10; task_private_parser has 2,073 allocations/2,073 frees,
7,314 retains, 7,704 releases, peak 275. Zero regions/slabs. No existing numeric
count changed. Count refresh passed in 72.641s.

## Actual scanner probe and remaining boundary

private_scanner.a imports stage1/typescript/scanner/scanner.ts unchanged and
constructs a Scanner inside each task. Original Node prints
ConstKeyword|LetKeyword. A permanent Node witness and exact refusal test passed
in 0.415s. The compiler now reaches scanner.ts:652:29 and refuses:

```text
task capture 'keywords' is not shareable: keywords is a mutable Map
```

scanner/tokens.ts:2 declares keywords as an unqualified new Map; punctuators
at tokens.ts:91 has the same mutable capture type. This is an observation of a
remaining boundary, not permission to smuggle a Map across it. A prior probe
stopped at Number.MAX_SAFE_INTEGER in a default argument; the final proof
recognizes the immutable primitive library constant and reaches the Map.
Stage1's token tables were not edited. Their Shareable views/proofs and recursive
private-call/callback summaries are the named remaining work before claiming
end-to-end parallel parsing.

## Widened validation

All outputs are log files, never piped from test execution.

- `go test ./internal/lower -run 'TestWorkerPrivate|TestTask' -count=1 -v
  -timeout 10m`: pass 0.409s, /tmp/step37-private-proof-settled.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  'TestTaskParserShapeNodeWitnesses|TestNativeAgreesWithNode/internal/oracle/testdata/concurrency/accepted/task_(parser_local|private_parser)'
  -count=1 -v -timeout 15m`: pass 1.593s, /tmp/step37-private-fixtures.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  'TestTaskPrivateReceiverSharedMutant|TestTaskWorkerPrivateClassesThreadSanitizer'
  -count=1 -v -timeout 15m`: pass 1.146s including the caught shared receiver
  mutation, /tmp/step37-private-race-final.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestTaskActualScannerShareableBoundary$' -count=1 -v -timeout 10m`: pass
  0.415s, /tmp/step37-actual-scanner-witness.log.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1
  -timeout 15m -args -update-counts`: pass 72.641s,
  /tmp/step37-private-counts.log.
- Initial `ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/ir
  ./internal/lower -count=1 -timeout 15m`: pass native 670.908s, ir 35.166s,
  lower 48.202s, /tmp/step37-packages.log. This preceded the widened private proof.
- Final `ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/ir -count=1
  -timeout 15m`: pass lower 58.669s, ir 37.604s,
  /tmp/step37-private-packages.log. No native source changed; the new C programs
  were rebuilt in all their oracle/sanitizer variants.
- Final `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  'TestTask|TestConcurrency|TestNativeAgreesWithNode/internal/oracle/testdata/concurrency|TestCountsAreRecorded'
  -count=1 -timeout 20m`: pass 102.175s,
  /tmp/step37-private-oracle-final.log. This covers every concurrency fixture,
  task/refusal mutant, explicit 1/4/16 TSan run and the entire recorded count table.

The exploratory broad gate including TestParallel schedules is recorded in
/tmp/step37-oracle-final.log (746.387s, stale dispatch refusal as described
above). It is not claimed green. The final gate above intentionally omits the
large existing parallel schedule benchmark suite. No full repository/stage1
uncached gate, full-oracle run or macOS verification is claimed for this unit.
