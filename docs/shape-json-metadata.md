# Runtime JSON object metadata

Runtime half of roadmap step 22. The interface is the whole contract at
`efe905da:docs/library-json-contract.md`, with shared definitions copied from
library `cf2cb58b`. Library's branch was fetched and inspected, not merged.
This change does not remove the lowerer's JSON object/union refusals or implement
an encoder.

## Shared definitions

`internal/native/runtime/json_stringify.h` is byte-identical to `cf2cb58b`, including
`adamic_json_kind`, `adamic_json_schema`, `adamic_json_field`,
`adamic_json_result`, `adamic_json_runtime` and the encoder declarations.
SHA-256: `7e75c5953e8d3f1cbd51066c850bf53b1779eac177f630030a4799dddcee9e9a`.
`TestJSONSharedHeaderMatchesLibrary` guards those bytes. No shared-definition
change needs mirroring. Runtime-only declarations live in `json_metadata.h`.
The tagged-null sentinel declaration and initializer are also copied from library's
`adamic.h` and `union.c`; its address distinguishes null from undefined.

## Producer handoff

The exported `adamic_json_runtime_providers` implements the three callbacks.
The library call that consumes it is:

```c
adamic_json_stringify_runtime(value, schema, replacer, replacer_schema,
                             space, space_schema, &adamic_json_runtime_providers);
```

Descriptors live in immutable static storage. An array owns its optional tag vector,
not the descriptors it points to. `describe` returns an array descriptor with a NULL
`element`; the encoder then calls `array_element`. That reader returns a borrowed,
resolved value/schema pair. Packed optional numbers are unpacked; optional booleans
use heap-tagged boxes. Null, undefined, false, zero and NaN are distinct tags/values;
none supplies evidence of a hole. Explicit undefined tags encode absent array
positions without reading their payload. Arbitrary sparse growth remains refused.

Each shape has its actual JSON schema and a typed hook pointer. Fields record
actual names, slot indexes and child storage descriptors. Reference unions are
resolved by heap tag and actual allocation, so nested containers and changing
method results do not reuse the first result's type. Numeric property names are
ordered as Node orders them, followed by other names in insertion order. IR
private-field flags exclude private storage; a public property named `#public`
remains enumerable. Tuple origins have distinct schemas from ordinary objects
with the same numeric keys.

A proven class method's hook thunk accepts the actual receiver and optional string
key, returning the shared tagged pair. The callback pins a borrowed receiver
through reentrant mutation. Results transfer one owner where applicable; the
encoder keeps the original owned union box through cleanup while decoding its
payload. Undefined results have an explicit descriptor. A thrown hook preserves
`adamic_thrown` and returns described undefined with no owner. The consumer calls
a hook once per position, before encoding, and does not invoke a returned object's
hook again at that position.

| Contract anchor | Producer / implementation |
| --- | --- |
| `adamic.h:358`: array ownership is insufficient | `adamic_array.json_element` and `json_elements`; `array.c:adamic_array_new_typed` takes a descriptor independently of `references`. |
| `json_stringify.c:176`: homogeneous element read | `json_metadata.c:adamic_json_array_element`; returns the slot with its descriptor, resolving packed/boxed values before consumption. |
| New: factory element metadata | `emit_expressions.go`, `emit_arrays.go`, `from.go`, `reuse.go`, `parallel.go`: literal, map, filter, from, fill, spread, reuse and parallel result types are supplied explicitly. |
| New: runtime-created scalar arrays | `string_split_impl.h`, `input.c`, `directory.c`, `record.c`, `class_features.c`, `library_language.c`, `library_array.c`, `regexp.c`: string producers carry string descriptors; optional-number sorting carries its packed-number descriptor. |
| New: collection-derived arrays | `map.c`, `set.c`, `library_object.c`: key/value descriptors travel from the compiler or known key constructor; entries carry tuple field descriptors; actual value schemas are retained. |
| New: copies and writes preserve metadata | `array.c`: slice, concat, append, splice, fill, set, reverse and sort preserve/replace tags with values. Sort moves value/schema pairs, including identical bits with different scalar tags. Reuse changes descriptors as slots change and resets the completed homogeneous result. |
| `json_stringify.h:11`, `json_stringify.c:184`: actual field schemas | `json_metadata.go:shapeJSON`, `emit_objects.go:shapeWithPrivate`: schema/shape identity includes JSON storage and privacy. `internal/regexp/native.go` supplies named-group field schemas, including string/indices unions. |
| `l2:json_stringify.c:147,163`: typed hook result and cleanup | `json_metadata.go:jsonHook`, `json_metadata.c:adamic_json_to_json`; inherited layouts share one thunk per native method, with actual receiver/key and owned-result/throw conventions. |
| `json_stringify.c:105,112,123`: union tags | `json_metadata.c:adamic_json_resolve` and `adamic_json_describe`; decode actual number/boolean boxes, string, intrinsic Map/Set, closure, array and object allocation descriptors; null uses the shared sentinel or proven nullable descriptor. |

Legacy C entry points lacking a descriptor remain usable for non-JSON callers.
They deliberately carry missing metadata: even an empty such array is refused
before a slot read. Compiler-produced arrays use the typed entry points. A generic
C caller must supply its scalar/result descriptor; ownership alone cannot establish
number versus boolean. No data is guessed or silently replaced with `{}`/`null`.

## Agreement and costs

Counted allocation checks compare JSON field kinds with the existing shape storage
kinds and ownership bitmap. Counted reads check concrete reference descriptors
against heap kind. These checks do not infer number versus boolean from their bits;
Node agreement detects the deliberately wrong scalar factory descriptor. Shipping
array allocation/tag writes compile without the counted checks.

On linux-amd64, each shape, array and Map gains two pointers (16 bytes). Object
instances and closures gain no fields. Homogeneous arrays have one element-schema
pointer and a NULL tag-vector pointer, with no tag-vector allocation. Mixed storage
allocates one descriptor pointer per capacity slot. A sort of such storage carries
value/schema pairs in temporary storage; homogeneous sorting retains its ordinary
work representation. Shapes add immutable field/key/schema data. Slot-cache layout,
atomic publication and cached slot indexes are unchanged; JSON-distinct layouts
no longer share a shape merely because their reference bits match.

The four stage 3 root families are tested as representative actual allocations,
including fields beyond a structural view; these are runtime producer tests, not
an admission claim for the full TypeScript compiler roots:

| Root | Node-held allocation family |
| --- | --- |
| `emitter.ts:1138:27` | BuildInfo: nested options, scalar arrays, string arrays and an extra own field. |
| `watchPublic.ts:686:47` | Compiler options: boolean, number, nested plugin objects and an extra own field. |
| `watchPublic.ts:687:80` | Project references: arrays of objects with string/boolean fields. |
| `commandLineParser.ts:2960:31` | Preset/container union: array, object and intrinsic Map, resolved from actual tags. |

`testdata/json-metadata/consumer.c` is a test-only consumer; it neither links nor
calls the library encoder. `expected.mjs` supplies Node 24.19.0's 65 output lines:
the root families, number/boolean/string factories, collection/parallel factories,
copies/writes, optional scalars, explicit absence, key order, private fields,
inherited/reentrant hooks, and nine hook result kinds at root/property/array
positions. Counted and ASan/UBSan builds match Node. The number-to-boolean
`adamic_array_filled_typed` mutant compiles and runs but fails this comparison.
Separate checks cover empty/untyped arrays, refusal before poisoned-slot reads and
a thrown hook's one-call/no-owner handoff.

Focused native controls, affected array/JSON/class/RegExp oracle fixtures, concurrency
controls, the existing array/order mutants and the runtime storage inventory passed.
`go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts` regenerated
`counts.md` with no final diff. Typed parallel ABI detection was updated so this
still pins parallel runs to one thread. No whole-package test invocation was used.
Logs are in `/workspace/scratch/shape-json-contract/`.

Verification machine: AMD EPYC 9V74 under KVM, five visible vCPUs with four CPUs of
cgroup quota. Load samples were 0.00/0.04/0.03 at setup and 0.00/0.00/0.00 after
verification. These are correctness comparisons and allocation-layout costs,
not a performance benchmark.

## Meeting library on main

1. Wire the encoder/emitter to the provider table and admit only the source cases
   whose runtime representations and own-field/initialization proofs are complete.
   This runtime branch deliberately leaves the current lowering refusals intact.
2. Library `cf2cb58b:internal/native/runtime/json_stringify.c:131` accepts only
   object/toJSON descriptors for an `adamic_kind_object` union. Native tuples use
   object slots and correctly describe themselves as `adamic_json_tuple`. Library
   must also accept that existing descriptor kind there to encode tuple-containing
   unions (including entries/index pairs). No shared enum/struct change is needed.
3. Library's test-only provider fixture's positional shape initializers must be
   updated for field kinds and the runtime-owned shape metadata fields when its
   tests meet this runtime. Replace stand-ins with these real providers at that
   integration point. Production UTF-16 escaping, replacers, number spelling and
   encoding/cleanup remain library's code; this consumer exercises the contract
   independently and does not substitute for that integration.

Unproven callable own `toJSON` fields, static own-presence/accessor cases, opaque
runtime carriers without schemas, unsupported heap kinds, BigInt/Symbol, exotics
and cycles remain refused. Sparse mutation and generic union method admission
still need their existing compiler/representation proofs. The runtime does not
invent those proofs from a structural view or an ownership bit.
