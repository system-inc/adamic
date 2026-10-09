# Step 20: the iteration protocol

## Evidence and counting

Delivery base: `dcdbb9098f77f30ad41790c56df1bd63ad462b63`, the merged
`origin/compiler/area-next-fixtures` tip. The delivery branch is `codex/scout-iteration`.
The ranking branch `6c4fc1afb019d0a16fea97082e45fef8fcbe1746` is evidence only;
it is not merged into this branch.

A root in these tables is a distinct diagnostic `where` for an exact kind/reason.
An attempted unit is counted separately: recovery can report the same site from
several enclosing units. These are diagnostic roots, not successfully compiled
entry roots. The inventory retains every site and unit in JSON.

Hidden bytes below are historical, outermost-boundary credits from the ranking's
compiler `ed6e29751ee47d86fad450cd1674139883bc0f70`. They are estimates of bytes
revealed if that reason alone were fixed, not bytes compiled by this change.
Zero means zero credited bytes; unmeasured means the ranking has no such row.
The inherited full census also predates this delivery base (`74fb6490...`);
neither artifact is presented as a fresh base measurement.

### Historical hidden ranking

| Kind | Exact reason | Diagnostic roots | Attempted units | Hidden bytes | Diagnostic witnesses |
| --- | --- | ---: | ---: | ---: | --- |
| NotYet | `for...of over an object` | 120 | 124 | 13,952 | `utilities.ts:10736:13`; `binder.ts:375:17`; `binder.ts:432:13` |
| NotYet | `a for...of destructuring an object` | 4 | 4 | 1,106 | `checker.ts:26333:13`; `checker.ts:42067:17`; `program.ts:798:13` |
| NotYet | `for...of over a union of differently held members` | 1 | 1 | 123 | `core.ts:438:5` |
| Refused | `a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)` | 1 | 1 | 109 | `parser.ts:2356:5` |
| NotYet | `a YieldExpression as a statement` | 1 | 1 | 26 | `core.ts:1045:9` |
| NotYet | `iterating a value` | 1 | 1 | 0 | `programDiagnostics.ts:219:17` |

Rows with fewer than three distinct sites list all real sites. Repeating or
inventing witnesses would conceal the size of the evidence. Full boundary records,
ranking provenance and artifact hashes are in [hidden.json](step-20-iteration/hidden.json).
The first row has 141 raw records, 120 distinct boundaries and 25 boundaries with
nonzero credits. The destructuring row has four records and two credited boundaries.

### Refusal table

The ranking bundles the refusal table from `d35a81d36fdafccf827bad0f572d311b2a0d4deb`.
It records `a generator function` at six sites and `yield (generators)` at eight;
all are original-source sites. Neither has a hidden-byte ranking row, so their
bytes are **unmeasured**. Its published examples are `checker.ts:21677`,
`checker.ts:21685` for generator functions and `checker.ts:21681`,
`checker.ts:21693` for yield. The table supplies only two witnesses per reason;
the fresh census below supplies additional real sites where observable.
The owner is `internal/lower/refusals.go` / `lowering.refuse`; the step 20 ruling authorizes replacing these
refusals with correctly implemented generator behavior.

### Scope exclusions

`reading iterator` and `reading iteratorValueStatement` describe variables in
TypeScript's emitter, not proof that those roots need runtime iterator support.
They are retained as contextual candidates in the census JSON, excluded from
step retirement totals. `for...in over an array ... use for...of for elements`
is property enumeration, outside this step. `YieldExpression` AST model views
are not runtime generators. Name matches alone do not assign a root to step 20.

### Fresh base census

The guarded measurement covers all **81 resolved non-library source files** on the
delivery base, from TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`
(v6.0.3), adapted by `stage3/apply.sh`. The entry is `src/tsc/tsc.ts`.
The checker reports **324 diagnostics**: this is a measurement on a rejected
entry, not a claim that the TypeScript compiler builds. No output was emitted.

| Kind | Exact reason | Roots | Units | Historical hidden bytes | Witnesses |
| --- | --- | ---: | ---: | ---: | --- |
| NotYet | `a YieldExpression as a statement` | 1 | 1 | 26 | `core.ts:1045:9` |
| NotYet | `a for...of destructuring an object` | 4 | 4 | 1,106 | `checker.ts:26333:18`; `checker.ts:42067:22`; `program.ts:798:18` |
| NotYet | `for...of over a union of differently held members` | 1 | 3 | 123 | `core.ts:2173:29` |
| NotYet | `for...of over an object` | 125 | 129 | 13,952 | `binder.ts:1296:36`; `binder.ts:2061:33`; `binder.ts:2809:33` |
| NotYet | `iterating a value` | 1 | 1 | 0 | `programDiagnostics.ts:219:39` |
| Refused | `a generator function` | 12 | 12 | unmeasured | `checker.ts:21677:5`; `checker.ts:21685:5`; `checker.ts:21781:30` |
| Refused | `a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read` | 14 | 1 | unmeasured | `sourcemap.ts:504:48`; `sourcemap.ts:505:52`; `sourcemap.ts:511:52` |
| Refused | `optional property return in ArrayIterator<JSDocLink \| JSDocLinkCode \| JSDocLinkPlain \| JSDocText> absent from structural source ArrayIterator<JSDocComment>, which can hide fields` | 1 | 1 | unmeasured | `parser.ts:9508:101` |
| Refused | `optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<JSDocComment>, which can hide fields` | 13 | 5 | unmeasured | `parser.ts:1054:92`; `parser.ts:1060:89`; `parser.ts:1073:89` |
| Refused | `optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<JsonObjectExpressionStatement>, which can hide fields` | 14 | 9 | unmeasured | `commandLineParser.ts:2445:13`; `commandLineParser.ts:2498:65`; `commandLineParser.ts:2503:65` |
| Refused | `yield (generators)` | 15 | 12 | unmeasured | `checker.ts:21681:13`; `checker.ts:21693:17`; `checker.ts:21782:33` |

The ArrayIterator optional-return refusals are iteration-related structural view
boundaries. They must preserve protocol state and exception behavior; removing
a structural check alone would not implement them. The IteratorResult refusal
separately exposes incompatible yield/completion payloads. Historical symbol-key
reflection remains in the table above even though recovery did not reach it here.
`strictBuiltinIteratorReturn` in a CompilerOptions property name is an unrelated
configuration view, excluded along with emitter variable-name matches.

### First shape by census

The source contains 651 for-of expressions. Genuine arrays account for 466
and already have a fast path; 143 are NodeArray views, six are
SortedReadonlyArray views and six are JSDocArray views. These are source loop
counts, separate from diagnostic roots. Matching diagnostic and expression
locations by file and line gives the refused shapes in [shapes.json](step-20-iteration/shapes.json).
NodeArray is the largest unsupported concrete shape and is implemented first.

### Reproduction and audit

The archive [base-census.jsonl.gz](step-20-iteration/base-census.jsonl.gz) has
one checker header and exactly one record per resolved source file.
[base.json](step-20-iteration/base.json) retains every diagnostic site and attempted unit.
The decompressed SHA-256 is
`d1336db1f97fc3ff86471d028d84df4e5a099b2fb11fdedb5a0fa0fa2b7b4c49`.

The initial full process was interrupted by an environment restart after 29
complete file records. The remaining 52 files were measured individually with
a scratch overlay filtering only the outer per-file emission loop; checker files,
declaration registration and attempted-unit recovery remained unchanged. Every
header matches exactly. A repeated corePublic.ts record matched the original
full run exactly. Interrupted processes have no claimed successful exit status.

Commands run (each redirected to its own log):

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/scout-overlay
python3 stage3/meter/entry_overlay.py /tmp/scout-overlay /tmp/scout-entry-overlay
go build -buildvcs=false -overlay=/tmp/scout-entry-overlay/overlay.json -o /tmp/scout-census ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/scout-census /tmp/scout-adapted/src/tsc/tsc.ts /tmp/scout-full.jsonl
python3 docs/step-20-iteration/census.py docs/step-20-iteration/base-census.jsonl.gz /tmp/base.json --compiler dcdbb9098f77f30ad41790c56df1bd63ad462b63
python3 docs/step-20-iteration/hidden_audit.py
```

The fresh extractor audits 14 exact candidate reasons. Mutants that increment
roots, increment units, invent a witness, or omit a source file each fail their
corresponding assertion. The historical audit verifies six reasons against the
pinned ranking; byte, root, witness and boundary mutations each fail. No byte
credits from different compiler bases are added together.


## Ruling and implementation design

@system_adamic's step 20 ruling (#yvgst44) authorizes this design. The original
proposal requirement has been satisfied by that ruling. The implementation must
still prove each represented operation against Node and test262; authorization
does not make an unimplemented operation compile correctly.

The contract follows [0.1.md](0.1.md), [escape-hatches.md](escape-hatches.md) and
[memory.md](memory.md). There is no unchecked cast, erased type, garbage collector
or fallback that guesses an object's layout. Current generator refusals are an
observed compiler outcome, not a request for another policy decision.

| Shape | Representation and lowering |
| --- | --- |
| Array and inherited readonly-array view | Keep the original counted array, identity, element storage and live length. Instantiate interface bases through the checker. Evaluate the RHS once, retain it across mutation and source reassignment, and give each iteration fresh lexical bindings. Keep metadata on the same value; an unsupported metadata operation stops with NotYet. |
| Structural Iterable | Specialize a known concrete type in the closed program; otherwise dispatch its own `[Symbol.iterator]`. Method receivers and arrow fields use the runtime's receiver-aware closure convention. No additional origin proof is required. |
| Iterator record | Own the iterator; look up and cache its `next` once. Keep the yielded T and completed TReturn in separate typed payloads beside `done`. Boolean and undefined-only payloads use their ordinary representations. Read `done` before the selected `value`. |
| Map and Set | Own the collection and maintain a live insertion-order cursor and sticky exhaustion. Saved iterators retain their progress. Entries allocate fresh tuples; the two elements of a Set entry have the same identity. Mutation cannot be replaced by a snapshot. |
| Built-in iterator object | Counted iterator state, stable identity, `[Symbol.iterator]()` returning itself, and protocol methods inherited from a shared iterator prototype. Refuse own-property reflection and copying until their layouts are implemented. |
| String | Counted string and code-point cursor. A supplementary pair yields one string; a lone surrogate yields one string. Preserve UTF-16 behavior for all operations on each yielded string. |
| Typed array | Counted iterator retaining the typed view and backing storage. Read the current indexed value, preserve view bounds and sticky exhaustion, and use the element's ordinary number representation. Unsupported backing-store resize behavior stops explicitly. |
| Object binding | Hold the current yielded object, then lower existing destructuring in source order before the body. Renamed properties retain their names; getters keep their receiver and throw at the corresponding read. Each capture belongs to that iteration. |
| Generator | Counted heap frame containing the program counter, suspension state, typed parameters and live slots, pending completion, and owned delegate iterator. Captures count like closures. Program-region members are plain pointers because that region outlives the frame. |

### Intrinsic proof and storage boundaries

The whole-program refusal pass rejects writes to `Symbol.iterator` on Array,
String, Map, Set and typed-array values or their prototypes, including checker-
identified symbol aliases. This is the proof for their fast paths. Property
expansion and prototype mutation retain their existing boundaries.

`NodeArray<T>` extends `ReadonlyArray<T>` and carries compiler metadata. An array
view preserves its original storage rather than projecting it into a plain object.
Additional fields still participate in ownership compatibility and cycle analysis;
the instantiated element type participates too. A structural record that satisfies
an array interface needs protocol dispatch rather than a native array cast. Until
that path exists, object construction through such a view is NotYet. Unsupported
metadata construction or access also stays NotYet. No metadata is silently copied,
zeroed or discarded. Mutable and intersection array aliases need their own storage
and variance checks before extending this first shape.

### Generator frames

Frame states are new, suspended, executing and completed. The frame's RC header,
resume function and program counter identify which typed slots are live. A live-slot
mask or statically selected destruction path releases exactly the initialized owned
slots. Suspension never keeps a stack address. Reentrant resume must throw as Node
does. Each resume returns a fresh result, with separately typed yield and completion
payloads; a returned value may differ physically from the yielded value.

`next`, `return`, `throw` and `yield*` follow ECMA-262. A pending completion carries
its kind and typed return/error payload through suspended `finally` bodies. A call
to `return()` runs finally, including finally bodies that themselves yield. Delegation
owns the delegate and forwards each operation with the specified presence and
result checks. Throwing into a new or completed frame, returning from a new frame,
and resuming a completed frame have their distinct specified outcomes.

Dropping the last reference to an abandoned frame releases its owned slots but
does not run finally or call user `return`. That distinction is observable on Node.
Frame and closure captures remain subject to the existing owning-cycle refusal
and explicit Weak cuts. Synchronous generators do not authorize async iteration.

### IteratorClose and exceptions

The iterator's `next` is cached once; `return` is looked up at close time. Missing
or null return means no close call. Continue within the loop and normal exhaustion
do not close. Break, return, outer transfer, and a throw during binding or the body
close an active iterator. Failures in iterator advancement, `done` or `value` reads
must be distinguished from a subsequent binding failure.

The incoming throw wins over a failing close, including a failing return lookup.
A close failure replaces a break or return completion. A nonthrowing close must
validate that the return result is an object. Close inner iterators before outer
ones. Cleanup releases owned iterator, result and value references on every path;
reference destruction itself invokes no user code. The existing active-flag lowering
must be audited before admitting additional throwing binding shapes.

### Library setting and retained refusals

The project's TypeScript lib setting decides whether iterator helpers exist. The
compiler does not invent declarations for them. Custom Set-shaped objects that
return generators use protocol dispatch. Optional numeric conditions use exact
JavaScript ToBoolean in both `.ts` and `.a`; explicit comparison is a taste rule.

Non-iterable records remain checker errors. Unsound casts, `any`, owning cycles,
prototype mutation and unsupported symbol-key reflection retain their soundness
boundaries. This structural symbol-key view still needs refusal until its own-key
representation is complete:

```a
const iterable = { value: 1, [Symbol.iterator]() {
    return { next() { return { done: true, value: 1 }; } };
} };
console.log(Object.keys(iterable));
```

Generator admission is authorized, including these previously refused programs;
none may be enabled by skipping the frame or unwind behavior:

```a
function* empty(): Generator<number, void, unknown> {}
function* single(value: number): Generator<number, void, unknown> { yield value; }
function* delegated(values: number[]): Generator<number, void, unknown> {
    try { yield* values; } finally { console.log("close"); }
}
for (const value of single(1)) { console.log(String(value)); break; }
```

### Silent-miscompile audit

- Preserve symbol identity, dynamic receivers, inherited methods, getter effects,
  and cached-next lookup. A source string key cannot impersonate a symbol slot.
- Evaluate the RHS and factory once. Read done before value, binding fields from
  left to right, and allocate fresh bindings and capture cells on each pass.
- Keep array and collection owners alive across pop, splice, clear, deletion,
  reinsertion, source reassignment, callbacks, throws and nested closing.
- Preserve live lengths, insertion-order cursors, sticky exhaustion, fresh entry
  tuples and Set entry identity. Do not materialize the remaining iterator values.
- Keep declared element storage through generic interfaces, including Weak slots;
  narrowing never converts a weak handle into a strong storage slot.
- Preserve metadata, discriminated result payloads, optional done semantics,
  undefined-only payloads, and ABI compatibility at every view and call boundary.
- Distinguish advancement failure from binding/body failure and preserve exception
  precedence through closing, delegation and suspended finally bodies.
- Retain surrogate pairs and lone surrogates exactly as Node does. Byte or code-unit
  iteration is not a substitute for code-point iteration.
- Free an abandoned frame without running finally. A literal string cannot stand
  in for a runtime allocation in an ownership fixture.
- Deduplicate diagnostic sites and attempted units separately. Historical hidden
  credits do not establish that a whole source function or entry now compiles.

Every implemented piece needs source Node, both backends, native sanitizers, release
and leak checks, and a mutant caught by the specific acceptance check. Report other
implementation boundaries directly; do not relabel them as a new policy decision.

## Acceptance fixtures on the base

The 16 fixtures in [outcomes.json](../stage3/fixtures/iteration/outcomes.json)
record exact Node output and compiler outcomes on the delivery base. Two shapes
already compile: concrete Map/Set iterators and strings by code point. Four
readonly-array fixtures stop at object iteration, three object-binding fixtures
stop at destructuring, two generators are Refused, and the structural iterable
and iterator-forwarding reductions retain their exact NotYet reasons. Three
intrinsic override counterexamples record the base's earlier boundaries.

The fixture [README](../stage3/fixtures/iteration/README.md) distinguishes source
reductions from authored stress and ownership probes. The immutable base snapshot
is [baseline-outcomes.json](../stage3/fixtures/iteration/baseline-outcomes.json).
The accepted fixtures pass source Node, the JavaScript backend, native ASan/UBSan,
release builds and leak checks. The wrong-reason snapshot mutant is killed by the
exact diagnostic comparison. The four original test262 cases pass independently
on Node in default and strict modes; this is eight runs, not full-suite coverage.

Verification commands, with complete output redirected to separate files:

```sh
go test ./internal/oracle -run '^TestStep20IterationOutcomes$|^TestNativeAgreesWithNode$/stage3/fixtures/iteration/(collections|strings).a$' -count=1 -timeout 30m -v
python3 stage3/fixtures/iteration/check_snapshot_mutant.py
node stage3/fixtures/iteration/test262_originals.mjs /tmp/scout-test262
npm ci --prefix stage3/api
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

The first counts attempt stopped at missing `@types/node` 25.3.3. The setup script
had not installed `stage3/api`'s locked dependencies; `npm ci --prefix stage3/api`
installed them. The final refresh is recorded separately. An earlier positive
snapshot run overlapped a temporary expected-reason mutation and therefore failed;
the final positive run passed after restoration. The committed mutant runner uses
a Go overlay and never changes the fixture snapshot.

Setup output and verification transcripts are in
[evidence](step-20-iteration/evidence/). Setup reported `nproc=5` and cgroup
`cpu.max=400000 100000`. Its timing lines are retained verbatim in
[setup.txt](step-20-iteration/evidence/setup.txt), including the cold build and
successful environment path. No full package test or full gate was run.

The final base counts refresh passed. It adds the two accepted iteration fixtures
with balanced allocations/frees. It also reconciles the inherited area-tip fixture
registry: `logical_and_reference_maybe.a` moves with its init registration and
`taste/17_binder_flow.a` is omitted because that fixture is registered as not
lowering. No count values for pre-existing counted fixtures change.

## First implementation: inherited readonly arrays

The first piece implements the most common unsupported concrete shape:
`NodeArray<T>` and compatible interfaces inheriting the library ReadonlyArray.
It reuses the existing array IR and both backends. The value keeps its original
array storage and identity. The numeric index must be readonly, mutator members
must be absent, and inherited canonical array members must remain library members.
Interface bases are instantiated through the checker, including generic and
multi-level inheritance. Structural records are stopped before a native array
cast. Metadata is not projected away; unsupported metadata reads stop with NotYet.

Element compatibility, contextual Weak storage, mutable-element variance and
ownership-cycle analysis follow the inherited element type while retaining
additional fields in their checks. The whole-program scan adds the approved
`adamic/intrinsic-iterator` refusal for writes on supported built-ins and their
prototypes, including Symbol.iterator aliases. This changes an implementation
boundary under the ruling, without introducing another language decision.

### Root retirement observations

All 117 candidate sites were replayed before and after, using the original census
entry, source and no-output guards. The official replay selects the smallest
attempted unit containing the site. **116 base signatures reproduced; 110 old
`for...of over an object` signatures disappear after the change.** At 109 sites
there is no replacement diagnostic at that exact location. At `binder.ts:2061:33`,
the union of two NodeArray element types now stops with an explicit array-element
representation gap. It is not a fully supported union loop.

Six JSDocArray sites retain the original diagnostic; their interface inherits
mutable Array and is outside this readonly-array piece. At
`transformers/es2015.ts:3572:32`, the smallest-unit base replay does not reproduce
the census signature: earlier casts and a binary-expression boundary obscure it.
That site is excluded from retirement totals. The full census found it from its
other enclosing recovery context. No root is credited merely because a replay
failed to reproduce on the base.

These are selected-unit diagnostic observations. They do not establish that entire
source functions compile, that the whole tsc entry compiles, or that historical
hidden bytes became emitted code. Other body boundaries remain in the recorded
findings. The unchanged 324 checker diagnostics still reject the complete entry.

[The replay archive](step-20-iteration/array-view-replays.json.gz) contains every
before/after unit and finding, source hashes, and losslessly deduplicated declaration
snapshots. [audit_retirements.py](step-20-iteration/audit_retirements.py) verifies
coverage, base reproduction, exact retirement classification, selected units,
measurement labels and snapshot hashes. Two official single-root records match
batch replay in both phases, including a repeated request after another unit.
A stale scratch expression overlay was discovered during preparation; all its
after results were discarded before the corrected replay. Only the corrected
results are archived.

### Verification and mutants

The final snapshot has 17 fixtures: six accepted, five NotYet and six Refused.
All source programs run on Node. The six accepted fixtures pass the JavaScript
backend, native ASan/UBSan, optimized native builds and leak checks. Thirteen
intrinsic-write cases and the mutable-element counterexample pass their explicit
refusal contracts. The scoped iterator, closure-cycle and mutable-container
regressions pass. Counts adds only the four newly accepted fixture rows, each with
balanced allocations/frees; no existing row changes.

Commands run, each with output redirected to its own log:

```sh
ADAMIC_STEP20_RECORD=/tmp/scout-first-current-outcomes.json go test ./internal/oracle -run '^TestStep20IterationOutcomes$' -count=1 -timeout 30m -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/stage3/fixtures/iteration/|^TestStep20IntrinsicIteratorWrites$|^TestStep20ArrayViewVariance$|^TestStep20IterationOutcomes$' -count=1 -timeout 30m -v
go test ./internal/lower -run '^TestIterator|^TestLiteralMethod|^TestGenericIterator|^TestCensusRestMutableElements$|^TestNested.*Cycle|^TestInheritanceCycleFinderIncludesInheritedFields$' -count=1 -timeout 30m -v
python3 stage3/fixtures/iteration/check_array_view_mutants.py /tmp/scout-delivery-mutants
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
python3 docs/step-20-iteration/audit_retirements.py docs/step-20-iteration/array-view-replays.json.gz
```

| Mutant | Specific catcher |
| --- | --- |
| Snapshot inherited-array RHS | Native and JavaScript stdout disagree with source Node under live mutation |
| Store a strong pointer in contextual Weak element storage | ASan SEGV in adamic_retain; generated C compiles |
| Bypass intrinsic-iterator refusal | Alias fixture returns the wrong diagnostic instead of the ruled Refused contract |
| Omit inherited-array container normalization in variance analysis | Mutable-element counterexample lowers when it must be Refused |
| Increment fresh root count | Root deduplication assertion |
| Increment fresh unit count | Unit deduplication assertion |
| Invent fresh witness | Witness-selection assertion |
| Omit a resolved census source file | Source-coverage assertion |
| Increment historical hidden bytes | Historical byte-credit assertion |
| Increment historical root count | Historical root-deduplication assertion |
| Invent historical witness | Witness-provenance assertion |
| Omit historical boundary | Exact boundary-extraction assertion |
| Record wrong fixture NotYet reason | Exact outcome comparison |
| Omit a replay root | Replay-coverage assertion |
| Flip retirement classification | Exact retirement-classification assertion |
| Invent replay witness | Exact replay-witness assertion |
| Alter shared declaration snapshot | Declaration SHA-256 assertion |

All 17 mutants fail through their intended catchers. The refusal mutants prove
those contracts are enforced; they do not claim that bypassing one check would
make the unsupported program compile. Full transcripts and mutant patches are in
[evidence](step-20-iteration/evidence/). No whole package or full gate was run.

### Remaining implementation scope

This piece does not implement general structural protocol dispatch, generator
frames or yield*, additional IteratorClose paths, object bindings, mutable branded
arrays, the NodeArray union-element gap, builtin iterator prototype/identity work,
or iterator helpers. Their approved contract is above; the remaining work is
implementation and verification. Concrete Map/Set iterator and string fixtures
already compile on the base and remain accepted. The original tests and refusal
snapshots remain explicit acceptance targets for the remaining shapes.
