# Shared HIR checkpoint API

Adamic workers import **`../../replay/index.ts`** from `passes/unit-N/`.
The root path is `stage1/cohere/high_level_intermediate_representation/replay/index.ts`.
No worker edits this directory, core, graph, dump, cache, the census exporter,
symbol selectors, or the shared arena index home.

## Read, run a pass, write

1. `readCheckpoint(text)` validates canonical `hir-checkpoint-v1` framing,
   function-nesting paths and every anchored identity. `$` is the root;
   `$/0/1` is child 1 of child 0. IDs are Go's existing function-local IDs,
   not arena slots. Lookup returns private-minted handles with checked reads.
2. `decodeCheckpoint(checkpoint)` constructs **fresh HIR arenas**, without parsing
   source, constructing SSA, invoking earlier passes, or reading expected outputs.
   It restores all 43 instruction variants, patterns, nested functions, phis,
   orphan table entries, outlined identities, source handles and detached lookup
   state. Reserved block IDs survive even when Go's lookup map has no node;
   `blockOrUndefined` is Go's `(Block, found)` lookup. A detached nil terminal
   is represented by `terminalPresent=false`, not an invented terminal.
3. Run the lane's own pass against the `ConstructedHIR`. Import SSA and
   mutation_aliasing algorithms by reference. Passes own their sidecar codecs.
4. `encodeCheckpoint(graph, before, passName, ownedRows, ownedIdentities)` rebuilds
   shared input metadata from the changed arena, retains the census key, and
   adds the lane's current sidecars. It preserves checker facts by exact source
   handle; a new AST source without an actual checker fact stops loudly.
5. `writeCheckpoint(output)` sorts identities and sidecars canonically and prints
   the complete bundle. Compare with the Go after-state, including failures.

`readGraph` / `writeGraph` are the lossless hir-v1 row layer.
`ReplayInstruction.payload` can be replaced there; arena passes normally use
`replaceInstructionValue` on the decoded graph. `Payload` is a checked arena
reader for the JSON payloads used by Go's printer, available to lane codecs.
ScopeIndex and ReactiveIndex are in the same private-constructor home as the
HIR indices; absence is `Index | undefined`. Arena slots never become Go IDs
without going through the reader's maps.

## Wire format and metadata

```
hir-checkpoint-v1
case<TAB>key<TAB>pass
graph-lines N
<N hir-v1 lines, unchanged>
identities N
identity<TAB>function-path<TAB>block|scope|reactive<TAB>Go-ID
sidecars N
sidecar<TAB>namespace<TAB>function-path<TAB>anchor-kind<TAB>Go-ID-or-<-><TAB>key<TAB>payload
end-checkpoint
```

Anchor kinds are function, identifier, declaration, block, instruction, scope,
and reactive. Functions have no numeric anchor ID. Other anchors require an ID.
Sidecar order is namespace, function path, anchor kind, numeric ID, key; identity
order is function path, kind, numeric ID. Duplicate keys and unknown references
fail. Payload content is owned by its namespace, and unknown namespaces survive
round-trip. Percent, controls and non-ASCII UTF-16 units use `%hhhh` escapes.
Invalid Go string bytes use surrogate escape U+DC80–U+DCFF, preserving internal
TypeScript-go byte-prefixed symbol names without replacement characters.

Shared **input** namespaces: input.source (source text, actual ts/tsx/js/jsx kind,
recorded symbol graph), input.ast (Go kind and UTF-8 span or nil), input.checker
(availability), input.types (alias/type symbol presence and names), input.identity
(block allocator high-water), input.blocks (detached map entries and nil terminals),
and input.pattern (pattern/LValue sharing). These are syntax/checker/representation
facts, never effect/reactivity answers. `InputFacts` offers checked generic
`require` / `optional`, `node`, `checkerAvailable`, exact `stableTypeName`,
`calleeOrigin`, source/file-kind/symbol selectors and the callee-signature input
selector. Unit 4 supplies signature payloads under `input.callee-signatures`;
the reader never invents a missing signature. Each analysis namespace supplies
its own tables and scope/reactive identity declarations.

The construction graph still says `scopes -`. Full later scope/reactive state
belongs in the lane's sidecars; that placeholder is not a scope codec.

## Go adapters and observation hook

Overlay these beside the Go HIR package, with `lintoracle`:

- `testdata/oracle_test.go`: unchanged-production Go hir-v1 printer.
- `replay/oracle_test.go`: `OracleCheckpoint`, `OracleExtraIdentity`,
  `OracleSidecarRow`, `OracleWriteCheckpoint` framing.
- `replay/inputs_test.go`: `OracleInputFacts`, `OracleDormantIdentities`,
  `OracleFunctionEncoder`, and the construction observation hook.
- `testdata/census_test.go`: full construction observer/exporter when generating
  the census. Original Go files and production cohere are never edited.

`OracleFunctionEncoder(key, pass, checker, extra)` returns the exact
`func(*Function) ([]byte,error)` that unit 3's `unit3Encode` accepts. The extra
callback runs at each boundary and returns the lane's current identities/rows,
propagating errors. There is no expected-state lookup. The encoder emits AST,
checker and source input facts automatically; symbol transcripts can be supplied
by the caller when needed. `OracleRegisterConstructionObserver` is called during
adapter init. The callback receives `(key, freshClone, checker, originalCaller)`
once for each observed context-distinct graph. It chooses the actual Go pipeline,
including preservation's outline/InferReactive prelude or cache's separate
Construct/drop/inclusive-inline/conditional-Construct path, and invokes the
lane's `unit3Observe` / unit-4 adapter at each real boundary. It must retain
errors and unsuccessful remaps, including parent, nested and captures.
The owner does not replace these pipelines with an invented shared sequence.

Exported `checkpoint-manifest.tsv` keeps all 1,465 originals, including 23 Flow
inputs, plus 72 separately marked probes. All current construction records are
successful Go graphs; pass failures are retained with their input graph and a
lane-owned outcome/error sidecar, never filtered out. `replay-fixture` sidecars
and synthetic scope/reactive IDs are identity-test witnesses, not analysis data.
Do not feed that test namespace to an analysis as inferred scopes.

## Unit 3 copy and cache contracts

`copyInstructionValueWithRemap(source, target, value, mapPlace, mapFunction)`,
`copyPatternWithRemap(source, target, pattern, mapPlace)`, and
`copyTerminalWithRemap(terminal, mapPlace, mapBlock)` return fresh writable record
and array storage. Index translations are explicit callbacks; no cast or bare
numeric minting is needed. Plain copy variants use identical Go IDs in a prepared
target arena. The caller preserves instruction order/range/source when appending
a copied Instruction. `Instruction.value` is replaceable. Outlined mappings are
`IdentifierIndex -> FunctionIndex`; the dump preserves Go's nested ordinal.
`named(..., existingDeclaration)` mints no declaration handle on that call;
default declarations remain Go's new identifier ID + 1. Source slots are required
nullable fields, reflecting Go's always-present nullable Node field.

`LowerFunction(file,index)` returns a fresh **unconstructed** graph.
`ForFunctionWithoutManualMemoization(file,index,drop,inclusiveInline)` is the
owner-owned cache hook. Unit 3 supplies its two plain function callbacks; no
algorithm is copied here. It declines nil checker/node, shares ForFunction for
functions that cannot mention memoization, and otherwise caches a separate fresh
lowering, Construct, drop, inclusive-inline, then conditional Construct when the
inline count is positive. Repeated hits run no pass again. Cooked escaped
identifier/string spellings follow Go's gate. Unit 3 owns actual memo erasure and
inlining; the contract fixture uses explicit no-op callbacks, not a certificate
of those algorithms.

## Request disposition

Read unit 3's `9cc68732` STOPPED.md and unit 4's `9519c348` REPORT.md.
Framing, fresh decoder/encoder, all original Go inputs including Flow, memo
variants, outlined identities, copy/rewrite/cache APIs, Go encoder/observer hook,
AST/type/module-origin selectors and checked scope/reactive handles are supplied.
Pass-specific before/after generation and analysis sidecar codecs remain with
those lanes, connected through the hook above.

**Declined here:** landing or inventing the mutation_aliasing port/API. It belongs
to its existing analysis lane, and the latest fetched lint area `6bf7bcec` still
has no `stage1/cohere/mutation_aliasing/`. No algorithm or guessed exports are
copied. The lane's imports of AliasingEffect, kinds, constructors and MutableRanges
must wait for that module's real landing; ProjectEffects cannot be certified
without it. This is an external dependency, not a replay prerequisite.

## Export the construction checkpoint corpus

From the repository root after setup, run:

```bash
HIR_REPLAY_CENSUS_EXPORT=/tmp/hir-replay-census go test ./stage1/cohere/high_level_intermediate_representation -run '^TestCheckpointReplayOracle$' -count=1 -v
```

This produces checkpoint-manifest.tsv and all 1,465 original checkpoints plus
72 probes, and certifies native/Node fresh decode and re-encode against Go.
Later pass workers use the tagged Go encoder/observer hooks above for their own
actual prepass and postpass checkpoints; construction output does not substitute
for those inputs.
