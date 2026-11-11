# Filter narrowing-overload census

Task `z5d0jye`. Report only. No compiler, adaptation, lane, or expected-result change is proposed.

**Compiler census and full sharded runtime count: complete.** Noncompiler production loading remains unavailable, as separated below. The source reading is **12 safe, 15 unknown** for the 27 compiler sites; it does not clear a copying adaptation.

Base: `efe887763f26e96653546fae904d57539139f7ae` (fetched main). TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`. Exact dependency/tool pins are in [environment.json](evidence/environment.json).

## Count and limits

Production `load.Load` accepted all **79 compiler roots together**, with its options and diagnostics unchanged. The driver asked `GetResolvedSignature` for every call expression in those roots, then selected declarations named `filter` owned by `src/compiler/core.ts`. It found **113 calls**, of which **27 resolve to a type-predicate overload**. This is a call-site census, not a count of the 18 units whose first stop mentions filter.

| Scope | Narrowing sites | Safe | Identity-compared | Written-while-live | Unknown | Signature evidence |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Compiler | 27 | 12 | 0 confirmed | 0 confirmed | 15 | Production checker |
| Services | 10 | 0 | 0 confirmed | 0 confirmed | 10 | Stock 6.0.3 supplement |
| Harness | 0 supplementary | | | | | Production loading rejected |
| Server | 0 supplementary | | | | | Production loading rejected |

**Unknown is not safe.** Zero confirmed identity/write findings does not establish their absence in unresolved paths. These readings do not clear a copying adaptation.

The noncompiler regions fail production loading. Their separate list comes from an exhaustive stock-checker call traversal with declaration verification, not an admitted production program. All 27 compiler locations agree between that scan and the exhaustive production traversal. The secondary scan saw 30 services, 3 harness and 7 server calls of core filter in total; only the 10 services calls selected a narrowing overload. The scope rejection receipts retain every diagnostic.

| Region | Roots | Production diagnostics |
| --- | ---: | ---: |
| compiler | 79 | 0 |
| services | 168 | 2774 |
| server | 15 | 3222 |
| harness | 38 | 3647 |

The mutable, readonly and optional overloads are all included. The eight exported overload declarations occupy core.ts:261 through 275; the implementation begins at 277.

| Narrowing overload | Declared array parameter | Sites | Safe | Unknown |
| --- | --- | ---: | ---: | ---: |
| 1 | `T[]` | 5 | 3 | 2 |
| 3 | `readonly T[]` | 9 | 1 | 8 |
| 5 | `T[] or undefined` | 5 | 4 | 1 |
| 7 | `readonly T[] or undefined` | 8 | 4 | 4 |

## Identity comparison excluded from the narrowing set

`src/compiler/utilities.ts:4686:31` calls `filter(jsDoc.tags, tag => ownsJSDocTag(hostNode, tag))`; the next line compares `jsDoc.tags === ownedTags`. The production checker resolves this call to core.ts:275, the boolean readonly-or-undefined overload. It is not C24: C24 is the other branch, using `isJSDocOverloadTag`. Copying the shared implementation for every overload would affect this existing boolean-overload identity test. A narrowing-call inventory alone is not clearance for that broader change.

## Runtime measurement

The complete adapted control tree and a separate counting copy were built with the gate product tasks `hereby local` and `hereby tests`. The private copy changes only 37 callee expressions in 15 source files. A typed curry retains the original overload set; its runtime wrapper evaluates the argument and predicate once, calls the unchanged filter, and returns the unchanged result. Both typescript.d.ts and typescript.internal.d.ts are byte-identical to the control. [instrumentation-proof.json](evidence/instrumentation-proof.json) and [instrumentation.patch.gz](evidence/instrumentation.patch.gz) retain that check.

The runner uses the unchanged `stage3/lane/measure_shards.py` and `STAGE3_ADAPTED_TREE` product path, on private views. All manifest units are represented; each selected attempt uses its exact manifest command and unchanged gate deadline. The initial concurrent attempt timed out and is excluded. Final counts select exactly one complete, attributed observer-v2 attempt for each of the **398 shards**: 331 from the serial main pass and 67 from the durable-observer repeats. Earlier observer versions and the cancelled concurrent run are excluded, not added to these totals.

`calls` counts attempts; `completed` counts normal returns; `identity_equal` counts `result === argument`, including undefined; `returned_argument_array` requires an actual Array. The latter avoids treating an undefined argument/result as an array alias. Process-start receipts, periodic snapshots, normal exit and default-signal flushes make loss visible. Uncatchable termination leaves an incomplete receipt and cannot be treated as zero. Controls check copying, identity, empty input, undefined input, throwing predicates, default SIGTERM, application SIGTERM handlers and SIGKILL loss detection.

**Later mutation was not measured.** No freeze or function-exit rewrite was used. Retaining every identity-returned array and element snapshot for all calls, or wrapping enclosing functions and escaped lifetimes, was not treated as a cheap semantics-preserving extension of this census. No claim of zero later mutations follows from the counters.

### Whole-lane result

All **398 selected shard gate checks passed**. Their complete test identities were independently checked against the manifest. The selected union has **106,366 passing, 1 expected Public API baseline failure, 0 pending**. The gate accepted that single sanctioned baseline failure; there were no unexpected failures. All **797 selected process receipts** are terminal observer-v2 receipts. Counts include per-shard setup and describe this sharded execution, not a hypothetical single-process run.

| Scope | Calls | Normal returns | result === argument | Returned argument Array | Undefined argument |
| --- | ---: | ---: | ---: | ---: | ---: |
| compiler | 1,163,183 | 1,163,183 | 1,085,224 | 1,069,796 | 15,428 |
| services | 1,017 | 1,017 | 224 | 160 | 64 |

### Per-site runtime counts

| Site | Source class | Calls | Normal returns | result === argument | Argument Array returned | Undefined argument |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| C01 | unknown | 80 | 80 | 0 | 0 | 0 |
| C02 | safe | 16 | 16 | 13 | 13 | 0 |
| C03 | safe | 0 | 0 | 0 | 0 | 0 |
| C04 | safe | 0 | 0 | 0 | 0 | 0 |
| C05 | safe | 1,840 | 1,840 | 1,621 | 1,621 | 0 |
| C06 | unknown | 148 | 148 | 0 | 0 | 0 |
| C07 | safe | 238,252 | 238,252 | 238,142 | 232,451 | 5,691 |
| C08 | safe | 2,516 | 2,516 | 371 | 371 | 0 |
| C09 | unknown | 816,898 | 816,898 | 814,915 | 814,915 | 0 |
| C10 | unknown | 35,872 | 35,872 | 1,873 | 1,873 | 0 |
| C11 | unknown | 5,968 | 5,968 | 5,967 | 5,967 | 0 |
| C12 | unknown | 2,245 | 2,245 | 30 | 30 | 0 |
| C13 | unknown | 9,902 | 9,902 | 9,900 | 1,197 | 8,703 |
| C14 | safe | 484 | 484 | 484 | 156 | 328 |
| C15 | unknown | 422 | 422 | 205 | 63 | 142 |
| C16 | unknown | 21,883 | 21,883 | 104 | 104 | 0 |
| C17 | unknown | 1,892 | 1,892 | 1,259 | 1,259 | 0 |
| C18 | safe | 88 | 88 | 85 | 85 | 0 |
| C19 | safe | 88 | 88 | 0 | 0 | 0 |
| C20 | safe | 318 | 318 | 314 | 266 | 48 |
| C21 | safe | 318 | 318 | 48 | 0 | 48 |
| C22 | unknown | 3,417 | 3,417 | 1,982 | 1,982 | 0 |
| C23 | unknown | 11,433 | 11,433 | 1,771 | 1,771 | 0 |
| C24 | safe | 2,124 | 2,124 | 628 | 160 | 468 |
| C25 | unknown | 2,893 | 2,893 | 2,225 | 2,225 | 0 |
| C26 | unknown | 4,080 | 4,080 | 3,281 | 3,281 | 0 |
| C27 | unknown | 6 | 6 | 6 | 6 | 0 |
| S01 | unknown | 64 | 64 | 0 | 0 | 0 |
| S02 | unknown | 69 | 69 | 69 | 5 | 64 |
| S03 | unknown | 34 | 34 | 1 | 1 | 0 |
| S04 | unknown | 53 | 53 | 53 | 53 | 0 |
| S05 | unknown | 3 | 3 | 0 | 0 | 0 |
| S06 | unknown | 33 | 33 | 5 | 5 | 0 |
| S07 | unknown | 27 | 27 | 4 | 4 | 0 |
| S08 | unknown | 37 | 37 | 6 | 6 | 0 |
| S09 | unknown | 30 | 30 | 4 | 4 | 0 |
| S10 | unknown | 667 | 667 | 82 | 82 | 0 |

Zero executed calls: C03, C04. A zero is not an unreachability or safety proof.

The site JSON carries these counts beside each exact location and reading. [runtime-summary.json](evidence/runtime-summary.json) records totals and selection provenance. [runtime-selected.json.gz](evidence/runtime-selected.json.gz) retains all selected counters and shard reports; [runtime-observations.json.gz](evidence/runtime-observations.json.gz) retains the complete test-identity observations. [runtime-attempts.json.gz](evidence/runtime-attempts.json.gz) preserves excluded and replaced attempts separately.

The exact adapted overload declarations and implementation are in [filter-overloads.txt](evidence/filter-overloads.txt).

## Source readings

The safe classification uses the actual last read/copy boundary of the narrowed array. Writes to a separate destination or to fields on its element nodes are not writes to that array. For unknown sites, the reading names the retained owner and the unclosed call/escape path. Identity checks include equality/inequality operators, Object.is, and Map/Set key use alongside the argument; writes include all requested mutators, index/length stores and indirect writers. No unknown is inferred safe from a readonly annotation or from a runtime zero.

Argument annotations are quoted where present. Inferred locals and temporary expressions have no declared annotation; their checker type is shown instead. JSON retains the declaration locations/text and the exact result context.

### C01: unknown - src/compiler/checker.ts:3138:58

Enclosing function: `isUsedInFunctionOrInstancePropertyWorker::<findAncestor callback>`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
declaration.parent.members
```

Type at call: `NodeArray<ClassElement>`.
Declared annotation(s): `NodeArray<ClassElement>`.

Predicate:
```ts
isClassStaticBlockDeclaration
```

declaration.parent.members -> filter -> staticBlocks -> isPropertyInitializedInStaticBlocks, which performs flow/type queries. The class node owns members and is reachable through declaration. No local identity test or array write is visible; transitive checker/reentrant callback writes were not exhaustively followed.

### C02: safe - src/compiler/checker.ts:6586:39

Enclosing function: `simplifyClassDeclaration`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (production).

Argument:
```ts
symbol.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[] | undefined`.

Predicate:
```ts
isClassLike
```

symbol.declarations -> filter -> classDeclarations -> first element only (originalClassDecl). After that indexed read, the array is never read again; modifier queries and factory updates receive nodes, not this array. No array identity test or intervening write. The symbol owns the argument, but later symbol writes occur after the result last use.

### C03: safe - src/compiler/checker.ts:9417:28

Enclosing function: `getNamesOfDeclaration`. Resolved overload: **1**, `src/compiler/core.ts:261:1` (production).

Argument:
```ts
map(statement.declarationList.declarations, getNameOfDeclaration)
```

Type at call: `(DeclarationName | undefined)[]`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isIdentifierAndNotUndefined
```

map(declarationList.declarations, getNameOfDeclaration) allocates a fresh array -> filter -> getNamesOfDeclaration return -> flatMap(excessExports) -> addRange copies -> map into ExportSpecifier nodes. No original wider array handle survives; flatMap reads and copies the result before later factory calls.

### C04: safe - src/compiler/checker.ts:9419:24

Enclosing function: `getNamesOfDeclaration`. Resolved overload: **1**, `src/compiler/core.ts:261:1` (production).

Argument:
```ts
[getNameOfDeclaration(statement as DeclarationStatement)]
```

Type at call: `(DeclarationName | undefined)[]`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isIdentifierAndNotUndefined
```

fresh one-element array literal -> filter -> getNamesOfDeclaration return -> flatMap(excessExports) -> addRange copy -> map into ExportSpecifier nodes. No other owner of the argument exists, and the caller copies rather than mutates or identity-compares the returned array.

### C05: safe - src/compiler/checker.ts:10370:48

Enclosing function: `serializeAsClass`. Resolved overload: **1**, `src/compiler/core.ts:261:1` (production).

Argument:
```ts
symbol.declarations
```

Type at call: `Declaration[]`.
Declared annotation(s): `Declaration[] | undefined`.

Predicate:
```ts
d => isClassDeclaration(d) || isClassExpression(d)
```

symbol.declarations -> filter -> immediate [0] -> setTextRange location node. Only the first node escapes; the filtered array dies at the index read, before setTextRange/addResult. No array identity comparison, key use or later array write while live.

### C06: unknown - src/compiler/checker.ts:11961:38

Enclosing function: `getTypeForVariableLikeDeclaration`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
declaration.parent.members
```

Type at call: `NodeArray<ClassElement>`.
Declared annotation(s): `NodeArray<ClassElement>`.

Predicate:
```ts
isClassStaticBlockDeclaration
```

declaration.parent.members -> filter -> staticBlocks -> length and getFlowTypeInStaticBlocks. The owning class node remains reachable. No local comparison/write is present, but all checker flow-query aliases and callbacks during that call were not closed.

### C07: safe - src/compiler/checker.ts:16764:27

Enclosing function: `getConstraintDeclaration`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (production).

Argument:
```ts
type.symbol && type.symbol.declarations
```

Type at call: `Declaration[] | undefined`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isTypeParameterDeclaration
```

type.symbol.declarations -> filter -> mapDefined(getEffectiveConstraintOfTypeParameter) -> [0]. mapDefined allocates its result and only invokes the constraint accessor on each element; the array itself is not passed to that accessor. Only a constraint node escapes. No identity/key use or array writer on this path.

### C08: safe - src/compiler/checker.ts:32639:25

Enclosing function: `discriminateContextualTypeByObjectMembers`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
node.properties
```

Type at call: `NodeArray<ObjectLiteralElementLike>`.
Declared annotation(s): `NodeArray<T>`.

Predicate:
```ts
(p): p is PropertyAssignment | ShorthandPropertyAssignment => {
                            if (!p.symbol) {
                                return false;
                            }
                            if (p.kind === SyntaxKind.PropertyAssignment) {
                                return isPossiblyDiscriminantValue(p.initializer) && isDiscriminantProperty(contextualType, p.symbol.escapedName);
                            }
                            if (p.kind === SyntaxKind.ShorthandPropertyAssignment) {
                                return isDiscriminantProperty(contextualType, p.symbol.escapedName);
                            }
                            return false;
                        }
```

node.properties -> filter -> map into pairs containing a type-query closure and escapedName -> concatenate -> discriminateTypeByDiscriminableItems. The map callback constructs pairs without querying types; the later closures capture individual properties, not the array. The filtered array is dead before those queries run. No identity/key use or array write before that copy boundary.

### C09: unknown - src/compiler/checker.ts:46866:33

Enclosing function: `checkUnmatchedJSDocParameters`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
getJSDocTags(node)
```

Type at call: `readonly JSDocTag[]`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isJSDocParameterTag
```

getJSDocTags(node), the reachable node.jsDoc.jsDocCache (or the shared emptyArray fallback) -> filter -> jsdocParameters. The result is retained across containsArgumentsReference, type lookup and error production, then traversed. No local identity test or write; the complete cache/reentrant checker write graph was not followed.

### C10: unknown - src/compiler/checker.ts:47509:16

Enclosing function: `getClassOrInterfaceDeclarationsOfSymbol`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (production).

Argument:
```ts
symbol.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[] | undefined`.

Predicate:
```ts
(d: Declaration): d is ClassDeclaration | InterfaceDeclaration => d.kind === SyntaxKind.ClassDeclaration || d.kind === SyntaxKind.InterfaceDeclaration
```

symbol.declarations -> filter -> getClassOrInterfaceDeclarationsOfSymbol return -> areTypeParametersIdentical loops declarations and performs type/constraint queries. Symbol ownership permits aliases outside this helper. Local reads show no comparison/write, but all transitive writes while this returned array is live were not closed.

### C11: unknown - src/compiler/commandLineParser.ts:2550:16

Enclosing function: `convertArrayLiteralExpressionToJson`. Resolved overload: **1**, `src/compiler/core.ts:261:1` (production).

Argument:
```ts
elements.map(element => convertPropertyValueToJson(element, elementOption))
```

Type at call: `(AdamicJsonRecoveryValue | undefined)[]`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
v => v !== undefined
```

elements.map allocates a fresh (AdamicJsonRecoveryValue | undefined)[] -> filter -> convertArrayLiteralExpressionToJson return -> recursive JSON object/array fields -> convertToJson/convertToObject return. convertObjectLiteralExpressionToJson stores value in its result and passes that value to jsonConversionNotifier.onPropertySet. The wider temporary is not retained separately, but the result escapes through JSON values and notifier callbacks. Under the requested rule covering returned aliases, later writes or identity uses through those holders were not closed.

### C12: unknown - src/compiler/commandLineParser.ts:3157:32

Enclosing function: `getConfigFileSpecs`. Resolved overload: **1**, `src/compiler/core.ts:261:1` (production).

Argument:
```ts
[outDir, declarationDir]
```

Type at call: `(string | undefined)[]`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
(d): d is string => !!d
```

fresh [outDir, declarationDir] literal -> filter -> excludeSpecs -> validateSpecs and config-dir substitution -> excludeSpecs/validatedExcludeSpecs fields of ConfigFileSpecs -> sourceFile.configFileSpecs. The literal has no separate surviving wider handle, but the result can stay live through the config fields while getFileNames invokes host callbacks and subsequent watchers retain the source file. Later writes and identity/key uses through returned/stored aliases were not closed.

### C13: unknown - src/compiler/commandLineParser.ts:3188:54

Enclosing function: `getConfigFileSpecs`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
filesSpecs
```

Type at call: `readonly string[] | undefined`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isString
```

raw files property -> getSpecsFromRaw/toPropValue -> filesSpecs -> filter -> validatedFilesSpecBeforeSubstitution. Both original and narrowed arrays are stored on returned ConfigFileSpecs; substitution may also return an alias. raw belongs to the config caller and remains reachable. Later config consumers and caller mutations while the returned config lives were not closed.

### C14: safe - src/compiler/factory/utilities.ts:761:12

Enclosing function: `getImportedHelpers`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (production).

Argument:
```ts
getEmitHelpers(sourceFile)
```

Type at call: `EmitHelper[] | undefined`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
helper => !helper.scoped
```

sourceFile emitNode.helpers via getEmitHelpers -> filter -> helpers in createExternalHelpersImportDeclarationIfNeeded. ESM path iterates helper.importName into fresh helperNames, then sorts only helperNames. Other path passes helpers to getOrCreateExternalHelpersModuleNameIfNeeded, which only tests some(helpers). No array identity test or write to the helpers array during these reads. Other emit-helper writers can reach sourceFile.emitNode.helpers, but no such write is called during this local lifetime.

### C15: unknown - src/compiler/transformers/classFields.ts:853:25

Enclosing function: `visitMethodOrAccessorDeclaration`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
node.modifiers
```

Type at call: `NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
(m): m is Modifier => isModifier(m) && !isStaticModifier(m) && !isAccessorModifier(m)
```

node.modifiers -> filter, as the first argument value kept on the call stack -> visitParameterList and visitFunctionBody visitor calls in later arguments -> factory.createFunctionExpression.modifiers -> factory.createAssignment -> getPendingExpressions().push. createNodeArray can retain an existing NodeArray. The original node still owns modifiers; visitor calls occur while the filtered result is live even before factory invocation, and the resulting pending expression can retain it afterward. Those visitor and escaped aliases were not closed.

### C16: unknown - src/compiler/transformers/classFields.ts:1708:16

Enclosing function: `getPrivateInstanceMethodsAndAccessors`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
node.members
```

Type at call: `NodeArray<ClassElement>`.
Declared annotation(s): `NodeArray<ClassElement>`.

Predicate:
```ts
isNonStaticMethodOrAccessorWithPrivateName
```

node.members -> filter -> getPrivateInstanceMethodsAndAccessors return -> classFields some/index reads and transformConstructorBody/initializer generation. Several uses cross visitor calls. The original class still owns members; transitive writes while the returned view is live were not closed.

### C17: unknown - src/compiler/transformers/ts.ts:1054:13

Enclosing function: `transformClassMembers`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
constructor.parameters
```

Type at call: `NodeArray<ParameterDeclaration>`.
Declared annotation(s): `NodeArray<ParameterDeclaration>`.

Predicate:
```ts
(p): p is ParameterPropertyDeclaration => isParameterPropertyDeclaration(p, constructor)
```

constructor.parameters -> filter -> parametersWithPropertyAssignments -> for-of creating synthetic properties -> newMembers. No local write to parameters or identity test. Factory/setOriginalNode calls occur before later iterations, so the full transitive callback/alias write graph of this shared constructor array remains unclosed.

### C18: safe - src/compiler/transformers/ts.ts:1087:38

Enclosing function: `injectClassTypeMetadata`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
modifiers
```

Type at call: `NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
isDecorator
```

modifiers parameter -> filter -> addRange(modifiersArray, result). modifiersArray is a fresh local destination; addRange only reads the result and pushes destination elements. The filtered array dies on return from addRange. Later metadata insertion and reassignment of modifiers do not write the input array. No identity/key use.

### C19: safe - src/compiler/transformers/ts.ts:1089:38

Enclosing function: `injectClassTypeMetadata`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (production).

Argument:
```ts
skipWhile(modifiers, isExportOrDefaultModifier)
```

Type at call: `(AbstractKeyword | AccessorKeyword | AsyncKeyword | ConstKeyword | DeclareKeyword | Decorator | ... 9 more ... | StaticKeyword)[] | undefined`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isModifier
```

modifiers -> skipWhile -> filter -> addRange into fresh modifiersArray. skipWhile may return modifiers itself when its prefix is empty, so the argument may be node-owned. addRange only reads it and copies elements; later operations use the separate destination. No identity/key use or input-array write while live.

### C20: safe - src/compiler/transformers/ts.ts:1101:42

Enclosing function: `injectClassElementTypeMetadata`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
modifiers
```

Type at call: `NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
isDecorator
```

modifiers parameter -> filter(isDecorator) -> addRange into fresh modifiersArray. Input is node-associated and shared, but addRange copies elements and does not call an input writer. The result is dead before subsequent addRange metadata insertion. No identity/key use.

### C21: safe - src/compiler/transformers/ts.ts:1103:42

Enclosing function: `injectClassElementTypeMetadata`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
modifiers
```

Type at call: `NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
isModifier
```

modifiers parameter -> filter(isModifier) -> addRange into fresh modifiersArray. The only destination write is to modifiersArray. The result dies before factory.createNodeArray and modifiers reassignment. No array identity comparison or input writer while live.

### C22: unknown - src/compiler/transformers/ts.ts:1384:13

Enclosing function: `transformConstructorBody`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
constructor.parameters
```

Type at call: `NodeArray<ParameterDeclaration>`.
Declared annotation(s): `NodeArray<ParameterDeclaration>`.

Predicate:
```ts
p => isParameterPropertyDeclaration(p, constructor)
```

constructor.parameters -> filter -> parametersWithPropertyAssignments -> some -> copyPrologue(visitor) -> mapDefined(transformParameterWithPropertyAssignment). The same constructor parameter array is node-owned and remains live across visitor execution. No direct local identity/write was found; arbitrary visitor alias writes were not closed.

### C23: unknown - src/compiler/transformers/utilities.ts:603:12

Enclosing function: `getStaticPropertiesAndClassStaticBlock`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
node.members
```

Type at call: `NodeArray<ClassElement>`.
Declared annotation(s): `NodeArray<ClassElement>`.

Predicate:
```ts
isStaticPropertyDeclarationOrClassStaticBlockDeclaration
```

node.members -> filter -> getStaticPropertiesAndClassStaticBlock return -> several classFields transformer paths, including transforms of initializers/static blocks. The class node retains members; the result can span visitor calls. All visitor and factory alias writes were not exhaustively closed.

### C24: safe - src/compiler/utilities.ts:4690:20

Enclosing function: `filterOwnedJSDocTags::<flatMap callback>`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
jsDoc.tags
```

Type at call: `NodeArray<JSDocTag> | undefined`.
Declared annotation(s): `NodeArray<JSDocTag> | undefined`.

Predicate:
```ts
isJSDocOverloadTag
```

jsDoc.tags -> filter -> flatMap mapper return -> addRange into flatMap result -> getJSDocCommentsAndTags. addRange(undefined, from) slices and later calls push into its separate destination. The filtered array does not escape this copy boundary. The adjacent ownedTags identity comparison is in the other branch and uses the boolean overload, not this site.

### C25: unknown - src/compiler/utilities.ts:8018:12

Enclosing function: `getInitializedVariables`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (production).

Argument:
```ts
node.declarations
```

Type at call: `NodeArray<VariableDeclaration>`.
Declared annotation(s): `NodeArray<VariableDeclaration>`.

Predicate:
```ts
isInitializedVariable
```

node.declarations -> filter -> getInitializedVariables return -> es2017, generators and ts transformers. The declaration list remains reachable from the input node. Its returned view crosses several transform calls; all downstream aliases and callbacks have not been closed. The function is exported.

### C26: unknown - src/compiler/utilitiesPublic.ts:1026:16

Enclosing function: `getDecorators`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
node.modifiers
```

Type at call: `NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
isDecorator
```

node.modifiers -> filter -> getDecorators return -> public callers can retain the same NodeArray. No comparison or write in this getter. The node owns the argument and callers can also reach node.modifiers; the escaped result lifetime and all external callers were not closed.

### C27: unknown - src/compiler/utilitiesPublic.ts:1032:16

Enclosing function: `getModifiers`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (production).

Argument:
```ts
node.modifiers
```

Type at call: `NodeArray<Modifier> | NodeArray<ModifierLike> | undefined`.
Declared annotation(s): `NodeArray<Modifier> | undefined`, `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
isModifier
```

node.modifiers -> filter -> getModifiers return -> public callers. The getter only reads. The node retains the same modifier array; public caller identity tests and writes while the returned view is retained were not exhaustively followed.

### S01: unknown - src/services/callHierarchy.ts:482:19

Enclosing function: `getIncomingCalls`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
FindAllReferences.findReferenceOrRenameEntries(program, cancellationToken, program.getSourceFiles(), location, /*position*/ 0, { use: FindAllReferences.FindReferencesUse.References }, convertEntryToCallSite)
```

Type at call: `(CallSite | undefined)[] | undefined`.
No variable/field annotation on this expression; the type above is inferred/instantiated.

Predicate:
```ts
isDefined
```

findReferenceOrRenameEntries returns a search result array -> filter(isDefined) -> calls -> group -> convertCallSiteGroupToIncomingCall. The original search array is not named locally. Producer retention/aliasing and all group conversion calls were not proved free of writes to the original, so this remains unknown.

### S02: unknown - src/services/codefixes/convertFunctionToEs6Class.ts:314:39

Enclosing function: `getModifierKindFromSource`. Resolved overload: **7**, `src/compiler/core.ts:273:1` (supplementary stock checker).

Argument:
```ts
source.modifiers
```

Type at call: `import("../_namespaces/ts.js").NodeArray<import("../_namespaces/ts.js").ModifierLike> | import("../_namespaces/ts.js").NodeArray<Modifier> | undefined`.
Declared annotation(s): `NodeArray<Modifier> | undefined`, `NodeArray<ModifierLike> | undefined`.

Predicate:
```ts
(modifier): modifier is Modifier => modifier.kind === kind
```

source.modifiers -> filter -> getModifierKindFromSource return -> concatenate/factory modifier construction in function-to-class conversion. The source node keeps its NodeArray. concatenate can preserve an input when the other side is empty; the result can reach synthetic node fields. Later transformed-node lifetime and aliases were not closed.

### S03: unknown - src/services/codefixes/convertToEsModule.ts:137:29

Enclosing function: `convertFileToEsModule`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (supplementary stock checker).

Argument:
```ts
sourceFile.statements
```

Type at call: `NodeArray<Statement>`.
Declared annotation(s): `NodeArray<Statement>`.

Predicate:
```ts
isVariableStatement
```

sourceFile.statements -> filter(isVariableStatement) -> for-of -> convertVariableStatement -> ChangeTracker. The input node remains reachable. No local array identity/write, but the transitive conversion/checker/change-tracker aliases while the iterator is live were not exhaustively closed.

### S04: unknown - src/services/codefixes/fixAddMissingParam.ts:143:46

Enclosing function: `getInfo`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
type.symbol.declarations
```

Type at call: `import("../_namespaces/ts.js").Declaration[] | undefined`.
Declared annotation(s): `Declaration[] | undefined`.

Predicate:
```ts
isConvertibleSignatureDeclaration
```

type.symbol.declarations -> filter -> convertibleSignatureDeclarations -> lastOrUndefined -> isSourceFileFromLibrary(program, ...), tryGetName and size checks -> getOverloads -> spread into a fresh declarations array -> SignatureInfo. Only individual declaration nodes escape after the spread; the later checker/typeToTypeNode loop does not extend this array lifetime. The shared symbol array is still live across the earlier Program/source-file-library query. Producer ownership and those earlier transitive alias paths were not closed; no local identity comparison or array write was found.

### S05: unknown - src/services/goToDefinition.ts:180:30

Enclosing function: `getDefinitionAtPosition`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (supplementary stock checker).

Argument:
```ts
classDecl.members
```

Type at call: `import("./_namespaces/ts.js").NodeArray<import("./_namespaces/ts.js").ClassElement>`.
Declared annotation(s): `NodeArray<ClassElement>`.

Predicate:
```ts
isClassStaticBlockDeclaration
```

classDecl.members -> filter -> staticBlocks -> symbolToString -> map into DefinitionInfo records. The class node retains the argument. The result crosses symbolToString before its final read; all reentrant checker writes during that call were not closed. Only DefinitionInfo records escape.

### S06: unknown - src/services/services.ts:745:94

Enclosing function: `getContextualDocumentationComment`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
this.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[]`.

Predicate:
```ts
isGetAccessor
```

this.declarations -> filter(isGetAccessor) -> getDocumentationComment -> JsDoc extraction and inherited checker/symbol calls. Only display parts are cached on this, not the filtered array. The symbol owns declarations. No direct array identity/write; transitive checker and inherited-symbol callbacks were not closed.

### S07: unknown - src/services/services.ts:755:94

Enclosing function: `getContextualDocumentationComment`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
this.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[]`.

Predicate:
```ts
isSetAccessor
```

this.declarations -> filter(isSetAccessor) -> getDocumentationComment -> JsDoc extraction and inherited checker/symbol calls. The filtered array is consumed synchronously; only display parts escape. The shared declarations array and transitive reentrant checker writes were not closed.

### S08: unknown - src/services/services.ts:780:81

Enclosing function: `getContextualJsDocTags`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
this.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[]`.

Predicate:
```ts
isGetAccessor
```

this.declarations -> filter(isGetAccessor) -> getJsDocTagsOfDeclarations -> JsDoc extraction and inherited checker/symbol calls. Only JSDocTagInfo elements are cached. The input remains symbol-owned. No local identity/write, but transitive checker/reentrant alias writes were not closed.

### S09: unknown - src/services/services.ts:790:81

Enclosing function: `getContextualJsDocTags`. Resolved overload: **5**, `src/compiler/core.ts:269:1` (supplementary stock checker).

Argument:
```ts
this.declarations
```

Type at call: `Declaration[] | undefined`.
Declared annotation(s): `Declaration[]`.

Predicate:
```ts
isSetAccessor
```

this.declarations -> filter(isSetAccessor) -> getJsDocTagsOfDeclarations -> JsDoc extraction and inherited checker/symbol calls. Only JSDocTagInfo results escape. Symbol declaration aliases and reentrant checker writes during consumption were not closed.

### S10: unknown - src/services/utilities.ts:2616:38

Enclosing function: `insertImports`. Resolved overload: **3**, `src/compiler/core.ts:265:1` (supplementary stock checker).

Argument:
```ts
sourceFile.statements
```

Type at call: `NodeArray<import("./_namespaces/ts.js").Statement> | readonly never[]`.
Declared annotation(s): `NodeArray<Statement>`, `readonly never[]`.

Predicate:
```ts
importKindPredicate
```

sourceFile.statements -> filter -> existingImportStatements -> comparer detection and insertion-index helpers -> ChangeTracker calls. The argument is shared through sourceFile. ChangeTracker queues text edits rather than mutating statements locally, but comparer and change-tracker transitive alias behavior was not exhaustively closed.

## Reproduction and evidence

Use fresh output directories for a rerun; the paths below identify this measurement. Source and Go implementation files outside this report directory are unchanged. Evidence Go is stored as `.go.txt`. Build it using a Go overlay that maps `stage3/census/tool/main.go` to `sites.go.txt`; no loader or checker file is overlaid.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/filter-sites-setup-main.log 2>&1
npm ci --prefix stage3/api > /tmp/filter-sites-npm.log 2>&1
npm ci --prefix stage3/typescript > /tmp/filter-sites-runtime-npm.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3
bash stage3/apply.sh /workspace/scratch/filter-sites-tree > /tmp/filter-sites-apply.log 2>&1
mkdir -p /workspace/scratch/filter-sites
python3 - <<'PYOVERLAY'
from pathlib import Path
import json
repository = Path.cwd()
Path('/workspace/scratch/filter-sites/overlay.json').write_text(json.dumps({'Replace': {
    str(repository / 'stage3/census/tool/main.go'): str(repository / 'stage3/census/filter-overload-sites/sites.go.txt')
}}))
PYOVERLAY
go build -overlay=/workspace/scratch/filter-sites/overlay.json -o /workspace/scratch/filter-sites/sites-scopes ./stage3/census/tool
/workspace/scratch/filter-sites/sites-scopes /workspace/scratch/filter-sites-tree /workspace/scratch/filter-sites/production-compiler.json compiler
node stage3/census/filter-overload-sites/supplement.cjs /workspace/scratch/filter-sites-tree /workspace/scratch/filter-sites/supplement-exhaustive.json
python3 stage3/census/filter-overload-sites/build-private.py /workspace/adamic /workspace/scratch/filter-sites-tree /workspace/scratch/filter-sites/build-control
cp -a /workspace/scratch/filter-sites-tree /workspace/scratch/filter-sites-counting
node stage3/census/filter-overload-sites/instrument.cjs /workspace/scratch/filter-sites-counting /workspace/scratch/filter-sites/supplement-exhaustive.json
python3 stage3/census/filter-overload-sites/build-private.py /workspace/adamic /workspace/scratch/filter-sites-counting /workspace/scratch/filter-sites/build-counting
FILTER_SITE_COUNTS=/workspace/scratch/filter-sites/counts-final NODE_OPTIONS=--require=/workspace/adamic/stage3/census/filter-overload-sites/observe.cjs STAGE3_ADAPTED_TREE=/workspace/scratch/filter-sites-counting python3 -B stage3/lane/measure_shards.py /workspace/scratch/filter-sites/lane-final --concurrency 1
```

Actual child exits and all retained logs are evidence, not inferred from tee/pipeline status. A premature pre-completion scan is excluded; all definitive static rows refer to the completed, hashed source snapshot. The initial stale-pin setup/build attempt is also excluded; the required setup succeeded after aligning main dependency pins.

Final counter aggregation used `collect.py --require-complete`; the 67 exact manifest-command repeats used `repeat-units.py`. The initial full serial run also passed all 398 units, but its 67 earlier-observer units are replaced in the counting cohort. The source tree used for production checking remained byte-identical to the completed pre-measurement snapshot.
