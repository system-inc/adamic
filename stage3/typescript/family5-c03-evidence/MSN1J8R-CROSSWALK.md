# C03 compared with #msn1j8r sub-shape 2-o

**Zero of the 14 C03 relations match the #msn1j8r cast sites.** The
crosswalk uses enclosing function and expression. Coordinates for C03
come from the applied main at `150b7b5e`. The Apple log uses `03c140ec`
and presents 24 distinct numbered source excerpts for the 29-site
inventory. Some excerpts contain multiple casts. The marked source
lines and their context identify explicit assertion operands.
[Apple site log at dea397b1](https://github.com/system-inc/adamic/blob/dea397b17195c75552e5a1a3dd0bb952ff334601/evidence-msn1j8r/msn1j8r-sites-03c140ec.txt)

| C03 site | Enclosing function and expression | In #msn1j8r? |
| --- | --- | --- |
| G03 S023 `6968:54` | `typeToTypeNodeWorker`, `type` in `typeParameterToName(type, context)` | No. Argument. Apple’s same-function entries show explicit casts in the earlier infer branch. |
| G03 S024 `7131:128` | `createMappedTypeNodeFromType`, `getConstraintTypeFromMappedType(type)` passed to `getConstraintOfTypeParameter` | No. Argument. Apple’s same-function cast uses `getTypeFromTypeNode(context, (type.declaration.typeParameter.constraint! as TypeOperatorNode).type) as TypeParameter`. |
| G03 S025 `7142:90` | `createMappedTypeNodeFromType`, `getModifiersTypeFromMappedType(target)` in the source array of `makeArrayTypeMapper` | No. Argument element. The same-function Apple cast is a different expression. |
| G03 S026 `9276:95` | `attachSymbolToLeftmostIdentifier`, `type` passed to `typeParameterToName` | No. Argument. The cast in the ternary’s other arm is to `Identifier`. |
| G03 S004 `13244:71` | `getBaseConstructorTypeOfClass`, `baseConstructorType` passed to `getConstraintFromTypeParameter` | No. Argument. |
| G03 S005 `13276:58` | `getImplementsTypes`, `implementsType` passed to `resolvedImplementsTypes.push` | No. Argument. |
| G03 S006 `13612:16` | `getDeclaredTypeOfTypeParameter`, `links.declaredType \|\| (links.declaredType = createTypeParameter(symbol))` | No. Return. The factory’s explicit `as TypeParameter` is a separate expression in `createTypeParameter`, and is in Apple’s log at `5571`. |
| G03 S007 `16607:33` | `createSignatureTypeMapper`, `getTypeParametersForMapper(signature)!` passed to `createTypeMapper` | No. Argument. `!` is a non-null assertion on the array. |
| G03 S008 `19948:30` | `getInferTypeParameters`, `append(result, getDeclaredTypeOfSymbol(symbol))` assigned to `result` | No. Assignment. Apple’s excerpt starting at `19950` includes this code as context. Its casts are `root.checkType as TypeParameter` in the following `isDistributionDependent`. |
| G03 S015 `23690:67` | `structuredTypeRelatedToWorker`, `source` passed to `getConstraintOfTypeParameter` | No. Argument. Apple’s same-function entry uses `source as TypeVariable` in `getConstraintOfType`. |
| G03 S016 `23696:71` | `structuredTypeRelatedToWorker`, `constraint` passed to `getConstraintOfTypeParameter` | No. Argument. The same-function Apple cast is a different expression. |
| G12 S017 `25216:31` | `getTypeOfPropertyInBaseClass`, `classType && getBaseTypes(classType)[0]` | No. Assignment initializer. It reports the `InterfaceType` source arm, so the source wording differs as well. |
| G12 S020 `31813:31` | `checkSuperExpression`, `classType && getBaseTypes(classType)[0]` | No. Assignment initializer. It also reports the `InterfaceType` source arm. |
| G03 S021 `35806:48` | `instantiateSignatureInContextOf`, `getTypeParametersForMapper(signature)!` passed to `createInferenceContext` | No. Argument. |

## Status of the fall-through at 150b7b5e

Sub-shape 2-o’s fall-through has **not landed at this pin**. The
assignable-cast branch of `castProof` calls `provenRelation` at line
237. At lines 242–246 it returns the proof error early when the types are
not all class instances, or when the optional-relation proof fails.
The separate interface-view attempt is lower in the function at line
300. The baseline refusal scan confirms the early result on the distinct
Apple cast in `createTypeParameter` at `checker.ts:5576:16`.
[Main castProof](https://github.com/system-inc/adamic/blob/150b7b5e8ec45cd88e294dbb728322ec41247737/internal/lower/cast_proof.go#L233)

These 14 sites are ordinary contextual relations. `optionalInSlot`
uses the slot type or implied target, then `optionalWidened` follows
container elements and union arms and reports the first absent optional
property. `optionalWideningRefusal` emits the family 5 wording
“optional property constraint in TypeParameter absent from structural
source …, which can hide fields.” `provenRelation` emits the different
“an unproven relation … optional field constraint has no proven
compatible presence/type” wording used on Apple’s cast road. Changing
the early return in `castProof` alone would leave all 14 contextual
relations on the family 5 road.
[Contextual rule](https://github.com/system-inc/adamic/blob/150b7b5e8ec45cd88e294dbb728322ec41247737/internal/lower/optional_widening.go#L363)
[Cast relation wording](https://github.com/system-inc/adamic/blob/150b7b5e8ec45cd88e294dbb728322ec41247737/internal/lower/proven_relations.go#L11)
