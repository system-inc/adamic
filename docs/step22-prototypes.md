# Step 22: live prototype links

Status: **interim prototype refusal retained**, as directed by the amended
October 8 point 3. Prototype edges have not yet joined the graph-region path.
The research and fixtures are complete; production prototype runtime work is deferred.

Branch `runtime/step22-prototype-links` starts at `origin/area/runtime`
`9bc8201f1c97c4c18a03d263a40ece933ffa752e`. This report records the October 8
point 3 amendment and its evidence. The initial counted-link premise failed on
mixed-edge ownership cycles; the amendment explicitly puts prototype links under
step 06 graph ownership instead.
No runtime C, runtime header, lowering or emission file is changed by this unit.

## Amended rule

system_adamic, October 8, amended step 22 point 3:

> a prototype link is an ordinary graph edge under step 06's cycle rules, nothing special. Where the cycle machinery can place it (both ends in the Program region, or a graph region that merges the two), it's admitted; where it can't yet, it's refused, naming the field and the link... Your interim refusal stands, with a fixture per debug.ts site saying which side it falls on, until prototype edges ride the graph-region path. Never a silent counted cycle.

Admitted prototype edges therefore use ordinary graph-region ownership and must
be included in allocation classification and owned-child enumeration. Strong
edges between graph members merge their regions and do not add outside counts.
Counted boundary references retain their existing ownership. ECMA-262's separate
prototype-chain check remains necessary: even a graph region must reject a cyclic
prototype chain with `TypeError: Cyclic __proto__ value`, matching Node. The Node
error fixture pins this requirement; no native prototype TypeError implementation
is claimed while those calls are refused.

## Why ordinary counted admission would leak

ECMA-262 OrdinarySetPrototypeOf, section 10.1.1.3, walks only `[[Prototype]]`
links when rejecting a cyclic chain. It does not walk ordinary data properties.
Node v24.19.0 accepts this program and prints `true`:

```javascript
const prototype = {};
const object = Object.create(prototype);
prototype.child = object;
console.log(object.child === object);
```

The prototype chain is `object -> prototype -> Object.prototype -> null` and
has no cycle. Counted ownership nevertheless cycles: `object` owns `prototype`
through its prototype link, and `prototype` owns `object` through `child`.
Both counts remain one after both locals go away. The typed probe uses separate
PrototypeOwner and PrototypeInstance interfaces: its declared data-slot type
graph has only Owner -> Instance (and primitive id), so the return link is
introduced exclusively by the prototype operation. Rejecting this as a cyclic
prototype chain would disagree with Node. Keeping the prototype weak would
violate the ruling and could free it while the object still needs inherited
properties. A prototype-only cycle check cannot make ordinary counting leak-clean.

`TestPrototypeMixedCountedCycleProbe` is the requested admission mutant for
`a = {}; b = { a }; Object.setPrototypeOf(a, b)`. Its `object` variable models
`a`, and `prototype` models `b`; two ordinary counted slots model `b.a` and
`a.[[Prototype]]`. It bypasses the compiler refusal on purpose, using the actual
retain/release runtime. This is an ownership model, **not an implementation of
Object.create or Object.setPrototypeOf**. It prints the same
`true` as Node, has no ASan or UBSan error when leak detection is off, and reports:

```text
adamic: counts: allocations 2 frees 0 retains 2 releases 2 peak 2 regions 0
```

The shared `internal/leakcheck` helper reports an indirect leak of **112 bytes in
two objects** on Linux. To make this isolated model's explicit dropping of both
outside references observable, its `__lsan_default_options` disables conservative
stack and register roots. The initial default-root run did not report a leak:
stale pointer bits masked it, while the counted build already proved neither
object was freed. No production sanitizer configuration is changed.

Two controls distinguish the cause. Removing the ordinary field back edge (`b.a`, slot zero in the C model) frees both
objects and defeats the required leak observation. Making both allocations graph
members and using `adamic_graph_hold` for both edges preserves Node's output and
is leak-clean. Thus graph ownership can carry the shape, but literal counted
prototype edges alone cannot. In particular, a counted retain on a prototype
that belongs to the receiver's own graph region would keep an outside count on
that region; it must use the same graph-aware internal-edge handling as fields.

The compiler must account for prototype edges in the full ownership graph before
allocation classification, or prove the admitted call sites cannot create a
mixed cycle. The existing source-level prototype refusal remains in place. This
unit does not install a new refusal that would contradict Node or silently admit
a leak. Under the amendment, the interim refusal is deliberately retained until those
edges are placed by the cycle machinery. This branch exposes no partial API
that could silently accept the counted cycle.

## Compiler census

Pinned source: TypeScript 6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8db46`, `src/compiler`.
Scout source: `codex/scout-22-tsc-objects` at
`6945a8a40b2a288b97e969ee9626ff3a77c8db46`,
`stage3/scout/22-tsc-objects/census.cjs`. All file and line references below use
that upstream TypeScript checkout, not Adamic's ported stage-1 inputs.

The scout parses 77 non-generated compiler `.ts` files. Its program reports 66
checker diagnostics in this checkout without the pinned development declarations;
this is an AST-site census, not a claim that this checkout type-checks cleanly.
The census separates executable operations from emitted JavaScript helper text.
A whole-directory text search corroborates every token occurrence, including
comments and the interpolated helper the scout does not parse.

Executable operations: **four Object.create calls; three Object.setPrototypeOf
calls and two availability tests; no live `__proto__` link access or mutation**.
There are four executable comparisons against the string `__proto__`, which
inspect syntax rather than mutate a JavaScript object's prototype.

| Site | Count | Use and inherited reads |
| --- | ---: | --- |
| debug.ts:554 | 1 | `typeof Object.setPrototypeOf` availability check. |
| debug.ts:558 | 1 | Flow debug prototype inherits from Object.prototype. |
| debug.ts:561 | 1 | FlowNode points to that shared prototype. Inherits `__tsDebuggerDisplay`, `__debugFlowFlags` getter, and `__debugToString`; bodies read the original node's `flags` and flow edges, not the prototype's fields (definitions at 518-549). The `in` check at 518 also searches inherited properties. |
| debug.ts:594 | 1 | Second `typeof Object.setPrototypeOf` check. |
| debug.ts:598 | 1 | NodeArray debug prototype inherits from Array.prototype. |
| debug.ts:601 | 1 | Actual array points to that prototype. Inherits `__tsDebuggerDisplay` (definition 573), while retaining Array.prototype operations. Its body receives the actual array as `this`. |
| debug.ts:859 | 1 | Mapper points to DebugTypeMapper.prototype, inheriting `__debugToString`. The method at 828-851 reads `kind`, `debugInfo`, `source`, `target`, `sources`, `targets`, `mapper1`, and `mapper2` on the original mapper. |
| debug.ts:944 | 1 | `Object.create(null)` is a dictionary of numeric flow-node ids. `links[id]` at 1004 and 1021 reads own entries only; no inherited reads. |
| factory/nodeFactory.ts:6096 | 1 | Redirected SourceFile inherits all non-overridden SourceFile fields from `redirectInfo.redirectTarget`, live rather than copied. `id` and `symbol` own accessors at 6099-6111 forward to the target. `redirectInfo` is own. program.ts:3509-3519 overrides fileName, path, resolvedPath, originalFileName, packageJsonLocations and packageJsonScope. Remaining fields, including text/statements, remain inherited. |
| utilities.ts:5470, 5471 | 2 | Compare an AST identifier/string-literal name with `__proto__`; no object link read. |
| transformers/jsx.ts:262 | 2 | The same two syntax-name comparisons, on one line. |

Emitted helper text is not live compiler prototype manipulation:

| Site in factory/emitHelpers.ts | Count | Emitted use |
| --- | ---: | --- |
| 855 | 1 Object.create call | Async generator wrapper inherits AsyncIterator.prototype, or Object.prototype fallback; inherits iterator methods/identity while next/throw/return and Symbol.asyncIterator are installed as own properties. |
| 942 | 1 Object.setPrototypeOf reference | Feature-selected static class inheritance implementation. Derived constructor inherits base static properties. |
| 943 | 2 `__proto__` occurrences | Prototype-setting object-literal feature probe `{ __proto__: [] }`; fallback setter `d.__proto__ = b` for static inheritance. |
| 953 | 1 Object.create call | Null base gives a null instance prototype; non-null branch inherits base instance members via `new __()`, with own constructor. |
| 1121 | 1 Object.create call | Generator wrapper inherits Iterator.prototype or Object.prototype fallback; resume operations are own properties. |
| 1157, 1176 | 2 Object.create references | Availability tests for module binding/property-descriptor helper branches; no Object.create call and no chain read here. |
| 1480 | 1 Object.create call | Null-prototype cache for advanced async-super getter/setter adapters. Own `cache[name]` entries only. This is an interpolated `helperString` template, omitted from the scout's helper counts and included here by source inspection. |

Comments only: `Object.create` at checker.ts:31721, checker.ts:31744 and
transformers/es2017.ts:1061; `__proto__` at utilitiesPublic.ts:841 (one occurrence),
parser.ts:2646 (one), utilities.ts:5464 (one), utilities.ts:5466 (two).
Across all text these totals are **13 Object.create occurrences on 13 lines,
six Object.setPrototypeOf occurrences on six lines, and 11 `__proto__`
occurrences on eight lines**. Of the 13 Object.create tokens, four are live calls,
four are emitted calls, two are emitted availability references and three are
comments. The scout reports only five emitted Object.create tokens because of
the interpolated template omission above.

## Per-site interim fixtures

`TestPrototypeDebugSiteRefusals` holds every `debug.ts` site to Node and the
current lowerer. Fixtures are under `internal/native/testdata/prototypes/sites`.
No prototype-link operation is currently admitted. Each operation is isolated
so an earlier Object.create, unchecked cast or descriptor call cannot mask the
setter being tested. The aggregate `debug_shapes.a` outside Node fixture covers
the complete inherited getter/method behavior separately.

| Fixture | Upstream site | Current result and reason |
| --- | --- | --- |
| debug_554.a | debug.ts:554 | Refused: reading setPrototypeOf as a value triggers `unbound-method`. This feature test introduces no ownership edge. |
| debug_558.a | debug.ts:558 | Refused Object.create: flowNodeProto.[[Prototype]] -> Object.prototype is not lowered as a graph edge yet. |
| debug_561.a | debug.ts:561 | Refused Object.setPrototypeOf: flowNode.[[Prototype]] -> flowNodeProto is not lowered as a graph edge yet; __debugFlowFlags lives on the shared prototype. |
| debug_594.a | debug.ts:594 | Refused: second availability read triggers `unbound-method`; no ownership edge. |
| debug_598.a | debug.ts:598 | Refused Object.create: nodeArrayProto.[[Prototype]] -> Array.prototype is not lowered as a graph edge yet. |
| debug_601.a | debug.ts:601 | Refused Object.setPrototypeOf: array.[[Prototype]] -> nodeArrayProto is not lowered as a graph edge yet; __tsDebuggerDisplay is inherited. |
| debug_859.a | debug.ts:859 | Refused Object.setPrototypeOf: mapper.[[Prototype]] -> DebugTypeMapper.prototype is not lowered as a graph edge yet; __debugToString is inherited. |
| debug_944.a | debug.ts:944 | Refused Object.create: even the null-prototype creation API is not lowered yet. There is no non-null target to merge or count. |

For each link fixture, the current diagnostic names Object.create or
Object.setPrototypeOf and gives the unchanged interim reason: "prototypes expose
or replace fields outside the declared shape; use a declared object or class with
composition". The table names the receiver field/link for handoff; the current
blanket diagnostic does not yet name a particular mixed-cycle data field. A
future selective refusal must do so under the amendment. Removing each actual
operation makes its control program lower successfully and defeats that refusal
observation; replacing either feature test with a plain constant does likewise.

`mixed_set_cycle.a` spells the ruling's exact mixed cycle and is refused by
Object.setPrototypeOf. Its ordinary-counted admission mutant prints Node's `true`
and has no ASan/UBSan fault, but the shared leak check catches its two leaked
objects. The graph-owned control prints `true` and is leak-clean. No counted
prototype admission is silently installed by this unit.

## Node's exact behavior

Observed on Node v24.19.0 on Linux. Error strings below include name plus message;
`error.message` itself omits the `TypeError: ` prefix.

| Operation | Observation |
| --- | --- |
| Object.setPrototypeOf(a, b), where b inherits from a | `TypeError: Cyclic __proto__ value` |
| `a.__proto__ = b` for that same chain | `TypeError: Cyclic __proto__ value` |
| Object.create(1), Object.setPrototypeOf({}, 1) | `TypeError: Object prototype may only be an Object or null: 1` |
| Same APIs, prototype undefined | `TypeError: Object prototype may only be an Object or null: undefined` |
| Same APIs, prototype true / "abc" / Symbol("p") / 1n / NaN | Same prefix followed by `true` / `abc` / `Symbol(p)` / `1` / `NaN`, respectively. |
| Object.setPrototypeOf(null, {}) | `TypeError: Object.setPrototypeOf called on null or undefined` |
| `a.__proto__ = 1` | Ignored; a's prototype is unchanged. This setter differs from Object.setPrototypeOf. |
| Object.setPrototypeOf(1, {}) | Returns primitive `1`; no change. |
| Frozen receiver, already has the requested prototype | Returns the receiver successfully. |
| Object.setPrototypeOf(Object.freeze({}), null) | `TypeError: #<Object> is not extensible` |

Cyclic-chain rejection must leave the old link and its count unchanged. A
successful replacement must retain the new prototype before releasing the old
one. A same-prototype call must not manufacture a new count. Own writes shadow
inherited data and cannot mutate a prototype slot returned by a read helper.

## Pending runtime contract, not implemented

No entry point below is declared or linkable on this branch. These are the
compiler-facing requirements to settle after mixed-edge ownership is addressed;
there is no partial runtime ABI for the compiler to start calling.

```c
// Object.create for a compiler-proven object-or-null prototype, borrowed input,
// one owned new receiver. Own shape excludes inherited properties. NULL means
// JavaScript null, never undefined. Allocation must already have the graph
// classification required by the full ownership graph, including this edge;
// graph selects graph allocation before any owned prototype child is installed.
adamic_object *adamic_prototype_create(const adamic_shape *own_shape,
                                      adamic_heap *prototype, bool graph);

// Object.setPrototypeOf for an object receiver and object-or-null prototype,
// both borrowed. On success return true; on TypeError return false and put one
// owned Error-shaped value named TypeError in adamic_thrown. Receiver identity
// and its old prototype are unchanged on failure. No result retain is implicit.
bool adamic_prototype_set(adamic_heap *receiver, adamic_heap *prototype);

// For compiler-detected invalid prototype arguments: preserve JavaScript's
// null/undefined distinction and format the primitive as Node does. Rendered
// input is borrowed; set an owned TypeError in adamic_thrown, then return.
void adamic_prototype_bad_argument(const adamic_string *rendered);

// Return the borrowed prototype, or NULL at the end of a chain. Lookup must not
// retain a prototype per read. Object.prototype and Array.prototype need their
// actual builtin identities, not fabricated per-receiver copies.
adamic_heap *adamic_prototype_of(const adamic_heap *receiver);

// A read searches receiver then its live prototype chain. Missing is undefined.
// Getters and methods use receiver as this, not the object owning the descriptor.
// Reference data results need receiver/holder-aware graph escaping; getter
// results already have the getter's ownership convention. A resolved-result
// descriptor must state both representation and ownership, since adamic_value
// alone cannot distinguish a number from a reference.
// Proposed result structure and read signature remain unsettled, not ABI.

// During destruction, detach optional link storage and enumerate its strong
// child using the same release callback as ordinary fields. Graph teardown must
// skip same-region edges; statement arenas and array cleanup must participate.
void adamic_prototype_free_children(adamic_heap *receiver,
                                   void (*release)(void *));
```

Optional link storage must add no bytes or allocation to objects without links.
The ordinary `adamic_object` header is currently 40 bytes on this machine; this
unit leaves its size and allocation policy unchanged. A side record or optional
extension needs its own ownership, destruction and concurrency proof. Adding an
unconditional pointer to the object header would violate the requested cost.

Existing own-property caches cannot cache an inherited slot under the receiver's
shape: two receivers can have identical shapes and different prototypes, and
setPrototypeOf can change the result while retaining the shape. A missing-own
cache is likewise not a permanent missing-chain cache. Numeric optional reads
must use the holder's shape representation, not subtract a prototype's slot
pointer from the receiver's slots. Prototype reads must not turn own-property
queries, spread, or assignment targets into chain reads.

The debug shapes also require prototype-backed accessors/method descriptors,
array receivers and builtin Array.prototype behavior. A data-slot-only walk
would not implement debug.ts. Existing class identity/virtual tables are separate
from these live links and cannot be silently substituted for them. `__proto__`
string keys in null dictionaries remain own data keys, not unconditional setters.

Exception support exists as `adamic_thrown`, Error-shaped `name`/`message`, and
emitted cleanup paths after calls marked throwing. It is not an automatic catch
for a C panic: new library-call effects must enter that cleanup graph. The current
catch narrowing specially handles `instanceof Error`; this unit does not claim
nominal `instanceof TypeError` support. Generic invalid receiver/primitive handling
and its null/undefined distinction belong in the lowering contract, not an
ambiguous NULL pointer ABI. In `.a`, mutating a prototype after creation remains
outside the sound subset as ruled. `.ts` admission is the compiler worker's work.

## Evidence and exact commands

`TestPrototypeNodeResearch` runs three `.a` inputs as outside Node references:
`debug_shapes` first, then exact errors, then the mixed cycle. None is registered
as an accepted Adamic oracle fixture. The debug fixture verifies inherited
getter/method receiver identity, changed flags, array inheritance, mapper methods,
a null dictionary and live redirected text. Mutants change the flags read through
the debug getter, replace the cycle-closing operation with a null prototype, and
remove the mixed cycle's real back edge. Each changes the required output. These
are fixture-input mutants, not evidence of a completed native prototype check.

`TestPrototypeMixedCountedCycleProbe` runs the admission mutant counted, with ASan+UBSan,
and through `internal/leakcheck`; then runs the no-back-edge control and the
existing graph-ownership control. Both controls are leak-clean; the graph control
also matches Node's output. This test intentionally expects the ownership-model
leak as evidence of the blocker, not as accepted native behavior.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step22-prototypes-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Extract only the scout script; do not merge compiler-owned lowering changes.
git show origin/codex/scout-22-tsc-objects:stage3/scout/22-tsc-objects/census.cjs > /tmp/step22-census.cjs
CENSUS_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript \
  node /tmp/step22-census.cjs /workspace/scratch/typescript-6.0.3 \
  /tmp/step22-census.json > /tmp/step22-census.log 2>&1
go test -count=1 -timeout 10m \
  -run '^TestPrototype' -v ./internal/native \
  > /tmp/step22-prototypes-probes.log 2>&1
go run ./cmd/adamic build internal/native/testdata/prototypes/mixed_cycle.a -o /tmp/step22-mixed-cycle \
  > /tmp/step22-prototypes-lowering.log 2>&1
go vet ./internal/native ./internal/leakcheck > /tmp/step22-prototypes-vet.log 2>&1
gofmt -l internal/native/prototype_contract_probe_test.go > /tmp/step22-prototypes-gofmt.log
gofmt -l cmd internal > /tmp/step22-prototypes-gofmt-all.log 2>&1
go vet ./... > /tmp/step22-prototypes-vet-all.log 2>&1
```

The probe test command passes: all three tests, including eight per-site subtests
and their refusal controls. The source build is an expected named refusal;
the first prototype operation is refused with:

```text
Adamic 0.1 refuses Object.create; prototypes expose or replace fields outside the declared shape; use a declared object or class with composition
```

Scoped and repository-wide vet and formatting checks pass. No full gate or three-way
prototype oracle is claimed: implementation stopped before runtime edits, and
native/JavaScript prototype lowering is still refused. Counts rows are unchanged,
with no accepted oracle fixture added.

Setup succeeded on Linux x86_64, Intel Xeon Platinum 8573C, `nproc=5`, cgroup
`cpu.max=400000 100000`, Go 1.27.1, clang 20.1.8, Node v24.19.0.
Setup timing lines are elapsed setup measurements, not estimates:

```text
go ready 0.076s
node ready 0.080s
markdown dependency validation step-duration 0.015s
markdown dependencies ready 0.198s
clang ready 0.421s
submodules ready 230.271s
go build ready 591.322s
test binaries deferred 591.525s
build cache warm 591.530s
done 591.590s
```
