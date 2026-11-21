# Readonly implements types local

Adamic #k8ed1f2, Nov 18 slot-by-slot ruling and Nov 21 delegation. In
`getImplementsTypes`, give the local `resolvedImplementsTypes` a readonly
array type and replace its one in-place push with a fresh-array assignment.
The stock TypeScript 6.0.3 AST identifies the owner, binding and guard. Both
edit targets and the equal arm must have their expected text before any
write. A second run leaves the file unchanged.

With ordinary array intrinsics, the first successful implements check
installs a fresh dense array and subsequent checks append the same type
objects in order. The array remains local until the final return. Its
identity before that return never escapes. The empty return retains the
sentinel, and the declaration and return type of the enclosing function stay
as they were. Repeated appends now copy the accumulated elements, so the
allocation and complexity of a long implements list differ. A modified
array iterator or push intrinsic is outside this equivalence premise.

## Ordering and scope

Later adapters are `99-readonly-family3-views`,
`99-readonly-family3-views-2`, `99-readonly-inferred-bindings`,
`99-readonly-script-target-entries`, `99-truthful-generated-name-emit-node`,
`99-truthful-module-export-bindings` and `99-wasm-crypto-import`. None pins
`getImplementsTypes`, its local, or the enclosing checker owner as a whole.
The inferred-bindings adapter uses `createTypeChecker` as a locality check
for two other declarations. Its hash-checked owners are separate functions
and one interface. Earlier adaptation `95-readonly-empty-array-slots`
already sets this function's return type to `readonly BaseType[]`.

The final return is the local's only escape. The private function keeps its
existing return type. This adapter has no public API sanction. The only
intended JavaScript delta is the writer line. `InterfaceType.resolvedBaseTypes`
and other sentinel users are outside this adapter.
