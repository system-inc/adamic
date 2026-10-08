# runtime-step37-classes

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

- May a task construct and exclusively own a mutable Parser with its Scanner,
  nodes and roots, and what construction/move proof is required for the whole
  graph and for results that return part of it?
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

Named remaining blocker: **task-local class construction and whole reachable
graph ownership**. The mutable stage1 Parser is not admitted. No end-to-end
parallel parser, interface/inherited method admission, move-class admission or
full stage1 repository gate is claimed.
