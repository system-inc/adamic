# Readonly inferred bindings, adaptation 99-readonly-inferred-bindings

Adamic, task #r11wppy, the ruling on class 1 and class 4. This is type-only:
the emitted JavaScript does not change. Every binding is local to its function
and no annotated function is exported, so the published declaration files do
not change, this directory carries no `api-sanctions.json`, and
`stage3/lane/sanctioned-api.json` is not regenerated.

## Why

An operand of `||`, `??` or `?:` is judged where the expression lands. When it
lands in an unannotated binding, or in an unannotated function's return, the
type tsc infers there is the mutable one, so tsc's shared `emptyArray`, typed
`never[]`, lands in a writable `T[]`, a view that could write a `T` where never
is read. Adamic ruled that these bindings and returns get `readonly` as their
annotation: the slot becomes a read-only view and nothing else moves.

## The edit

`adapt.cjs` inserts `: <annotation>` into 16 slots in `src/compiler`, 12 local
bindings and 4 returns, for example:

```ts
const realMembers: readonly Symbol[] = locationMap.get("real") || emptyArray;
function resolveBaseTypesOfClass(type: InterfaceType): readonly BaseType[] {
```

It parses each file with stock TypeScript 6.0.3 and finds each slot by name.
A binding is found by its keyword and name, with the named function as its
nearest enclosing function, and the annotation goes right after the binding's
name. A return is found by its function's name, the only declaration of that
name in the file, local to the function the table names or a module function
with no modifiers, and the annotation goes right after the parameter list's
`)`. It refuses before any write unless every slot has no annotation and its
statement or function text starts with upstream's first line, or already
carries exactly its annotation, which it leaves, so a second run changes
nothing. No source coordinate is a rewrite address.

The directory sorts after every 99-correctness adaptation and before
99-wasm-crypto-import. Two of them pin the whole text of a function this edit
touches: 99-correctness-axolotl-r pins checker.ts's `getTypeArguments`, which
holds `typeArguments`, and 99-correctness-binding-views pins program.ts's
`createProgram`, which holds `getOptionsDiagnosticsOfConfigFile`, through
95-checker-axolotl-h's `diagnostic-owners.json`. Both must still read those
functions as they expect when they run. No other adaptation reads or pins any
of the 16 slots or a declaration that contains one.

## Certificate

Certified by stock tsc over tsc's whole `src`: a binding that's written
through, pushed, spliced, index-written or stored into a mutable slot, fails
tsc at the write. The 16 type-check clean under stock tsc 6.0.3 with
adaptations 95-readonly-empty-array-slots and 96-readonly-declared-members
(J-7dywgkr). One candidate did fail, #67 in the classing table, and it is left out: checker.ts's
`let typeParameters` in `getObjectTypeInstantiation`, stored into the mutable
`NodeLinks.outerTypeParameters`. It stays unannotated, and this adaptation
names no slot for it.

## The slots

Sites are on apply's tree, as the classing run measured them. The five sites of
`resolveBaseTypesOfClass` are its five returns, which share the one return
annotation.

| Slot | File | Annotation | Sites it holds |
|---|---|---|---|
| inferPreference, `const specifiers` | utilities.ts | `readonly StringLiteralLike[]` | utilities.ts:10099:13 |
| serializeModule, `const realMembers` | checker.ts | `readonly Symbol[]` | checker.ts:10037:64 |
| serializeModule, `const mergedMembers` | checker.ts | `readonly Symbol[]` | checker.ts:10038:68 |
| serializeAsNamespaceDeclaration, `const localProps` | checker.ts | `readonly Symbol[]` | checker.ts:10194:73 |
| serializeAsClass, `const privateProperties` | checker.ts | `readonly ClassElement[]` | checker.ts:10335:21 |
| resolveBaseTypesOfClass, return | checker.ts | `readonly BaseType[]` | checker.ts:13324:20, 13348:24, 13354:20, 13361:20, 13365:20 |
| resolveAnonymousTypeMembers, `let constructSignatures` | checker.ts | `readonly Signature[]` | checker.ts:14670:132 |
| resolveReverseMappedTypeMembers, `const indexInfos` | checker.ts | `readonly IndexInfo[]` | checker.ts:14720:203 |
| getTypeArguments, `const typeArguments` | checker.ts | `readonly Type[] \| undefined` | checker.ts:16967:43 |
| getIntendedTypeFromJSDocTypeReference, `const indexInfo` | checker.ts | `readonly IndexInfo[]` | checker.ts:17357:157 |
| getExactOptionalUnassignableProperties, return | checker.ts | `readonly Symbol[]` | checker.ts:24928:64 |
| createEmptyObjectTypeFromStringLiteral, `const indexInfos` | checker.ts | `readonly IndexInfo[]` | checker.ts:26445:131 |
| addImplementationSuccessElaboration, `const failedSignatureDeclarations` | checker.ts | `readonly Declaration[]` | checker.ts:36778:93 |
| getScriptTransformers, return | transformer.ts | `readonly TransformerFactory<SourceFile \| Bundle>[]` | transformer.ts:128:26 |
| getOptionsDiagnosticsOfConfigFile, return | program.ts | `readonly DiagnosticWithLocation[]` | program.ts:3277:41 |
| getSemanticDiagnosticsOfNextAffectedFile, `let result` | builder.ts | `readonly Diagnostic[] \| undefined` | builder.ts:2192:41 |

`getScriptTransformers` is a module function of transformer.ts that nothing
exports; `getTransformers` stores its result into
`EmitTransformers.scriptTransformers`, already readonly. The other three
returns are local to `createTypeChecker` or `createProgram`.

## Proof

Sealed, by scan-all's instrumented probe over sys.ts's closure in apply's tree,
with stock TypeScript 6.0.3 (J-wq7r8xc):

| Measurement | Before | After |
|---|---:|---:|
| Distinct stops, every rule | 3,319 | 3,299 |
| never[] stops, distinct | 89 | 70 |
| Stops new against before | | 0 |
| Stock TypeScript errors over src/compiler | 1 | 1, the same |

"Before" is apply's tree with the 16 annotations put back exactly as upstream
writes them. Twenty stops go, and none is new: checker.ts:10037:64, 10038:68,
10194:73, 10335:21, 13324:20, 13348:24, 13354:20, 13361:20, 13365:20, 14670:132,
14720:203, 16967:43, 17357:157, 24928:64, 26445:131 and 36778:93; builder.ts:2192:41;
program.ts:3272:13 and 3277:41; and transformer.ts:128:26. One of them,
program.ts:3272:13, is a DiagnosticWithLocation[] seen as Diagnostic[], cleared by
the same annotation. They clear on the annotations alone, without the separate
compiler rule on where an operand lands: an annotated binding is already a
declared slot to the compiler.

The `specifiers` binding at utilities.ts:10099 holds no stop this scan reaches,
so the mutant that takes its annotation back out moves nothing. It's kept
because stock TypeScript certifies it, and its site is listed in the scan at
03c140ec. The adapted tree's product test and stage3/lane pass.

## Re-running

`node adapt.cjs <tree>` uses stock typescript@6.0.3 through NODE_PATH or
`CENSUS_TYPESCRIPT`, as apply runs every adaptation.
