# HIR compiler gaps — roadmap step 28

Ruling (a), @system_adamic 09:44, authorizes a presence bit beside the boolean
value, as selector gap 3 and values gap 2 do. Mixed-union lowering remains owned
by compiler step 17 (#qgr7mm4). The port applies that representation and keeps
the original proving program intact. `gaps_test.go` asserts the proving program's Node output and the exact
`lower.NotYet.What`; it fails as soon as lowering succeeds or the refusal changes.

## Optional boolean field read

[Shortest proving program](testdata/optional-boolean-gap.a):

```typescript
interface Terminal { readonly optional?: boolean; }
function read(terminal: Terminal): boolean { return terminal.optional ? true : false; }
console.log(read({ optional: true }) ? 'true' : 'false');
```

Node stdout is `true\n`. Exact `lower.NotYet.What`:

```
a field of type boolean | undefined
```

Go `cohere/internal/lint/ecmascript/high_level_intermediate_representation/terminal.go:218`
has `Optional bool`. Adamic's tagged record in `core.ts` has
`readonly optional?: boolean`; the refusal occurs when `dump.ts` reads that field.
The dump reads a presence/value record with required boolean fields. The graph
and hir-v1 meaning remain unchanged. All workaround sites carry `gap 1 (GAPS.md)`
comments in core.ts, lower.ts and dump.ts. When the gap test fails because lowering
succeeds, remove these sites together and recertify the whole corpus.

## Construction corpus impact

The independent Go hir-v1 dumps identify the graphs requiring an Optional terminal
field read, including nested functions included in a root dump. The census keeps
all **1,465** original context-distinct functions and explicitly records Flow
exclusions. An optional computed-load or call flag is a required boolean and does
not itself require the refused field type.

The owned impact test records all matching census keys and provenance, rather
than estimating the affected count from syntax or the current admission grammar.
**74 / 1,465** original graphs require the field, and **1,391** do not. Before
the authorized workaround the whole native executable failed to build, even for
functions that did not execute the Optional branch. The presence/value record
now restores native compilation. The census writes `optional-boolean-impact.json`
with every affected key and original test-call provenance.

## Gap 2: CloneFunction spread with private-index static constructors

Go `clone.go:96` copies `Instruction.Value` (the `InstructionValue` interface),
through `inline_remap.go:271`'s deep copy. Adamic `clone.ts` copies the equivalent
`ValueType` discriminated record, its writable arrays and its arena edges. The
required concrete private-index classes have static minting methods. Stage 0
refuses combining these static constructor objects with object record spreads.
The compiler ruling authorizes explicit typed record copies in clone.ts. Every
workaround region is marked `gap 2 (GAPS.md)`; remove these copies in favor of
spread when the selector-style proving test detects that lowering is fixed.

Shortest proving program, `testdata/static-constructor-spread-gap.a`:

```typescript
class Index { readonly slot: number; private constructor(slot: number) { this.slot = slot; } static push(): Index { return new Index(0); } }
const index = Index.push();
console.log(`${{...{slot: index.slot}}.slot}`);
```

Node stdout: `0\n`. Compiler message verbatim:

```
adamic: static-constructor-spread-gap.a: stage 0 can't lower spreading in a program with static constructor objects yet
```

Exact `lower.NotYet.What`: `spreading in a program with static constructor objects`.
The selector-style test fails when lowering succeeds or its refusal changes.
With the recorded workaround, CloneFunction is certified on native and Node
for all 1,442 admitted originals and 72 probes. The compiler gap remains open;
the unmodified proving program and exact refusal are retained.

## Gap 3: instruction arena views and accessor representation dispatch

Go `cohere/internal/lint/ecmascript/high_level_intermediate_representation/high_level_intermediate_representation.go:387`
defines `Instruction`; its `Value InstructionValue` field is at line 397.
The arena port stores `ValueType[]` columns in `InstructionArena` and exposes
`get value(): ValueType` / `set value(value: ValueType)` on checked
`InstructionIndex` views. The persistent `Instruction[]` has been removed per
the arena ruling. No weak edge, public constructor or numeric brand is used.

The compiler cannot dispatch an accessor and an unrelated plain field with the
same name when their native representations differ. Smallest proving program,
retained in `testdata/instruction-arena-accessor-gap.a`:

```adamic
class View { get value(): number { return 1; } }
const view = new View(); const flag = {value: true};
console.log(`${view.value} ${flag.value}`);
```

Node stdout: `1 true\n`. Compiler message verbatim:

```
adamic: instruction-arena-accessor-gap.a: stage 0 can't lower accessors sharing a name with different native representations yet
```

Exact `lower.NotYet.What`: `accessors sharing a name with different native representations`.
`TestInstructionArenaAccessorGap` fails when the refusal closes or changes.
This is a stop, with no workaround. Construction/clone/replay native certification
on this shared branch cannot proceed; the completed `hir/land-08` landing retains
its independent native = Node certificate. Unit 3 also needs single-read locals
before narrowing instruction getters in its own `optional_sources.ts:70`; that
is the sound subset's ordinary getter rule, not a request to weaken it.
