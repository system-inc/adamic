# Speculative IR field policies

`ir_fields.go` is the explicit exported-field inventory. Each entry names the IR
owner and field, exact declared type, action, and reason. It is never regenerated
by `make_overlay.py`. `loadGenerator` compares it with the real IR declarations
before generating an overlay: unnamed fields and changed types fail with the
field name and source location. Copied policies are also checked recursively
against the real underlying types, so a mutable aggregate cannot be copied by
assignment. Embedded fields require a new explicit decision too.

The general speculative copier receives its mutable field indexes from this
checked inventory. The Function slice optimization receives a generated typed
copier from that same inventory. There is no separate runtime Function field
allowlist. A future field must get a policy before either path builds.

For every Function field:

| Fields | Action | Why |
| --- | --- | --- |
| SourceLength, ProgramRegion, GraphClosure, CheckedUnionNarrow, Name, MethodName, Position, ArgumentsCount, ReadsArguments, ForwardsArguments, Receiver, RestElement, Returns, Async, AsyncSite, AsyncReturns, Generator, Yields, ResumeMayThrow, Closure, NestedParent, ForwardedNestedParent, NestedFrame, FrameIdentity, MayThrow | Copied by value | Scalars, enums, strings, or the scalar-only SourcePosition aggregate. They contain no mutable descendants. |
| Boundary | Deep-copied | Own its parameter slice, return schema pointer, and nested schema graph. The shared `seen` map preserves aliases between schema references within the snapshot. |
| Source | Deep-copied | Own its Declaration ancestry slice while preserving the stable identity strings. |
| Parameters, OptionalParameters, Body, Environment, FrameEnvironment, ReferenceParents | Deep-copied | These are lowering-owned slices, maps, or dynamic statement graphs. Failed lowering must not mutate the checkpoint. |

Private fields retain the single `irPrivateFields` table in `main.go`:

| Field | Action | Why |
| --- | --- | --- |
| Program.argumentFacts | Started empty; require the source to be nil | Backend-only derived facts; census lowering never populates them. A populated value remains an explicit invalid-state refusal, not an unnamed-field schema failure. |
| Program.methodTables | Started empty | The IR rebuilds the cache from copied method declarations. |
| BoundaryStringTable.indexes | Started empty | The IR rebuilds it from copied Values. |
| GeneratedType.identity | Share the owning GeneratedType pointer | Immutable owner identity is compared by pointer. Copying its facade would change identity. |

All other exported IR fields are individually named in `ir_fields.go`, with
copied scalar values or deep copies of mutable graphs. Deep traversal uses the
same private-cache and shared-pointer policies. No mutable private cache is
retained by a shallow struct assignment.
