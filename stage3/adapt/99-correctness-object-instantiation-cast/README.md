# Match the object instantiation cast to its parameter

In TypeScript 6.0.3, `instantiateTypeWorker` passes
`type as TypeReference | AnonymousType | MappedType` to
`getObjectTypeInstantiation`. The adaptation changes the cast to
`type as AnonymousType | DeferredTypeReference`, which is the parameter's
declared type.

The preceding branch returns for every reference without a `node`. A
`TypeReference` that reaches the call therefore carries a `node` and is a
`DeferredTypeReference`. Both `MappedType` and `InstantiationExpressionType`
extend `AnonymousType`.

The change is type-only. The JavaScript emitted from the before and after
`checker.ts` snapshots is byte-identical, and public declaration emit is
unchanged. The compiler still reports a NotYet at the cast's checked view,
"checked view field target of type AnonymousType | GenericType | undefined".
That checked view is outside the scope of this adaptation. Loom job J-7sw6btk
on main counted 2,565 sites. With the cast applied, J-mmk96mc counted 2,564
because the Refused at the call went away.

Run the snapshot checks from the repository root with the applied tree as
`TREE`.

    NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-object-instantiation-cast/test.cjs TREE
