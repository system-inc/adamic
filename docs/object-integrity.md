# Object integrity on proven shapes (roadmap step 22)

`Object.seal`, `Object.freeze`, and `Object.preventExtensions` preserve receiver identity.
`Object.isSealed`, `Object.isFrozen`, and `Object.isExtensible` observe shared receiver state.
The JavaScript backend calls the original Object method. Native and WASI use the same runtime
implementation, extending the existing V8 SetIntegrityLevel/TestIntegrityLevel port recorded in
THIRD_PARTY_NOTICES.md.

Plain literal shapes have configurable, writable data properties. Constructor-proven Date,
Map, and Set instances have no own properties: preventing extensions makes them both sealed
and frozen. Their internal slots are not property descriptors. Map and Set entries remain
mutable after any of these operations. Collection integrity state is atomic and belongs to the
receiver, so aliases see the same state.

A constructor-proven RegExp has one own property, the non-configurable, writable `lastIndex`.
Preventing extensions already seals it, but does not freeze it. Sealing preserves `lastIndex`
writes and matching. Freezing RegExp remains NotYet until writes by both user code and the
matching engine can produce catchable Node TypeErrors.

The lowering proof follows constructors, immutable aliases, and integrity calls that return
the original receiver. Structural interfaces, parameters, arrays, and unimplemented host
constructors do not establish that proof. `new Object()` with no arguments is a fresh empty
plain object. An untyped fresh empty Map/Set is admitted only as the direct argument of an
integrity call whose result is discarded. Its empty contents prove a vacuous representation;
no any/unknown collection is allowed to escape into a binding or subsequent operation.

Errors remain refused for integrity mutation: native synthetic fields do not yet model
Node's own stack accessor and optionally present message property. Boxed primitives, buffers,
typed arrays, weak collections, promises, and dynamically constructed functions need their own
constructor and descriptor representations. This unit does not change descriptor functions.

Fixtures `library_object_seal_shapes.a`, `library_object_seal_collections.a`, and
`library_object_seal_regexp.a` compare original source on Node 24.19.0 with both backends and
WASI. Four clean-exit mutants are caught solely by stdout comparison: replacing preventExtensions
with seal, replacing the empty-own-property frozen query with extensibility, skipping collection
sealing, and confusing sealed with frozen on RegExp. Allocation counts are recorded in
internal/oracle/counts.md. Test262 comparisons pin `TZ=UTC`, with adaptation disabled.
