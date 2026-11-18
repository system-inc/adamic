# The 14 C03 relations at main 150b7b5e

All coordinates below refer to the unmodified Stage 3 apply of
`150b7b5e8ec45cd88e294dbb728322ec41247737` onto TypeScript 6.0.3
`050880ce59e30b356b686bd3144efe24f875ebc8`. The before-scan site
locations and exact expressions agree with [the checked site map](target-map.json).

Stock tsc accepts these structurally. The tag tests are facts used by
the running compiler, while `Type.flags` is a bitmask and the tests do
not narrow these expressions to the compiler interface `TypeParameter`
in stock tsc. “Construction” includes the symbol-dispatch path back to
the type-parameter cache, with a fresh result from `createTypeParameter`
on a cache miss. “No sufficient evidence” refers to the value actually
crossing the relation.

| Evidence for the related value | G03 | G12 | Total |
| --- | ---: | ---: | ---: |
| `TypeFlags.TypeParameter` on the value or a repeated cached lookup | 4 | 0 | 4 |
| Construction or lookup through the type-parameter cache | 3 | 0 | 3 |
| No sufficient evidence visible | 5 | 2 | 7 |

There are ten argument relations (including an array element in an argument),
three assignments and one return. The 14 relations contain zero casts
to `TypeParameter`.

| Site | Enclosing function | Road and exact expression | Evidence |
| --- | --- | --- | --- |
| G03 S023 `checker.ts:6968:54` | `typeToTypeNodeWorker` | Argument `type` in `typeParameterToName(type, context)` | Tag check `type.flags & TypeFlags.TypeParameter` at `checker.ts:6966`. |
| G03 S024 `checker.ts:7131:128` | `createMappedTypeNodeFromType` | Argument `getConstraintTypeFromMappedType(type)` in `getConstraintOfTypeParameter(getConstraintTypeFromMappedType(type))` | Left operand of the inner `&&` is `getConstraintTypeFromMappedType(type).flags & TypeFlags.TypeParameter` at `checker.ts:7131`. The guard repeats the lookup rather than binding a local. The helper caches the result in `type.constraintType` at `checker.ts:14927–14929`. |
| G03 S025 `checker.ts:7142:90` | `createMappedTypeNodeFromType` | Argument element `getModifiersTypeFromMappedType(target)` in `makeArrayTypeMapper([getTypeParameterFromMappedType(target), getModifiersTypeFromMappedType(target)], [typeParameter, newConstraintParam])` | No sufficient evidence. The homomorphic helper tests `if (typeVariable.flags & TypeFlags.TypeParameter) {` at `checker.ts:20882` after `getActualTypeVariable` at `:20881`. The argument instead comes from the modifiers getter, which can return an instantiated operand or another general type at `checker.ts:14955–14970`. `newConstraintParam` is constructed in the branch and is a different value in the second argument. |
| G03 S026 `checker.ts:9276:95` | `attachSymbolToLeftmostIdentifier` | Argument `type` in `typeParameterToName(type, context)` | Construction/cache path. `type = getDeclaredTypeOfSymbol(sym!)` at `checker.ts:9275`. The taken arm tests `sym!.flags & SymbolFlags.TypeParameter` at `checker.ts:9276` and uses the dispatcher at `checker.ts:13620–13632`. |
| G03 S004 `checker.ts:13244:71` | `getBaseConstructorTypeOfClass` | Argument `baseConstructorType` in `getConstraintFromTypeParameter(baseConstructorType)` | Tag check `if (baseConstructorType.flags & TypeFlags.TypeParameter) {` at `checker.ts:13243`. |
| G03 S005 `checker.ts:13276:58` | `getImplementsTypes` | Argument `implementsType` in `resolvedImplementsTypes.push(implementsType)` | No sufficient evidence. `getTypeFromTypeNode(node)` supplies a `Type`. `!isErrorType(implementsType)` at `checker.ts:13271` excludes errors. The destination is `BaseType[]`, whose `TypeVariable` arm includes `TypeParameter`. |
| G03 S006 `checker.ts:13612:16` | `getDeclaredTypeOfTypeParameter` | Return `links.declaredType \|\| (links.declaredType = createTypeParameter(symbol))` | Construction/cache. Cache miss calls the factory on that line. The hit is typed `Type` in `SymbolLinks.declaredType` at `types.ts:6078` and relies on the type-parameter cache invariant. |
| G03 S007 `checker.ts:16607:33` | `createSignatureTypeMapper` | Argument `getTypeParametersForMapper(signature)!` in `createTypeMapper(getTypeParametersForMapper(signature)!, typeArguments)` | No sufficient evidence for all elements. `signature.typeParameters` starts as `readonly TypeParameter[]`, but `checker.ts:16603` maps a branch through `instantiateType(tp, tp.mapper)`, returning `Type`. The non-null assertion removes absence of the array. |
| G03 S008 `checker.ts:19948:30` | `getInferTypeParameters`, `node.locals.forEach` callback | Assignment RHS `append(result, getDeclaredTypeOfSymbol(symbol))` in `result = append(result, getDeclaredTypeOfSymbol(symbol))` | Construction/cache path. `if (symbol.flags & SymbolFlags.TypeParameter) {` at `checker.ts:19947` selects the type-parameter symbol before lookup. The dispatcher at `checker.ts:13620–13632` leads to the same cache. |
| G03 S015 `checker.ts:23690:67` | `structuredTypeRelatedToWorker` | Argument `source` in `getConstraintOfTypeParameter(source)` | Tag check `if (relation === comparableRelation && sourceFlags & TypeFlags.TypeParameter) {` at `checker.ts:23687`. `let sourceFlags = source.flags;` at `checker.ts:23564`. |
| G03 S016 `checker.ts:23696:71` | `structuredTypeRelatedToWorker` | Argument `constraint` in `getConstraintOfTypeParameter(constraint)` | No sufficient evidence for the whole argument. The loop tests `while (constraint && someType(constraint, c => !!(c.flags & TypeFlags.TypeParameter))) {` at `checker.ts:23692`. For a nonunion this tests the value itself. For a union it tests whether any arm passes, per `checker.ts:28530–28531`. The argument remains the whole constraint. |
| G12 S017 `checker.ts:25216:31` | `getTypeOfPropertyInBaseClass` | Assignment initializer `classType && getBaseTypes(classType)[0]` | No TypeParameter evidence. The short circuit tests presence of the class. The lookup obtains a `BaseType` union element. The finding compares its `InterfaceType` view with the union’s `TypeParameter` arm. |
| G12 S020 `checker.ts:31813:31` | `checkSuperExpression` | Assignment initializer `classType && getBaseTypes(classType)[0]` | No TypeParameter evidence. The preceding `as InterfaceType` is on the class-symbol lookup at `checker.ts:31812`. The base lookup is the same `BaseType` union road as S017. |
| G03 S021 `checker.ts:35806:48` | `instantiateSignatureInContextOf` | Argument `getTypeParametersForMapper(signature)!` in `createInferenceContext(getTypeParametersForMapper(signature)!, signature, InferenceFlags.None, compareTypes)` | No sufficient evidence for the mapped elements, for the same reason as S007. |

The constructor at `checker.ts:5575–5576` returns
`createTypeWithSymbol(TypeFlags.TypeParameter, symbol!) as TypeParameter`.
The symbol-based sites above rely on that construction through dispatch,
rather than holding the factory result directly. Five sites visibly test
`TypeFlags.TypeParameter`. Four tests apply to the related value, and S016
has the existential union case described above.

## Optional members declared by TypeParameter and absent from Type

TypeScript 6.0.3 checker symbol lookup confirms all six on the composed
tree. The slots inherited from `InstantiableType` are already present on
`Type` in this apply.

| Declaration | Location |
| --- | --- |
| `constraint?: Type` | `src/compiler/types.ts:6897` |
| `default?: Type` | `src/compiler/types.ts:6899` |
| `target?: TypeParameter` | `src/compiler/types.ts:6901` |
| `mapper?: TypeMapper` | `src/compiler/types.ts:6903` |
| `isThisType?: boolean` | `src/compiler/types.ts:6905` |
| `resolvedDefaultType?: Type` | `src/compiler/types.ts:6907` |
