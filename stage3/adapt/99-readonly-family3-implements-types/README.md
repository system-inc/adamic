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

## Loom results

The census ran with and without the adapter at main c02b3023, both sides in
one job on one box (J-fm0jqg4): 1,798 stops without, 1,797 with. An earlier
pair at 42edc2ed, J-khemp94 without and J-h43ax3b with, read the same three
lines. The gate passed in J-w4na21h. The restoring check copied the original source
back.

- Cleared target at `checker.ts:13264:51`: `never[] seen as BaseType[], which can write BaseType where never is read`.
- Moved on one line at `checker.ts:13276`:
  `refuseOptionalWidening` ('optional property constraint in TypeParameter absent from structural source Type').
  Without the adapter it stops at column 58. With it, the stop is at column 84.
  It's the same token both times, `implementsType`, stopping for the same
  family 5 reason: the adapter rewrites the writer on this line, which starts
  at column 29, and the rewrite pushes the token 26 columns right. So the
  line still stops and nothing new appears. This is a moved stop, not a clear.
