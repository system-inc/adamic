conditions: full latent census

Ref `origin/codex/conditions`; compiler `a4218d46fc5610442caefcf6d998baaaecb740b2`.

Counts deduplicate (kind, where, reason, text) across refusal scanning and lowering. Actual-lowering columns retain phase lowering only.
Owner groups follow the exact pinned dispositions: compiler lesson -> compiler; adaptation -> stage3 adaptation. Absent exact reasons are UNOWNED; no inferred family disposition is silently substituted.

The tsc entry reach is the primary scoreboard. The compiler-directory census is retained alongside it. Both are measurement-only.

## compiler

measured on a checker-rejected program.

79 source files; 324 checker diagnostics; unit statuses {'attempted': 10544, 'split_checker_body': 147}.
Excluded nested functions: 110; gross body bytes 437999 (can overlap); union bytes 414509.
State-snapshot failure boundaries: 0.

| kind | combined | actual_lowering |
| --- | --- | --- |
| NotYet | 13934 | 13913 |
| Refused | 9457 | 2067 |
| panic | 20 | 18 |
| error | 2 | 2 |
| SkippedDependency | 4 | 4 |
| Boundary | 20628 | 20628 |

| owner | NotYet | Refused |
| --- | --- | --- |
| compiler | 9460 | 37 |
| stage3 adaptation | 134 | 5093 |
| UNOWNED | 4340 | 4327 |

### compiler

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| NotYet | a method call through a structural signature in a program with statics; use typeof the declaring class | 1824 | 1824 | compiler lesson |  | False |
| NotYet | reading node | 1081 | 1081 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a value and a value | 516 | 516 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a boolean | 321 | 321 | compiler lesson |  | False |
| NotYet | reading result | 234 | 234 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a number | 221 | 221 | compiler lesson |  | False |
| NotYet | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 191 | 191 | compiler lesson |  | False |
| NotYet | assigning to an Identifier | 165 | 165 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a value | 143 | 143 | compiler lesson |  | False |
| NotYet | a call through ?. (an optional call) | 129 | 129 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number and a boolean | 125 | 125 | compiler lesson |  | False |
| NotYet | for...of over an object | 112 | 112 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a number | 109 | 109 | compiler lesson |  | False |
| NotYet | reading type | 102 | 102 | compiler lesson |  | True |
| NotYet | a function returning T | 101 | 101 | compiler lesson |  | False |
| NotYet | a void call used as a value | 85 | 85 | compiler lesson |  | False |
| NotYet | an ElementAccessExpression | 82 | 82 | compiler lesson |  | False |
| NotYet | a function returning T &#124; undefined | 75 | 75 | compiler lesson |  | False |
| NotYet | a declaration directly in a case (wrap the case in a block) | 61 | 61 | compiler lesson |  | False |
| NotYet | reading performance | 57 | 57 | compiler lesson |  | True |
| NotYet | a value of type T | 48 | 48 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a number | 47 | 47 | compiler lesson |  | False |
| NotYet | an array of T | 47 | 47 | compiler lesson |  | False |
| NotYet | a rest array of DiagnosticArguments | 46 | 46 | compiler lesson |  | False |
| NotYet | a call returning void &#124; undefined | 43 | 43 | compiler lesson |  | False |
| NotYet | reading expression | 42 | 42 | compiler lesson |  | True |
| NotYet | reading symbol | 42 | 42 | compiler lesson |  | True |
| NotYet | reading host | 40 | 40 | compiler lesson |  | True |
| NotYet | reading updated | 38 | 38 | compiler lesson |  | True |
| NotYet | reading name | 37 | 37 | compiler lesson |  | True |
| Refused | a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 37 | 0 | compiler lesson | internal/lower/predicates.go / predicateProof.refused | False |
| NotYet | this outside a method | 34 | 34 | compiler lesson |  | False |
| NotYet | a function value returning union of differently held members | 33 | 33 | compiler lesson |  | False |
| NotYet | reading statement | 29 | 29 | compiler lesson |  | True |
| NotYet | reading container | 28 | 28 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number &#124; undefined and a number | 27 | 27 | compiler lesson |  | False |
| NotYet | reading resolved | 25 | 25 | compiler lesson |  | True |
| NotYet | reading clone | 24 | 24 | compiler lesson |  | True |
| NotYet | reading declaration | 24 | 24 | compiler lesson |  | True |
| NotYet | a field of type string &#124; NodeArray<JSDocComment> &#124; undefined | 23 | 23 | compiler lesson |  | False |
| NotYet | reading compilerOptions | 23 | 23 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a string and a string | 21 | 21 | compiler lesson |  | False |
| NotYet | a value of type SolutionBuilderState<T> | 21 | 21 | compiler lesson |  | False |
| NotYet | new an Identifier | 21 | 21 | compiler lesson |  | False |
| NotYet | reading sourceFile | 21 | 21 | compiler lesson |  | True |
| NotYet | a value of type unknown | 20 | 20 | compiler lesson |  | False |
| NotYet | reading options | 20 | 20 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a value | 19 | 19 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a number | 19 | 19 | compiler lesson |  | False |
| NotYet | reading target | 19 | 19 | compiler lesson |  | True |
| NotYet | a value of type ResolvedConfigFilePath | 18 | 18 | compiler lesson |  | False |
| NotYet | reading c | 18 | 18 | compiler lesson |  | True |
| NotYet | reading types | 18 | 18 | compiler lesson |  | True |
| NotYet | a generic function as a value | 17 | 17 | compiler lesson |  | False |
| NotYet | reading expr | 17 | 17 | compiler lesson |  | True |
| NotYet | a value of type object | 16 | 16 | compiler lesson |  | False |
| NotYet | reading visited | 16 | 16 | compiler lesson |  | True |
| NotYet | replacing a represented method at runtime | 16 | 16 | compiler lesson |  | False |
| NotYet | optional chaining to .size on a value | 15 | 15 | compiler lesson |  | False |
| NotYet | reading modifiers | 15 | 15 | compiler lesson |  | True |
| NotYet | reading parameter | 15 | 15 | compiler lesson |  | True |
| NotYet | reading diag | 14 | 14 | compiler lesson |  | True |
| NotYet | reading reference | 14 | 14 | compiler lesson |  | True |
| NotYet | a value of type T &#124; undefined | 13 | 13 | compiler lesson |  | False |
| NotYet | reading block | 13 | 13 | compiler lesson |  | True |
| NotYet | reading sig | 13 | 13 | compiler lesson |  | True |
| NotYet | a function returning undefined | 12 | 12 | compiler lesson |  | False |
| NotYet | a value of type ResolvedConfigFileName | 12 | 12 | compiler lesson |  | False |
| NotYet | reading cache | 12 | 12 | compiler lesson |  | True |
| NotYet | reading file | 12 | 12 | compiler lesson |  | True |
| NotYet | a value of type NonNullable<T> | 11 | 11 | compiler lesson |  | False |
| NotYet | reading child | 11 | 11 | compiler lesson |  | True |
| NotYet | reading diagnostics | 11 | 11 | compiler lesson |  | True |
| NotYet | reading identifier | 11 | 11 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a number | 10 | 10 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a boolean | 10 | 10 | compiler lesson |  | False |
| NotYet | a function returning U &#124; undefined | 10 | 10 | compiler lesson |  | False |
| NotYet | for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10 | 10 | compiler lesson |  | False |
| NotYet | reading exception | 10 | 10 | compiler lesson |  | True |
| NotYet | reading existing | 10 | 10 | compiler lesson |  | True |
| NotYet | reading firstDecorator | 10 | 10 | compiler lesson |  | True |
| NotYet | reading flowType | 10 | 10 | compiler lesson |  | True |
| NotYet | reading objectFlags | 10 | 10 | compiler lesson |  | True |
| NotYet | reading specifier | 10 | 10 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a string | 9 | 9 | compiler lesson |  | False |
| NotYet | a SpreadElement | 9 | 9 | compiler lesson |  | False |
| NotYet | a function returning T[] | 9 | 9 | compiler lesson |  | False |
| NotYet | a value of type T["kind"] | 9 | 9 | compiler lesson |  | False |
| NotYet | a value of type readonly T[] &#124; undefined | 9 | 9 | compiler lesson |  | False |
| NotYet | an optional chain longer than one step | 9 | 9 | compiler lesson |  | False |
| NotYet | assigning a field of a value | 9 | 9 | compiler lesson |  | False |
| NotYet | reading decl | 9 | 9 | compiler lesson |  | True |
| NotYet | reading id | 9 | 9 | compiler lesson |  | True |
| NotYet | reading iterationTypes | 9 | 9 | compiler lesson |  | True |
| NotYet | reading left | 9 | 9 | compiler lesson |  | True |
| NotYet | reading links | 9 | 9 | compiler lesson |  | True |
| NotYet | reading parsed | 9 | 9 | compiler lesson |  | True |
| NotYet | reading temp | 9 | 9 | compiler lesson |  | True |
| NotYet | reading value | 9 | 9 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a boolean &#124; undefined | 8 | 8 | compiler lesson |  | False |
| NotYet | a value of type K | 8 | 8 | compiler lesson |  | False |
| NotYet | new a ParenthesizedExpression | 8 | 8 | compiler lesson |  | False |
| NotYet | reading related | 8 | 8 | compiler lesson |  | True |
| NotYet | reading statements | 8 | 8 | compiler lesson |  | True |
| NotYet | regex replacement other than a string | 8 | 8 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a boolean &#124; undefined | 7 | 7 | compiler lesson |  | False |
| NotYet | a rest parameter outside a nongeneric named function | 7 | 7 | compiler lesson |  | False |
| NotYet | a value of type ((node: Node) => boolean) &#124; undefined | 7 | 7 | compiler lesson |  | False |
| NotYet | a value of type BindableStaticNameExpression | 7 | 7 | compiler lesson |  | False |
| NotYet | a value of type readonly T[] | 7 | 7 | compiler lesson |  | False |
| NotYet | assigning to a NonNullExpression | 7 | 7 | compiler lesson |  | False |
| NotYet | reading array | 7 | 7 | compiler lesson |  | True |
| NotYet | reading autoGenerate | 7 | 7 | compiler lesson |  | True |
| NotYet | reading buildOrder | 7 | 7 | compiler lesson |  | True |
| NotYet | reading candidate | 7 | 7 | compiler lesson |  | True |
| NotYet | reading constraint | 7 | 7 | compiler lesson |  | True |
| NotYet | reading current | 7 | 7 | compiler lesson |  | True |
| NotYet | reading directoryWatcher | 7 | 7 | compiler lesson |  | True |
| NotYet | reading errorNode | 7 | 7 | compiler lesson |  | True |
| NotYet | reading flags | 7 | 7 | compiler lesson |  | True |
| NotYet | reading includes | 7 | 7 | compiler lesson |  | True |
| NotYet | reading parent | 7 | 7 | compiler lesson |  | True |
| NotYet | reading targetType | 7 | 7 | compiler lesson |  | True |
| NotYet | reading text | 7 | 7 | compiler lesson |  | True |
| NotYet | reading token | 7 | 7 | compiler lesson |  | True |
| NotYet | reading typeNode | 7 | 7 | compiler lesson |  | True |
| NotYet | reading watcher | 7 | 7 | compiler lesson |  | True |
| NotYet | ?.[] on a value | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a string | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a boolean | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a string | 6 | 6 | compiler lesson |  | False |
| NotYet | a case whose type differs from the switch's | 6 | 6 | compiler lesson |  | False |
| NotYet | a value of type HasJSDoc &#124; undefined | 6 | 6 | compiler lesson |  | False |
| NotYet | a value of type TOuterState | 6 | 6 | compiler lesson |  | False |
| NotYet | an array of U | 6 | 6 | compiler lesson |  | False |
| NotYet | an array of boolean &#124; undefined | 6 | 6 | compiler lesson |  | False |
| NotYet | reading bundle | 6 | 6 | compiler lesson |  | True |
| NotYet | reading classType | 6 | 6 | compiler lesson |  | True |
| NotYet | reading e | 6 | 6 | compiler lesson |  | True |
| NotYet | reading elem | 6 | 6 | compiler lesson |  | True |
| NotYet | reading innerExpression | 6 | 6 | compiler lesson |  | True |
| NotYet | reading lexicallyScopedSymbol | 6 | 6 | compiler lesson |  | True |
| NotYet | reading location | 6 | 6 | compiler lesson |  | True |
| NotYet | reading queue | 6 | 6 | compiler lesson |  | True |
| NotYet | reading reduced | 6 | 6 | compiler lesson |  | True |
| NotYet | reading res | 6 | 6 | compiler lesson |  | True |
| NotYet | reading scope | 6 | 6 | compiler lesson |  | True |
| NotYet | reading t | 6 | 6 | compiler lesson |  | True |
| NotYet | reading typeParameters | 6 | 6 | compiler lesson |  | True |
| NotYet | reading v | 6 | 6 | compiler lesson |  | True |
| NotYet | reading valueDeclaration | 6 | 6 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a string and a value | 5 | 5 | compiler lesson |  | False |
| NotYet | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type HasJSDoc | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type string &#124; null &#124; undefined | 5 | 5 | compiler lesson |  | False |
| NotYet | an array of ResolvedConfigFileName | 5 | 5 | compiler lesson |  | False |
| NotYet | assigning to an ObjectLiteralExpression | 5 | 5 | compiler lesson |  | False |
| NotYet | reading args | 5 | 5 | compiler lesson |  | True |
| NotYet | reading assignedName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading body | 5 | 5 | compiler lesson |  | True |
| NotYet | reading buildOptions | 5 | 5 | compiler lesson |  | True |
| NotYet | reading cached | 5 | 5 | compiler lesson |  | True |
| NotYet | reading directoryExists | 5 | 5 | compiler lesson |  | True |
| NotYet | reading emittedExpression | 5 | 5 | compiler lesson |  | True |
| NotYet | reading exportSpecifiers | 5 | 5 | compiler lesson |  | True |
| NotYet | reading index | 5 | 5 | compiler lesson |  | True |
| NotYet | reading indexInfo | 5 | 5 | compiler lesson |  | True |
| NotYet | reading indexType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading info | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isAmbient | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isAsync | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isGenerator | 5 | 5 | compiler lesson |  | True |
| NotYet | reading moduleSymbol | 5 | 5 | compiler lesson |  | True |
| NotYet | reading objectType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading parseNode | 5 | 5 | compiler lesson |  | True |
| NotYet | reading path | 5 | 5 | compiler lesson |  | True |
| NotYet | reading pos | 5 | 5 | compiler lesson |  | True |
| NotYet | reading propType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading propertyName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading source | 5 | 5 | compiler lesson |  | True |
| NotYet | reading start | 5 | 5 | compiler lesson |  | True |
| NotYet | reading targetParam | 5 | 5 | compiler lesson |  | True |
| NotYet | reading typeName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading yieldType | 5 | 5 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a value | 4 | 4 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a boolean | 4 | 4 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a value | 4 | 4 | compiler lesson |  | False |
| NotYet | a computed field name | 4 | 4 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 4 | 4 | compiler lesson |  | False |
| NotYet | a for...of destructuring an object | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning CompilerOptionsValue | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning NodeArray<T> | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning U | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type (ConstructorDeclaration & { body: Block; }) &#124; undefined | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type CompilerOptionsValue | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type ImmediatelyInvokedArrowFunction | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type T &#124; Program | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type T[] | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type object &#124; undefined | 4 | 4 | compiler lesson |  | False |
| NotYet | assigning an element of a value | 4 | 4 | compiler lesson |  | False |
| NotYet | reading argument | 4 | 4 | compiler lesson |  | True |
| NotYet | reading asteriskToken | 4 | 4 | compiler lesson |  | True |
| NotYet | reading bindings | 4 | 4 | compiler lesson |  | True |
| NotYet | reading cacheKey | 4 | 4 | compiler lesson |  | True |
| NotYet | reading chain | 4 | 4 | compiler lesson |  | True |
| NotYet | reading clause | 4 | 4 | compiler lesson |  | True |
| NotYet | reading commonSourceDirectory | 4 | 4 | compiler lesson |  | True |
| NotYet | reading config | 4 | 4 | compiler lesson |  | True |
| NotYet | reading contextualType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading cooked | 4 | 4 | compiler lesson |  | True |
| NotYet | reading descriptorName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading element | 4 | 4 | compiler lesson |  | True |
| NotYet | reading emitNode | 4 | 4 | compiler lesson |  | True |
| NotYet | reading emitResult | 4 | 4 | compiler lesson |  | True |
| NotYet | reading excludeRegex | 4 | 4 | compiler lesson |  | True |
| NotYet | reading exportStatement | 4 | 4 | compiler lesson |  | True |
| NotYet | reading exportedName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading extensionGroup | 4 | 4 | compiler lesson |  | True |
| NotYet | reading fakeScope | 4 | 4 | compiler lesson |  | True |
| NotYet | reading hasDefaultClause | 4 | 4 | compiler lesson |  | True |
| NotYet | reading importDeclaration | 4 | 4 | compiler lesson |  | True |
| NotYet | reading initializer | 4 | 4 | compiler lesson |  | True |
| NotYet | reading jsxFactorySymbol | 4 | 4 | compiler lesson |  | True |
| NotYet | reading key | 4 | 4 | compiler lesson |  | True |
| NotYet | reading kind | 4 | 4 | compiler lesson |  | True |
| NotYet | reading map | 4 | 4 | compiler lesson |  | True |
| NotYet | reading mapping | 4 | 4 | compiler lesson |  | True |
| NotYet | reading nameType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading nodeModulesFolderExists | 4 | 4 | compiler lesson |  | True |
| NotYet | reading normalized | 4 | 4 | compiler lesson |  | True |
| NotYet | reading openParenPosition | 4 | 4 | compiler lesson |  | True |
| NotYet | reading original | 4 | 4 | compiler lesson |  | True |
| NotYet | reading output | 4 | 4 | compiler lesson |  | True |
| NotYet | reading propertyOriginalNode | 4 | 4 | compiler lesson |  | True |
| NotYet | reading props | 4 | 4 | compiler lesson |  | True |
| NotYet | reading rawText | 4 | 4 | compiler lesson |  | True |
| NotYet | reading refPath | 4 | 4 | compiler lesson |  | True |
| NotYet | reading referencedName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading restParameter | 4 | 4 | compiler lesson |  | True |
| NotYet | reading restType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading seen | 4 | 4 | compiler lesson |  | True |
| NotYet | reading signature | 4 | 4 | compiler lesson |  | True |
| NotYet | reading sourceIndex | 4 | 4 | compiler lesson |  | True |
| NotYet | reading typeArguments | 4 | 4 | compiler lesson |  | True |
| NotYet | reading varStatement | 4 | 4 | compiler lesson |  | True |
| NotYet | reading watchCompilerHost | 4 | 4 | compiler lesson |  | True |
| NotYet | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 3 | 3 | compiler lesson |  | False |
| NotYet | Object.entries on a shape not proven by a plain literal or its const binding | 3 | 3 | compiler lesson |  | False |
| NotYet | RegExp with a nonconstant pattern | 3 | 3 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a boolean | 3 | 3 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a union of differently held members | 3 | 3 | compiler lesson |  | False |
| NotYet | a ConditionalExpression as a statement | 3 | 3 | compiler lesson |  | False |
| NotYet | a Map of false &#124; MutableFileSystemEntries | 3 | 3 | compiler lesson |  | False |
| NotYet | a PrefixUnaryExpression on a number | 3 | 3 | compiler lesson |  | False |
| NotYet | a comparator that doesn't take two elements and return a number | 3 | 3 | compiler lesson |  | False |
| NotYet | a destructured name that isn't plain | 3 | 3 | compiler lesson |  | False |
| NotYet | a field holding union of differently held members | 3 | 3 | compiler lesson |  | False |
| NotYet | a field of type NodeArray<ParameterDeclaration> &#124; readonly JSDocParameterTag[] | 3 | 3 | compiler lesson |  | False |
| NotYet | a field of type string &#124; number &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning ResolvedConfigFileName | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning T[] &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning object | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning readonly T[] &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type BindableAccessExpression | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type BindableObjectDefinePropertyCall | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type Children &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type EndOfFileToken | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type InitializedVariableDeclaration | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type U | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type WrappedExpression<AnonymousFunctionDefinition> | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type X | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type false &#124; TypeOnlyAliasDeclaration &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | an array of NonNullable<T> | 3 | 3 | compiler lesson |  | False |
| NotYet | an array of undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | destructuring anything but a tuple into [names] | 3 | 3 | compiler lesson |  | False |
| NotYet | lastIndexOf with these arguments | 3 | 3 | compiler lesson |  | False |
| NotYet | reading addUndefined | 3 | 3 | compiler lesson |  | True |
| NotYet | reading awaitedType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading baseTypes | 3 | 3 | compiler lesson |  | True |
| NotYet | reading buildInfo | 3 | 3 | compiler lesson |  | True |
| NotYet | reading callee | 3 | 3 | compiler lesson |  | True |
| NotYet | reading configFile | 3 | 3 | compiler lesson |  | True |
| NotYet | reading constructor | 3 | 3 | compiler lesson |  | True |
| NotYet | reading declaredType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading diagnosticStart | 3 | 3 | compiler lesson |  | True |
| NotYet | reading diff | 3 | 3 | compiler lesson |  | True |
| NotYet | reading dupFile | 3 | 3 | compiler lesson |  | True |
| NotYet | reading emitSuperHelpers | 3 | 3 | compiler lesson |  | True |
| NotYet | reading enclosingDeclaration | 3 | 3 | compiler lesson |  | True |
| NotYet | reading encodeURI | 3 | 3 | compiler lesson |  | True |
| NotYet | reading entityName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading evaluated | 3 | 3 | compiler lesson |  | True |
| NotYet | reading exports | 3 | 3 | compiler lesson |  | True |
| NotYet | reading filesSpecs | 3 | 3 | compiler lesson |  | True |
| NotYet | reading filtered | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstAccessorWithDecorators | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstDecl | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstDeclaration | 3 | 3 | compiler lesson |  | True |
| NotYet | reading functionName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading getCommonSourceDirectory | 3 | 3 | compiler lesson |  | True |
| NotYet | reading graphNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading hasRestParameter | 3 | 3 | compiler lesson |  | True |
| NotYet | reading hostNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading inferredProp | 3 | 3 | compiler lesson |  | True |
| NotYet | reading introducesError | 3 | 3 | compiler lesson |  | True |
| NotYet | reading isImmediatelyInvoked | 3 | 3 | compiler lesson |  | True |
| NotYet | reading isSimilarNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading jsFilePath | 3 | 3 | compiler lesson |  | True |
| NotYet | reading jsdocAliasDecl | 3 | 3 | compiler lesson |  | True |
| NotYet | reading keyPropertyName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading lastDecorator | 3 | 3 | compiler lesson |  | True |
| NotYet | reading lastError | 3 | 3 | compiler lesson |  | True |
| NotYet | reading length | 3 | 3 | compiler lesson |  | True |
| NotYet | reading loadPackageJsonMainState | 3 | 3 | compiler lesson |  | True |
| NotYet | reading mappedType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading match | 3 | 3 | compiler lesson |  | True |
| NotYet | reading method | 3 | 3 | compiler lesson |  | True |
| NotYet | reading methodType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading moveModifiers | 3 | 3 | compiler lesson |  | True |
| NotYet | reading named | 3 | 3 | compiler lesson |  | True |
| NotYet | reading oldEmitKind | 3 | 3 | compiler lesson |  | True |
| NotYet | reading oldOptions | 3 | 3 | compiler lesson |  | True |
| NotYet | reading operand | 3 | 3 | compiler lesson |  | True |
| NotYet | reading outerTypeParameters | 3 | 3 | compiler lesson |  | True |
| NotYet | reading parameters | 3 | 3 | compiler lesson |  | True |
| NotYet | reading parentType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading prop | 3 | 3 | compiler lesson |  | True |
| NotYet | reading property | 3 | 3 | compiler lesson |  | True |
| NotYet | reading questionToken | 3 | 3 | compiler lesson |  | True |
| NotYet | reading range | 3 | 3 | compiler lesson |  | True |
| NotYet | reading resolution | 3 | 3 | compiler lesson |  | True |
| NotYet | reading resolvedRequire | 3 | 3 | compiler lesson |  | True |
| NotYet | reading restElement | 3 | 3 | compiler lesson |  | True |
| NotYet | reading rhsValue | 3 | 3 | compiler lesson |  | True |
| NotYet | reading root | 3 | 3 | compiler lesson |  | True |
| NotYet | reading setter | 3 | 3 | compiler lesson |  | True |
| NotYet | reading state | 3 | 3 | compiler lesson |  | True |
| NotYet | reading suggestion | 3 | 3 | compiler lesson |  | True |
| NotYet | reading thisType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading tupleType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeArgumentTypes | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeCopy | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeSymbol | 3 | 3 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a boolean | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a number &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a string | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a number &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a Map of HostFileInfo | 2 | 2 | compiler lesson |  | False |
| NotYet | a Map of VisitResult<ExportAssignment &#124; LateVisibilityPaintedStatement &#124; undefined> | 2 | 2 | compiler lesson |  | False |
| NotYet | a destructured name held otherwise than its field | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type AnyBuildOrder &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type boolean &#124; (() => boolean) &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type false &#124; string[] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning NonNullable<T> | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning PackageJson[K] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning SortedReadonlyArray<T> | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning V | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning WatchFactory<X, Y>[T] | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning readonly T[] | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning void &#124; "skip" | 2 | 2 | compiler lesson |  | False |
| NotYet | a number &#124; undefined argument to slice | 2 | 2 | compiler lesson |  | False |
| NotYet | a number &#124; undefined argument to substring | 2 | 2 | compiler lesson |  | False |
| NotYet | a rest array of (string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined)[] | 2 | 2 | compiler lesson |  | False |
| NotYet | a rest parameter other than an array | 2 | 2 | compiler lesson |  | False |
| NotYet | a tagged template other than the intrinsic String.raw | 2 | 2 | compiler lesson |  | False |
| NotYet | a tuple element of type string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type AnonymousFunctionDefinition | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type CompilerHost & ReadBuildProgramHost | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type K &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type LeftHandSideExpression & Identifier | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ParameterPropertyDeclaration | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type SourceFile | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ThisCapturingVariableDeclaration | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type U &#124; readonly U[] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type U &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type V &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type false &#124; RegExpExecArray &#124; null | 2 | 2 | compiler lesson |  | False |
| NotYet | an array of Child | 2 | 2 | compiler lesson |  | False |
| NotYet | an array of V | 2 | 2 | compiler lesson |  | False |
| NotYet | destructuring a value | 2 | 2 | compiler lesson |  | False |
| NotYet | indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 | 2 | compiler lesson |  | False |
| NotYet | passing union of differently held members to a function value | 2 | 2 | compiler lesson |  | False |
| NotYet | push with other than one value | 2 | 2 | compiler lesson |  | False |
| NotYet | reading aComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading accessorDeclarations | 2 | 2 | compiler lesson |  | True |
| NotYet | reading allAccessors | 2 | 2 | compiler lesson |  | True |
| NotYet | reading allowStructuralFallback | 2 | 2 | compiler lesson |  | True |
| NotYet | reading antecedents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading arg | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assignClassAliasInStaticBlock | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assignment | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assumeInitialized | 2 | 2 | compiler lesson |  | True |
| NotYet | reading baseClassType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading baseObjectType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading basePath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading cacheAssignment | 2 | 2 | compiler lesson |  | True |
| NotYet | reading cachedPackageJson | 2 | 2 | compiler lesson |  | True |
| NotYet | reading call | 2 | 2 | compiler lesson |  | True |
| NotYet | reading callSignatures | 2 | 2 | compiler lesson |  | True |
| NotYet | reading callbackToAdd | 2 | 2 | compiler lesson |  | True |
| NotYet | reading capturedLeft | 2 | 2 | compiler lesson |  | True |
| NotYet | reading classDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading columns | 2 | 2 | compiler lesson |  | True |
| NotYet | reading commentRange | 2 | 2 | compiler lesson |  | True |
| NotYet | reading comments | 2 | 2 | compiler lesson |  | True |
| NotYet | reading compilerHost | 2 | 2 | compiler lesson |  | True |
| NotYet | reading constructSignatures | 2 | 2 | compiler lesson |  | True |
| NotYet | reading constructorDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading containingFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading contextualAwaitedType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading curCategory | 2 | 2 | compiler lesson |  | True |
| NotYet | reading currentNode | 2 | 2 | compiler lesson |  | True |
| NotYet | reading data | 2 | 2 | compiler lesson |  | True |
| NotYet | reading declBlocked | 2 | 2 | compiler lesson |  | True |
| NotYet | reading declarations | 2 | 2 | compiler lesson |  | True |
| NotYet | reading decorator | 2 | 2 | compiler lesson |  | True |
| NotYet | reading defaultIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading defaultType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading diagnostic | 2 | 2 | compiler lesson |  | True |
| NotYet | reading diagnosticWithLocation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading directoryPath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading elementTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading emitFlags | 2 | 2 | compiler lesson |  | True |
| NotYet | reading emittedOperand | 2 | 2 | compiler lesson |  | True |
| NotYet | reading enclosingClass | 2 | 2 | compiler lesson |  | True |
| NotYet | reading endLabel | 2 | 2 | compiler lesson |  | True |
| NotYet | reading equalsToken | 2 | 2 | compiler lesson |  | True |
| NotYet | reading errorInfo | 2 | 2 | compiler lesson |  | True |
| NotYet | reading errorRecord | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportEquals | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportStarFunction | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exported | 2 | 2 | compiler lesson |  | True |
| NotYet | reading externalHelpersImportDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading extraInitializersName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading first | 2 | 2 | compiler lesson |  | True |
| NotYet | reading firstArgument | 2 | 2 | compiler lesson |  | True |
| NotYet | reading firstParameterIsThis | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fixedEndLength | 2 | 2 | compiler lesson |  | True |
| NotYet | reading flatDiagnostics | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fromCache | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fromComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fullName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading func | 2 | 2 | compiler lesson |  | True |
| NotYet | reading generatorYieldType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading getCanonicalFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading getModifiedTime | 2 | 2 | compiler lesson |  | True |
| NotYet | reading getter | 2 | 2 | compiler lesson |  | True |
| NotYet | reading gutterWidth | 2 | 2 | compiler lesson |  | True |
| NotYet | reading hasDefault | 2 | 2 | compiler lesson |  | True |
| NotYet | reading helper | 2 | 2 | compiler lesson |  | True |
| NotYet | reading heritageClause | 2 | 2 | compiler lesson |  | True |
| NotYet | reading immediate | 2 | 2 | compiler lesson |  | True |
| NotYet | reading implementsTypeNodes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading indexInfos | 2 | 2 | compiler lesson |  | True |
| NotYet | reading initialType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading innerModuleSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading instantiation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading intrinsicElementsType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading invalidatedProject | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isDeferredMappedIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isFinite | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isJSDoc | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isOptionalChain | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isSetonlyAccessor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isUsed | 2 | 2 | compiler lesson |  | True |
| NotYet | reading jsdocParameters | 2 | 2 | compiler lesson |  | True |
| NotYet | reading label | 2 | 2 | compiler lesson |  | True |
| NotYet | reading last | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lastModifier | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lastParamVariadicType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading leadingNewlines | 2 | 2 | compiler lesson |  | True |
| NotYet | reading leftType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading len | 2 | 2 | compiler lesson |  | True |
| NotYet | reading line | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lineNumber | 2 | 2 | compiler lesson |  | True |
| NotYet | reading linesBeforeDot | 2 | 2 | compiler lesson |  | True |
| NotYet | reading linkType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading list | 2 | 2 | compiler lesson |  | True |
| NotYet | reading local | 2 | 2 | compiler lesson |  | True |
| NotYet | reading mapped | 2 | 2 | compiler lesson |  | True |
| NotYet | reading maxErrors | 2 | 2 | compiler lesson |  | True |
| NotYet | reading maybeParameters | 2 | 2 | compiler lesson |  | True |
| NotYet | reading members | 2 | 2 | compiler lesson |  | True |
| NotYet | reading methodSignatures | 2 | 2 | compiler lesson |  | True |
| NotYet | reading min | 2 | 2 | compiler lesson |  | True |
| NotYet | reading minor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading missing | 2 | 2 | compiler lesson |  | True |
| NotYet | reading modifier | 2 | 2 | compiler lesson |  | True |
| NotYet | reading mustBeRemoved | 2 | 2 | compiler lesson |  | True |
| NotYet | reading names | 2 | 2 | compiler lesson |  | True |
| NotYet | reading namespaceDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading needsParens | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newAliasSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newParametersArray | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newParsedCommandLine | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newSignature | 2 | 2 | compiler lesson |  | True |
| NotYet | reading nextKey | 2 | 2 | compiler lesson |  | True |
| NotYet | reading noTruncation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading nodeId | 2 | 2 | compiler lesson |  | True |
| NotYet | reading ok | 2 | 2 | compiler lesson |  | True |
| NotYet | reading openBracePosition | 2 | 2 | compiler lesson |  | True |
| NotYet | reading operatorKind | 2 | 2 | compiler lesson |  | True |
| NotYet | reading operatorToken | 2 | 2 | compiler lesson |  | True |
| NotYet | reading originalClassDecl | 2 | 2 | compiler lesson |  | True |
| NotYet | reading otherAccessor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading override | 2 | 2 | compiler lesson |  | True |
| NotYet | reading ownKey | 2 | 2 | compiler lesson |  | True |
| NotYet | reading paramSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parentSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parenthesizerRule | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parsedCommandLine | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pathAndExtension | 2 | 2 | compiler lesson |  | True |
| NotYet | reading patterns | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pendingDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading possiblyOutOfBoundsType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pragma | 2 | 2 | compiler lesson |  | True |
| NotYet | reading predicate | 2 | 2 | compiler lesson |  | True |
| NotYet | reading primaryTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading programDiagnosticsInFile | 2 | 2 | compiler lesson |  | True |
| NotYet | reading propName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading questionDotToken | 2 | 2 | compiler lesson |  | True |
| NotYet | reading real | 2 | 2 | compiler lesson |  | True |
| NotYet | reading reducedTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading relatedInfo | 2 | 2 | compiler lesson |  | True |
| NotYet | reading relativeFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading remove | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolutions | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedProject | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedTypeSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading restIdent | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnMethod | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnStatement | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnTypeNode | 2 | 2 | compiler lesson |  | True |
| NotYet | reading right | 2 | 2 | compiler lesson |  | True |
| NotYet | reading s | 2 | 2 | compiler lesson |  | True |
| NotYet | reading savedPreserveSourceNewlines | 2 | 2 | compiler lesson |  | True |
| NotYet | reading semanticDiagnostics | 2 | 2 | compiler lesson |  | True |
| NotYet | reading shorterParamType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading shortest | 2 | 2 | compiler lesson |  | True |
| NotYet | reading signatureDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceEmitHelpers | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFilePath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFileWithAddedExtension | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFiles | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceRoot | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceSymbolFile | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading spread | 2 | 2 | compiler lesson |  | True |
| NotYet | reading str | 2 | 2 | compiler lesson |  | True |
| NotYet | reading substitute | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetReturnType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading terminalWidth | 2 | 2 | compiler lesson |  | True |
| NotYet | reading testedSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading thisContainer | 2 | 2 | compiler lesson |  | True |
| NotYet | reading timerToUpdateChildWatches | 2 | 2 | compiler lesson |  | True |
| NotYet | reading toComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading toWatch | 2 | 2 | compiler lesson |  | True |
| NotYet | reading transform | 2 | 2 | compiler lesson |  | True |
| NotYet | reading trueType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading ts | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeLiteralNode | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeOnlyDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeOnlyDeclarationIsExportStar | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeParameter | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeParams | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typesVersions | 2 | 2 | compiler lesson |  | True |
| NotYet | reading valueSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading valueType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading varDecl | 2 | 2 | compiler lesson |  | True |
| NotYet | reading variable | 2 | 2 | compiler lesson |  | True |
| NotYet | reading versionPaths | 2 | 2 | compiler lesson |  | True |
| NotYet | reading visibleDefaultBinding | 2 | 2 | compiler lesson |  | True |
| NotYet | reading visitedAccessorName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading widened | 2 | 2 | compiler lesson |  | True |
| NotYet | reading yieldedType | 2 | 2 | compiler lesson |  | True |
| NotYet | storing true &#124; Node &#124; undefined in a field | 2 | 2 | compiler lesson |  | False |
| NotYet | .length on a value | 1 | 1 | compiler lesson |  | False |
| NotYet | ?. to a number, which would be number &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | JSON.stringify a union containing containers without runtime element metadata | 1 | 1 | compiler lesson |  | False |
| NotYet | Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a number &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number and a union of differently held members | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a string | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a union of differently held members | 1 | 1 | compiler lesson |  | False |
| NotYet | a ClassExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a DeleteExpression as a statement | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of CompilerOptionsValue | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of ResolvedConfigFilePath | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of T | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of string &#124; number | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map whose key and value types aren't known | 1 | 1 | compiler lesson |  | False |
| NotYet | a PostfixUnaryExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 | 1 | compiler lesson |  | False |
| NotYet | a YieldExpression as a statement | 1 | 1 | compiler lesson |  | False |
| NotYet | a base that isn't a declared class | 1 | 1 | compiler lesson |  | False |
| NotYet | a call returning T | 1 | 1 | compiler lesson |  | False |
| NotYet | a case that isn't a constant | 1 | 1 | compiler lesson |  | False |
| NotYet | a destructured parameter beside a parameter with a default | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "list" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type 0 &#124; boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type boolean &#124; (() => boolean) | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type string &#124; false | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type string &#124; false &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ((...args: A) => R) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning () => T | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (ConstructorDeclaration & { body: Block; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (arg: A) => T | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning A | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning AnyValidImportOrReExport | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning AnyValidImportOrReExport &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning BuildInvalidedProject<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning BuildInvalidedProject<T> &#124; UpdateOutputFileStampsProject | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning CapturedThis | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassExpression &#124; ImmediatelyInvokedArrowFunction | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassThisAssignmentBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning Declaration & HasModifiers | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning EvaluatorResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning HasJSDoc &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ImmediatelyInvokedArrowFunction | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning InferenceContext &#124; (T & undefined) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning InvalidatedProject<T> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MemberName &#124; (Expression & (NumericLiteral &#124; StringLiteralLike)) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MissingList<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ModeAwareCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ModifierToken<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ModuleOrTypeReferenceResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MultiMap<K, V> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning NodeArray<NonNullable<T>> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning NonRelativeNameResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning PerDirectoryResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning R | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ResolvedConfigFilePath | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ReusableDiagnosticMessageChain | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning SolutionBuilder<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning SyntheticSuper | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; EmptyStatement &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; Identifier | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; NumericLiteral &#124; StringLiteral &#124; BooleanLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; StringLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; readonly T[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T1 & T2 | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TEntry &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TOut | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TOut &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TPrivateEntry &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TResult | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning Token<TKind> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TransformationResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TypeMapper &#124; (T & undefined) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TypeOnlyAliasDeclaration &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning U[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning U[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning V &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning V[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning VisitResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning WatchCompilerHostOfConfigFile<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning WatchCompilerHostOfFilesAndCompilerOptions<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning object &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly Resolution[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly U[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly U[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning string &#124; object | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning unknown | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; number &#124; Symbol | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { [P in K as `${P}`]?: T[]; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { modifiers: NodeArray<Modifier> &#124; undefined; referencedName: Expression &#124; undefined; name: PropertyName; initializersName: Identifier &#124; undefined; descriptorName: Identifier &#124; undefined; thisArg: Identifier &#124; undefined; extraInitializersName?: never; } &#124; ... | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { readonly min: number; readonly max: number; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function value taking CreateSourceFileOptions &#124; ScriptTarget | 1 | 1 | compiler lesson |  | False |
| NotYet | a parameter that isn't a plain name | 1 | 1 | compiler lesson |  | False |
| NotYet | a rest array of T[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a spread that adds a field the source doesn't have | 1 | 1 | compiler lesson |  | False |
| NotYet | a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type "" &#124; ResolvedConfigFileName &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (AmbientModuleDeclaration & { name: StringLiteral; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) &#124; (... & ... 1 more ... & BinaryExpression) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (ModuleDeclaration & { name: StringLiteral; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (VariableDeclaration & { name: Identifier; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (element: Node) => "quit" &#124; boolean | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (s: string) => void | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AccessExpression &#124; RequireOrImportCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AccessorDeclaration & { readonly name: BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AliasDeclarationNode | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ArrowFunction &#124; BinaryExpression &#124; BindingElement &#124; Block &#124; BreakStatement &#124; CallSignatureDeclaration &#124; ... 64 more ... &#124; EndOfFileToken | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindablePropertyAssignmentExpression &#124; PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindableStaticAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CallExpression &#124; BindableStaticAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Child | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassElement &#124; ParameterPropertyDeclaration | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassNamedEvaluationHelperBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CustomTransformerFactory &#124; TransformerFactory<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type EntityNameExpression &#124; (LeftHandSideExpression & BindableStaticNameExpression) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type IncludeTypeSpaceImports | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type JSDocImportTag &#124; CanHaveModuleSpecifier | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<Path, ModeAwareCache<T>> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<string, [K, V[]]> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type MapLike<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NamedDeclaration & { name: DeclarationName; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral &#124; StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NodeArray<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NonNullable<K> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NonNullable<U> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrimitiveLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrivateEnvironment<TData, TEntry> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrivateIdentifierInExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAccessExpression &#124; SyntheticSuper | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyDeclaration &#124; ParameterPropertyDeclaration | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ReplaceableIndexedAccessType | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type RequireOrImportCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ResolvedConfigFileName &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Set<K> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type SortedArray<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Source | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type SourceFileOrString | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type T &#124; T[] &#124; readonly T[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type T1 | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TData | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TEntry | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TKind | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TKind &#124; Token<TKind> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TNode | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TransformedSuperCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type UnaryExpression & (BigIntLiteral &#124; NumericLiteral) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type UnaryExpression & NumericLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type V | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclaration & { name: Identifier; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclarationList & { _usingBrand: void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type boolean &#124; V &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly Child[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] &#124; readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly K[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type string &#124; null | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type string &#124; object &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of TState | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of object | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of unknown | 1 | 1 | compiler lesson |  | False |
| NotYet | destructuring a string | 1 | 1 | compiler lesson |  | False |
| NotYet | for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | compiler lesson |  | False |
| NotYet | incrementing a NonNullExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | new Map from something that isn't [key, value] pairs | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToMultiMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToNumericMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of createBinaryExpressionTrampoline with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of createToken with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachAncestorDirectory with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachLeadingCommentRange with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachTrailingCommentRange with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of getOriginalNode with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of mutateMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of mutateMapSkippingNewValues with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of resolveTypeReferenceDirectiveNamesReusingOldState with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of setSerializerContextAnd with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of sortAndDeduplicate with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 2 of arrayFrom with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 3 of group with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | push on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 | 1 | compiler lesson |  | False |
| NotYet | reading aParts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading absoluteSourceFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading accessibleSymbolsFromExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading accessorType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading actualFileName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading add | 1 | 1 | compiler lesson |  | True |
| NotYet | reading affectedSourceFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading afterImportPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading afterImportTagPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading aliasDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading allComponentComputedNamesSerializable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alreadyTransformed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alternateResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alternateResultMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ambientModuleDeclare | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ancestor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ancestorFacts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading annotatedNodes | 1 | 1 | compiler lesson |  | True |
| NotYet | reading applicableByArity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading arrayType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading arrowExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading awaitToken | 1 | 1 | compiler lesson |  | True |
| NotYet | reading awaitedLeftType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading base64SourceMapText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseCandidates | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseIndexedAccess | 1 | 1 | compiler lesson |  | True |
| NotYet | reading bindingList | 1 | 1 | compiler lesson |  | True |
| NotYet | reading blockedByExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading build | 1 | 1 | compiler lesson |  | True |
| NotYet | reading buildArray | 1 | 1 | compiler lesson |  | True |
| NotYet | reading buildOrderFromState | 1 | 1 | compiler lesson |  | True |
| NotYet | reading builderProgram | 1 | 1 | compiler lesson |  | True |
| NotYet | reading byteOrderMarkIndicator | 1 | 1 | compiler lesson |  | True |
| NotYet | reading callArgument | 1 | 1 | compiler lesson |  | True |
| NotYet | reading callTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canSuggestTypeof | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canUseBreakOrContinue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading candidateExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canonical | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canonicalFileName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading changed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading check | 1 | 1 | compiler lesson |  | True |
| NotYet | reading checkBody | 1 | 1 | compiler lesson |  | True |
| NotYet | reading childComponents | 1 | 1 | compiler lesson |  | True |
| NotYet | reading classDeclarations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading className | 1 | 1 | compiler lesson |  | True |
| NotYet | reading classThis | 1 | 1 | compiler lesson |  | True |
| NotYet | reading cloned | 1 | 1 | compiler lesson |  | True |
| NotYet | reading closingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading collidingSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commentEnd | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commentText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commonResolved | 1 | 1 | compiler lesson |  | True |
| NotYet | reading compareResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading comparer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading components | 1 | 1 | compiler lesson |  | True |
| NotYet | reading conditionPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading configFileText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constantValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constraintNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorFunction | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorLikeName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containerObjectType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containingDirectoryPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containingSourceFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading content | 1 | 1 | compiler lesson |  | True |
| NotYet | reading contextType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading create | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentDetachedCommentInfo | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentDirectory | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentGlobalDiagnostics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentNamespace | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentWriterIndentSpacing | 1 | 1 | compiler lesson |  | True |
| NotYet | reading customTransformer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading d | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationTransform | 1 | 1 | compiler lesson |  | True |
| NotYet | reading decorators | 1 | 1 | compiler lesson |  | True |
| NotYet | reading defaultDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading deprecatedTag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading diagnosticMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading diags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dir | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dirPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading directlyRelated | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminant | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminantCombinations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminantType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dist | 1 | 1 | compiler lesson |  | True |
| NotYet | reading doneType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading downleveledImport | 1 | 1 | compiler lesson |  | True |
| NotYet | reading effectiveExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading elementType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitAsSingleStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitExplicitInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitFileKey | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitSourceMaps | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitTrailingComma | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emittedAsTopLevel | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emittedCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enclosingBlockScopeContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enclosingContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading end | 1 | 1 | compiler lesson |  | True |
| NotYet | reading entry | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enumResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enumStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading envVarStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errorMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errorSpan | 1 | 1 | compiler lesson |  | True |
| NotYet | reading everyClauseChecks | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exclamationToken | 1 | 1 | compiler lesson |  | True |
| NotYet | reading excludeRe | 1 | 1 | compiler lesson |  | True |
| NotYet | reading excludeSpecs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingPending | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingProp | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exitStatus | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expando | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportClause | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportedNamesStorageRef | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expressionPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expressionResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extended | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extendedConstraint | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extendsType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extensionless | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extensions | 1 | 1 | compiler lesson |  | True |
| NotYet | reading externalHelpersModuleName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading externalHelpersModuleReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading facts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading failed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading failedSignatureDeclarations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading falseSubtype | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fileSystemEntryExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fileToErrorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filesForEmit | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filesInError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filteredTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | reading finished | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstAccessor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstInterfaceDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstNonzeroSegment | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstThisParameterOfUnionSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstVariableMatch | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fixed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading following | 1 | 1 | compiler lesson |  | True |
| NotYet | reading forInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading forStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading format | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatedName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatorFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatorInstantiation | 1 | 1 | compiler lesson |  | True |
| NotYet | reading genericDiag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading getAccessorType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading getFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading globalCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading grandParent | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasEmptyObject | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasExistingReasonToReportErrorOn | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasJSDocFunctionType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasPrivateModifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasSyntheticDefault | 1 | 1 | compiler lesson |  | True |
| NotYet | reading headerPadding | 1 | 1 | compiler lesson |  | True |
| NotYet | reading helpers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hostSourceFileInfo | 1 | 1 | compiler lesson |  | True |
| NotYet | reading i | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iife | 1 | 1 | compiler lesson |  | True |
| NotYet | reading implDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading impliedNodeFormat | 1 | 1 | compiler lesson |  | True |
| NotYet | reading importClause | 1 | 1 | compiler lesson |  | True |
| NotYet | reading importDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading imports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inAmbientContextOrInterface | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inTupleContext | 1 | 1 | compiler lesson |  | True |
| NotYet | reading includeRe | 1 | 1 | compiler lesson |  | True |
| NotYet | reading includeSpecs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading indexedAccessType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inferences | 1 | 1 | compiler lesson |  | True |
| NotYet | reading init | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initialLocationForSecondaryLookup | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializerStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializerWithoutParens | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializersName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inlinable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading instantiatedTemplateType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading instantiations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading internalFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading intrinsicAttribs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading intrinsics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading invalidElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading invokedExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isAnonymous | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isArrowFunctionInJsx | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCallToReadHelper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCallbackTag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCapturedInFunction | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isClassWithConstructorReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isConfigIdentical | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isDerivedClass | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isDosStyle | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isEitherEnum | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isEmpty | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isExternalImportAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isIllegalExportDefaultInCJS | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isInExternalModule | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isLeftNaN | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isMarkdownOrJSDocLink | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isOptional | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isOverload | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isPropertyName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isReservedWord | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isRest | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isVoidPromiseError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading issuedDiagnostic | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iterator | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iteratorValueStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsDoc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsDocType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsxFragPragma | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsxSpecific | 1 | 1 | compiler lesson |  | True |
| NotYet | reading keyAttr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading labeledElementDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastChild | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastJSDocParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastPart | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftSpread | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftThisArg | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftmost | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftmostExpressionKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lineText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading linesAfterDot | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literal | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literalText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literals | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localCheckDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localIndexDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mainExport | 1 | 1 | compiler lesson |  | True |
| NotYet | reading major | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mappedSource | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mapper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mappings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading markAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading matching | 1 | 1 | compiler lesson |  | True |
| NotYet | reading max | 1 | 1 | compiler lesson |  | True |
| NotYet | reading maxLength | 1 | 1 | compiler lesson |  | True |
| NotYet | reading maxNonRestParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading meaning | 1 | 1 | compiler lesson |  | True |
| NotYet | reading member | 1 | 1 | compiler lesson |  | True |
| NotYet | reading memberProps | 1 | 1 | compiler lesson |  | True |
| NotYet | reading merged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading messageChain | 1 | 1 | compiler lesson |  | True |
| NotYet | reading metadataReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading minArgumentCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading missingNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mod | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleBlock | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleFileToTry | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moreThanOneRealChildren | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nameStr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nameText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading namedBindings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needCheckInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needSyncEval | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsModifierPreservingWrapper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsOutParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading negative | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newItem | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newParams | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading next | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeConstructors | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeContextFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeModulesAtTypesExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading normalizedElements | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ns | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nullishSemantics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading numParameters | 1 | 1 | compiler lesson |  | True |
| NotYet | reading objectLiterals | 1 | 1 | compiler lesson |  | True |
| NotYet | reading objectProperties | 1 | 1 | compiler lesson |  | True |
| NotYet | reading offset | 1 | 1 | compiler lesson |  | True |
| NotYet | reading oldSignature | 1 | 1 | compiler lesson |  | True |
| NotYet | reading oldSourceFiles | 1 | 1 | compiler lesson |  | True |
| NotYet | reading onlyRecordFailuresForIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading onlyRecordFailuresForPackageFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading openBracketPosition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading operandPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optional | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionalDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionsOfCurCategory | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionsType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalClass | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalModuleSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalReadFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading other | 1 | 1 | compiler lesson |  | True |
| NotYet | reading otherFiles | 1 | 1 | compiler lesson |  | True |
| NotYet | reading otherOption | 1 | 1 | compiler lesson |  | True |
| NotYet | reading outputDir | 1 | 1 | compiler lesson |  | True |
| NotYet | reading overloadSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ownKeys | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ownOutputFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageFileResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageJsonMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageRootPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading param | 1 | 1 | compiler lesson |  | True |
| NotYet | reading paramCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parameterIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parameterNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parametersWithPropertyAssignments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parentComponents | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parentDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pathList | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pattern | 1 | 1 | compiler lesson |  | True |
| NotYet | reading peerDependencies | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pendingKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pollScheduled | 1 | 1 | compiler lesson |  | True |
| NotYet | reading possibleOption | 1 | 1 | compiler lesson |  | True |
| NotYet | reading postSuper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading precedingLineBreak | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prerelease | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prereleaseArray | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prevNodeIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading previousDuration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading previousGlobalDiagnostics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading primaryDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading program | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prologueStatementCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading promise | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propContext | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propertyAssignmentType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading proto | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prototype | 1 | 1 | compiler lesson |  | True |
| NotYet | reading qualifiedName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reachable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading react | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reactExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading readExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading readFileWithCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading realDeclarationPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading recursionIdentity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading recursiveInnerModule | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reducedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading referencedMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading relativePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading relevantTypeParameterConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading remainingPaths | 1 | 1 | compiler lesson |  | True |
| NotYet | reading removeUndefined | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rename | 1 | 1 | compiler lesson |  | True |
| NotYet | reading requiresAddingUndefined | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolutionsChanged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedFromFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedMethodReturnType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedTypeReferenceDirective | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedValueSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading restParameterSymbols | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resultType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading returnOrPromisedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rightExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rootExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rootSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading runtimeImportSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading savedInStrictMode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading seenSymbols | 1 | 1 | compiler lesson |  | True |
| NotYet | reading segments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading semanticDiagnosticsPerFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading separatingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading serializedName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setProp | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setReadFileCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setterModifiers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sharedLength | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldConvertCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldEmitDetachedComment | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldEmitDotDot | 1 | 1 | compiler lesson |  | True |
| NotYet | reading signatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading signaturesWithCorrectTypeArgumentArity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading singleLine | 1 | 1 | compiler lesson |  | True |
| NotYet | reading skipBindingPatterns | 1 | 1 | compiler lesson |  | True |
| NotYet | reading skipped | 1 | 1 | compiler lesson |  | True |
| NotYet | reading snippetElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sortedIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceFileNoExtension | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceIsJSConstructor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapRange | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapUrlPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMappings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceSignature | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading space | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spacesToEmit | 1 | 1 | compiler lesson |  | True |
| NotYet | reading specifierType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading specifiers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spreadElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spreadIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading startPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading startsOnNewLine | 1 | 1 | compiler lesson |  | True |
| NotYet | reading stateVariable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading statementExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading statementOffset | 1 | 1 | compiler lesson |  | True |
| NotYet | reading staticBlocks | 1 | 1 | compiler lesson |  | True |
| NotYet | reading strName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading subsequentNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading substituteConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superCallShouldBeRootLevel | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading system | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tagExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetCheckType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetDeclarationKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetHasStringIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetIsJSConstructor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetReturn | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetSymbolFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tempVar | 1 | 1 | compiler lesson |  | True |
| NotYet | reading templateType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading templates | 1 | 1 | compiler lesson |  | True |
| NotYet | reading testedNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisAccess | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisNodeOrAnySubNodesHasError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisParameters | 1 | 1 | compiler lesson |  | True |
| NotYet | reading timerToInvalidateFailedLookupResolutions | 1 | 1 | compiler lesson |  | True |
| NotYet | reading timerToUpdateProgram | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tokenSourceMapRanges | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tokenText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trailingComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trailingNewlines | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tripleSlash | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trueCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tryStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tsPriority | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tsconfigTime | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tupleTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeArgument | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeOfArrayLiteral | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeOfObjectLiteral | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeParameterCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typePredicateVariable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeReferenceResolutionsChanged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeRoots | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typescriptVersion | 1 | 1 | compiler lesson |  | True |
| NotYet | reading uniqueFilled | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unmatched | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwidenedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwrappedExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwrappedExprType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading updatedText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading usageMode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading useCaseSensitiveFileNames | 1 | 1 | compiler lesson |  | True |
| NotYet | reading useStrictDirective | 1 | 1 | compiler lesson |  | True |
| NotYet | reading val | 1 | 1 | compiler lesson |  | True |
| NotYet | reading valid | 1 | 1 | compiler lesson |  | True |
| NotYet | reading validatedFilesSpecBeforeSubstitution | 1 | 1 | compiler lesson |  | True |
| NotYet | reading valueDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading values | 1 | 1 | compiler lesson |  | True |
| NotYet | reading variableList | 1 | 1 | compiler lesson |  | True |
| NotYet | reading verbatimFromExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading version | 1 | 1 | compiler lesson |  | True |
| NotYet | reading visibilityResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading watchDirectoryKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading watchFileKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading widenedTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | slice with an index that isn't a number | 1 | 1 | compiler lesson |  | False |
| NotYet | spreading an array of other elements | 1 | 1 | compiler lesson |  | False |
| NotYet | storing any in a field | 1 | 1 | compiler lesson |  | False |
| NotYet | storing false &#124; Type in a field | 1 | 1 | compiler lesson |  | False |
| NotYet | storing string &#124; number in a field | 1 | 1 | compiler lesson |  | False |

### stage3 adaptation

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| Refused | a cast the runtime can't check | 3645 | 1736 | adaptation | internal/lower/cast_proof.go / lowering.castProof | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 414 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| NotYet | a value of type any | 91 | 91 | adaptation |  | False |
| Refused | a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | 90 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | 59 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile seen as SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 35 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TupleTypeReference seen as TypeReference, which can write GenericType where TupleType is read | 25 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| NotYet | an array of any | 23 | 23 | adaptation |  | False |
| Refused | a type predicate whose return is not proven (branch is not a trusted parameter check) | 20 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type FlowNode seen as FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 19 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (liftToBlock would lose its object, and this with it) | 18 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type ResolvedType seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 16 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 15 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | yield (generators) | 15 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| NotYet | a function returning any | 14 | 14 | adaptation |  | False |
| Refused | a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as string[], which can write string where never is read | 13 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | the void operator | 13 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a generator function | 12 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Node[] seen as unknown[], which can write unknown where Node is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Signature[], which can write Signature where never is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | in | 11 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Symbol &#124; undefined seen as Type &#124; undefined, which can write TypeFlags where SymbolFlags is read | 10 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Expression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (realpath would lose its object, and this with it) | 8 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on type) | 8 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type never[] seen as Type[], which can write Type where never is read | 8 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowNode seen as FlowNode[] &#124; FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (readFile would lose its object, and this with it) | 6 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on n) | 6 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type Expression seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type FlowNode &#124; undefined seen as false &#124; FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeNode &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| NotYet | a VoidExpression as a statement | 5 | 5 | adaptation |  | False |
| Refused | a method read as a value (trace would lose its object, and this with it) | 5 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where undefined is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowArrayMutation &#124; FlowAssignment seen as FlowNode, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 4 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol &#124; undefined seen as Declaration &#124; undefined, which can write number &#124; undefined where number is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking readonly Modifier[] &#124; undefined seen as one taking readonly ModifierLike[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getSourceFile would lose its object, and this with it) | 4 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Declaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type NodeArray<Statement> seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as Node[], which can write Node where never is read | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (trackSymbol would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on d) | 3 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type Diagnostic seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GenericType &#124; ResolvedType seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Signature &#124; undefined seen as Type &#124; undefined, which can write TypeFlags where SignatureFlags is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Symbol &#124; undefined, which can write SymbolFlags where TypeFlags is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as TypeNode &#124; undefined, which can write number &#124; undefined where number is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Declaration[], which can write Declaration where never is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking ((node: Node) => VisitResult<Node>) &#124; undefined seen as one taking ((node: Node) => VisitResult<Node &#124; undefined>) &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking JSDocTypeExpression &#124; undefined seen as one taking JSDocTypeExpression &#124; JSDocTypeLiteral &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Node seen as one taking [node: Node] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Visitor seen as one taking Visitor<TIn, Node &#124; undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type (identifierOrPrivateName: Identifier &#124; PrivateIdentifier) => string seen as (name: GeneratedIdentifier &#124; GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BindingElement[] seen as unknown[], which can write unknown where BindingElement is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as Expression &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassStaticBlockDeclaration &#124; PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type DiagnosticWithLocation &#124; undefined seen as Diagnostic &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowCall seen as FlowNode, which can write BinaryExpression &#124; CallExpression where CallExpression is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type MethodDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as NodeLinks &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ParameterDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol seen as Declaration &#124; undefined, which can write number &#124; undefined where number is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type false &#124; FlowNode &#124; undefined seen as false &#124; FlowAssignment &#124; FlowLabel &#124; FlowReduceLabel &#124; FlowStart &#124; FlowSwitchClause &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (Identifier &#124; StringLiteral)[], which can write Identifier &#124; StringLiteral where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as VariableDeclaration[], which can write VariableDeclaration where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Ternary | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| NotYet | a call returning any | 1 | 1 | adaptation |  | False |
| Refused | a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken &#124; BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Expression seen as one taking Expression &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking NodeArray<T> seen as one taking NodeArray<T> &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> &#124; undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] &#124; ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: Node] seen as one taking BindingElement &#124; OmittedExpression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: Node] seen as one taking Declaration (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [typeParameters: readonly TypeParameterDeclaration[] &#124; undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] &#124; [modifiers: readonly Modifier[] &#124; undefined, typeParameters: ... &#124; undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking readonly T[] &#124; undefined seen as one taking readonly T[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking string seen as one taking string &#124; MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (cloneNode would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createComma would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createTemplateTail would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (getSymlinkCache would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on arg) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on func) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on m) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on t) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on tagName) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type (ClassDeclaration &#124; EnumDeclaration &#124; ExportAssignment &#124; ExportDeclaration &#124; FunctionDeclaration &#124; ... 5 more ... &#124; VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration &#124; EnumDeclaration &#124; ExportAssignment &#124; ExportDeclaration &#124; FunctionDeclaration &#124; ... 5 more ... &#124; VariableStatement is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined, prefix?: string &#124; GeneratedNamePart &#124; undefined, suffix?: string &#124; undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined, prefix?: string &#124; ... 1 more ... &#124; undefined, suffix?: string &#124; undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull &#124; undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName &#124; undefined, which can write string &#124; undefined where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType &#124; DeferredTypeReference seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType &#124; GenericType seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AutoAccessorPropertyDeclaration &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AwaitedTypeInstantiation &#124; Type seen as Type, which can write Symbol &#124; undefined where Symbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; JsxNamespacedName &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; JsxNamespacedName &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Block &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CallExpression &#124; undefined seen as NodeArray<Expression> &#124; undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Children seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassDeclaration &#124; ClassExpression &#124; InferTypeNode &#124; InterfaceDeclaration &#124; JSDocCallbackTag &#124; ... 4 more ... &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration[] seen as (ArrayTypeNode &#124; Declaration &#124; NodeWithTypeArguments &#124; TupleTypeNode)[], which can write ArrayTypeNode &#124; Declaration &#124; NodeWithTypeArguments &#124; TupleTypeNode where Declaration is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as Diagnostic &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] &#124; undefined seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type EntityNameExpression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type EqualsGreaterThanToken seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ExportDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedPrivateIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GetAccessorDeclaration &#124; MethodDeclaration &#124; PropertyAssignment &#124; SetAccessorDeclaration &#124; ShorthandPropertyAssignment &#124; SpreadAssignment &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Identifier &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type IndexInfo &#124; undefined seen as IndexSignatureDeclaration &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IntrinsicType[] &#124; undefined seen as TypeParameter[] &#124; undefined, which can write TypeParameter where IntrinsicType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IterationTypes[] seen as (IterationTypes &#124; undefined)[], which can write IterationTypes &#124; undefined where IterationTypes is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSignature &#124; SignatureDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeExpression &#124; undefined seen as Signature &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause &#124; VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause &#124; VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type MemberName &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration &#124; SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode &#124; undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode &#124; undefined; } can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as EmitNode &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NumberLiteralType seen as LiteralType, which can write string &#124; number &#124; PseudoBigInt where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type ParameterDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyAccessExpression &#124; SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyAccessExpression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyName &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ResolvedType &#124; TypeReference seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Set<__String> seen as Set<__String &#124; undefined>, which can write __String &#124; undefined where __String is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile seen as Mutable<ModuleDeclaration &#124; SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile seen as SourceFileLike &#124; undefined, which can write readonly number[] &#124; undefined where readonly number[] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile &#124; undefined seen as NodeLinks &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Statement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type StringLiteralType seen as LiteralType, which can write string &#124; number &#124; PseudoBigInt where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SuperExpression seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol seen as Type &#124; undefined, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as GenericType &#124; undefined, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as Node &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as SourceFile &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol[] seen as (Symbol &#124; undefined)[], which can write Symbol &#124; undefined where Symbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SyntheticSuper seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ClassDeclaration &#124; ClassExpression &#124; GetAccessorDeclaration &#124; MethodDeclaration &#124; ParameterDeclaration &#124; PropertyDeclaration &#124; SetAccessorDeclaration can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint MethodDeclaration &#124; MethodSignature &#124; PropertyAssignment &#124; PropertyDeclaration &#124; PropertySignature &#124; AccessorDeclaration can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint Node &#124; undefined can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type seen as TypeNode, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Expression &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as JSDocTypeExpression &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Signature &#124; undefined, which can write SignatureFlags where TypeFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeParameterDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeVariable[] seen as (Type &#124; undefined)[], which can write Type &#124; undefined where TypeVariable is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type[] seen as (Type &#124; undefined)[], which can write Type &#124; undefined where Type is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypedefTag)[], which can write JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypedefTag where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (readonly (Identifier &#124; StringLiteral)[])[], which can write readonly (Identifier &#124; StringLiteral)[] where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as BindingElement[], which can write BindingElement where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as SymbolTable[], which can write SymbolTable where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Symbol[] &#124; undefined, which can write Symbol where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type readonly TypeParameter[] &#124; undefined seen as TypeParameterDeclaration[] &#124; undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type string[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type { introducesError: boolean; node: Identifier &#124; PropertyAccessEntityNameExpression; sym?: never; } &#124; { introducesError: boolean; node: Identifier &#124; PropertyAccessEntityNameExpression; sym: Symbol &#124; undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier &#124; PropertyAccessEntityNameExpression is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type { noInferenceFallback?: boolean &#124; undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile &#124; undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node &#124; undefined where ModuleDeclaration is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven relation from NodeArray<BindingElement> &#124; NodeArray<Expression> &#124; NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 | 0 | adaptation | internal/lower/proven_relations.go / lowering.provenRelation | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ElementFlags.Optional | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot NodeFlags.Const | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature &#124; SyntaxKind.MethodDeclaration &#124; SyntaxKind.Constructor &#124; SyntaxKind.GetAccessor &#124; SyntaxKind.SetAccessor &#124; SyntaxKind.CallSignature &#124; SyntaxKind.ConstructSignature &#124; ... 6 more ... &#124; SyntaxKind.JSDocFunctionType | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot VarianceFlags.Contravariant | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |

### UNOWNED

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| NotYet | a function inside a function (a closure) | 3238 | 3238 | not in pinned table |  | False |
| Refused | an object refinement using an open numeric enum as a literal tag | 2776 | 0 | not in pinned table |  | False |
| NotYet | a value of type __String | 174 | 174 | not in pinned table |  | False |
| Refused | var | 168 | 168 | not in pinned table |  | False |
| NotYet | a value of type Path | 118 | 118 | not in pinned table |  | False |
| Refused | &#124;&#124;= | 105 | 0 | not in pinned table |  | False |
| NotYet | an array of never | 79 | 79 | not in pinned table |  | False |
| NotYet | a BinaryExpression as a statement | 72 | 72 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (there is no body proving this parameter) | 71 | 0 | not in pinned table |  | False |
| Refused | optional property id in Node absent from structural source never, which can hide fields | 55 | 0 | not in pinned table |  | False |
| Refused | the comma operator | 47 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<JsonObjectExpressionStatement>, which can hide fields | 26 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 24 | 11 | not in pinned table |  | False |
| NotYet | a value of type __String &#124; undefined | 23 | 23 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeParameterDeclaration>, which can hide fields | 23 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 | 0 | not in pinned table |  | False |
| NotYet | reading context | 21 | 21 | not in pinned table |  | True |
| Refused | a method in object destructuring | 21 | 21 | not in pinned table |  | False |
| Refused | optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 21 | 0 | not in pinned table |  | False |
| NotYet | a call to a PropertyAccessExpression | 20 | 20 | not in pinned table |  | False |
| NotYet | a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 17 | 17 | not in pinned table |  | False |
| Refused | optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 17 | 0 | not in pinned table |  | False |
| NotYet | a function returning __String | 16 | 16 | not in pinned table |  | False |
| NotYet | a function returning __String &#124; undefined | 16 | 16 | not in pinned table |  | False |
| NotYet | a boolean &#124; undefined variable a function value captures | 14 | 14 | not in pinned table |  | False |
| Refused | a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 0 | not in pinned table |  | False |
| NotYet | a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 13 | 13 | not in pinned table |  | False |
| NotYet | a function returning Path | 12 | 12 | not in pinned table |  | False |
| Refused | a definite assignment assertion ! | 12 | 0 | not in pinned table |  | False |
| Refused | a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 12 | 0 | not in pinned table |  | False |
| Refused | a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of any would replace | 11 | 0 | not in pinned table |  | False |
| Refused | an index signature | 11 | 0 | not in pinned table |  | False |
| NotYet | a field of type string &#124; DiagnosticMessageChain | 10 | 10 | not in pinned table |  | False |
| NotYet | an enum inside a function or block; declare it at module scope | 10 | 0 | not in pinned table |  | False |
| Refused | a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program &#124; undefined where Program is read | 10 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, never[]> seen as Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>, which can write string[] where never[] is read | 10 | 0 | not in pinned table |  | False |
| Refused | optional property id in BigIntLiteral absent from structural source never, which can hide fields | 10 | 0 | not in pinned table |  | False |
| NotYet | a field of type true &#124; Node &#124; undefined | 9 | 9 | not in pinned table |  | False |
| Refused | a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] &#124; readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 | 0 | not in pinned table |  | False |
| NotYet | reading origin | 7 | 7 | not in pinned table |  | True |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 7 | 0 | not in pinned table |  | False |
| Refused | optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 7 | 0 | not in pinned table |  | False |
| Refused | optional property id in Identifier absent from structural source never, which can hide fields | 7 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source IntrinsicType, which can hide fields | 7 | 0 | not in pinned table |  | False |
| NotYet | an overloaded function as a value | 6 | 6 | not in pinned table |  | False |
| NotYet | reading lateSymbol | 6 | 6 | not in pinned table |  | True |
| NotYet | reading params | 6 | 6 | not in pinned table |  | True |
| Refused | a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (fileExists would lose its object, and this with it) | 6 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 | 0 | not in pinned table |  | False |
| Refused | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 | 0 | not in pinned table |  | False |
| Refused | optional property id in Expression absent from structural source never, which can hide fields | 6 | 0 | not in pinned table |  | False |
| NotYet | a class method through a view that erases its prototype origin | 5 | 5 | not in pinned table |  | False |
| NotYet | reading containingDirectory | 5 | 5 | not in pinned table |  | True |
| NotYet | reading fakespace | 5 | 5 | not in pinned table |  | True |
| Refused | a method read as a value (directoryExists would lose its object, and this with it) | 5 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getCanonicalFileName would lose its object, and this with it) | 5 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerOptionsValue seen as TsConfigSourceFile &#124; CompilerOptionsValue, which can write string &#124; number where string is read | 5 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as CompilerOptionsValue, which can write string &#124; number where string is read | 5 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CompilerOptions &#124; BuilderFileEmit | 5 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Declaration> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<HeritageClause>, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] &#124; undefined; }, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 | 0 | not in pinned table |  | False |
| NotYet | a literal method through a view that erases its receiver | 4 | 4 | not in pinned table |  | False |
| NotYet | a namespace object used as a value; use qualified members or named module imports | 4 | 0 | not in pinned table |  | False |
| NotYet | a value of type string &#124; (void & { __escapedIdentifier: void; }) &#124; (string & { __escapedIdentifier: void; }) | 4 | 4 | not in pinned table |  | False |
| NotYet | reading cachedChain | 4 | 4 | not in pinned table |  | True |
| NotYet | reading childrenTargetType | 4 | 4 | not in pinned table |  | True |
| NotYet | reading filePath | 4 | 4 | not in pinned table |  | True |
| NotYet | reading getCurrentDirectory | 4 | 4 | not in pinned table |  | True |
| NotYet | reading oldProgram | 4 | 4 | not in pinned table |  | True |
| NotYet | reading regularType | 4 | 4 | not in pinned table |  | True |
| NotYet | reading resolvedModule | 4 | 4 | not in pinned table |  | True |
| NotYet | reading restParamSymbol | 4 | 4 | not in pinned table |  | True |
| NotYet | reading reversed | 4 | 4 | not in pinned table |  | True |
| NotYet | reading typeSet | 4 | 4 | not in pinned table |  | True |
| Refused | Object.defineProperties | 4 | 4 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | 4 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ClassElement>, which can hide fields | 4 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<T>, which can hide fields | 4 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TypeElement>, which can hide fields | 4 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 4 | 0 | not in pinned table |  | False |
| NotYet | a union of differently held members variable a function value captures | 3 | 3 | not in pinned table |  | False |
| NotYet | a value of type Identifier &#124; __String | 3 | 3 | not in pinned table |  | False |
| NotYet | reading arrayLiteral | 3 | 3 | not in pinned table |  | True |
| NotYet | reading decorationStatements | 3 | 3 | not in pinned table |  | True |
| NotYet | reading inference | 3 | 3 | not in pinned table |  | True |
| NotYet | reading item | 3 | 3 | not in pinned table |  | True |
| NotYet | reading languageVersion | 3 | 3 | not in pinned table |  | True |
| NotYet | reading lines | 3 | 3 | not in pinned table |  | True |
| NotYet | reading metaPropertySymbol | 3 | 3 | not in pinned table |  | True |
| NotYet | reading newBaseType | 3 | 3 | not in pinned table |  | True |
| NotYet | reading newSymbol | 3 | 3 | not in pinned table |  | True |
| NotYet | reading regularNew | 3 | 3 | not in pinned table |  | True |
| NotYet | reading rootNames | 3 | 3 | not in pinned table |  | True |
| NotYet | reading sources | 3 | 3 | not in pinned table |  | True |
| NotYet | reading spreadType | 3 | 3 | not in pinned table |  | True |
| NotYet | reading symbolFromSymbolTable | 3 | 3 | not in pinned table |  | True |
| NotYet | reading tag | 3 | 3 | not in pinned table |  | True |
| NotYet | reading targetPropertySymbol | 3 | 3 | not in pinned table |  | True |
| Refused | a filter callback that doesn't return a boolean | 3 | 3 | not in pinned table |  | False |
| Refused | a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getDirectories would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (now would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on block) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type FileReference seen as { pos: number &#124; undefined; end: number &#124; undefined; }, which can write number &#124; undefined where number is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type FutureSourceFile &#124; SourceFile seen as Pick<SourceFile, "fileName" &#124; "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier &#124; TextRange seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier &#124; undefined seen as Identifier &#124; TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, string[] &#124; never[]> seen as Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>, which can write string where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile seen as SourceFile &#124; SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type TsConfigSourceFile &#124; CompilerOptionsValue seen as string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined, which can write string &#124; number where string is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type YieldExpression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as string[] &#124; never[], which can write string where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 | 0 | not in pinned table |  | False |
| Refused | optional property id in FalseLiteral absent from structural source never, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<EnumMember>, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateSpan>, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<ExpressionStatement>, which can hide fields | 3 | 0 | not in pinned table |  | False |
| NotYet | a field of type string &#124; number &#124; boolean &#124; DiagnosticMessage &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a function returning ModeAwareCacheKey | 2 | 2 | not in pinned table |  | False |
| NotYet | a function returning Path &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | 2 | 2 | not in pinned table |  | False |
| NotYet | a namespace member other than a function, type, initialized binding or nested namespace | 2 | 0 | not in pinned table |  | False |
| NotYet | a value of type Canonicalized | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFileId | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type Path &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type RedirectsCacheKey | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type never | 2 | 2 | not in pinned table |  | False |
| NotYet | an array of Path | 2 | 2 | not in pinned table |  | False |
| NotYet | incrementing an Identifier | 2 | 2 | not in pinned table |  | False |
| NotYet | reading _createProgramOptions | 2 | 2 | not in pinned table |  | True |
| NotYet | reading attributesType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading baseSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading checkFlags | 2 | 2 | not in pinned table |  | True |
| NotYet | reading checkTypeDeferred | 2 | 2 | not in pinned table |  | True |
| NotYet | reading childPropName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading childrenPropName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading classLikeDeclaration | 2 | 2 | not in pinned table |  | True |
| NotYet | reading containingCall | 2 | 2 | not in pinned table |  | True |
| NotYet | reading contextFile | 2 | 2 | not in pinned table |  | True |
| NotYet | reading derived | 2 | 2 | not in pinned table |  | True |
| NotYet | reading emitSkipped | 2 | 2 | not in pinned table |  | True |
| NotYet | reading escapedText | 2 | 2 | not in pinned table |  | True |
| NotYet | reading fileInfos | 2 | 2 | not in pinned table |  | True |
| NotYet | reading firstType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading flattenContext | 2 | 2 | not in pinned table |  | True |
| NotYet | reading freshType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading hasArguments | 2 | 2 | not in pinned table |  | True |
| NotYet | reading hooks | 2 | 2 | not in pinned table |  | True |
| NotYet | reading indent | 2 | 2 | not in pinned table |  | True |
| NotYet | reading indexSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading isCallExpression | 2 | 2 | not in pinned table |  | True |
| NotYet | reading isZero | 2 | 2 | not in pinned table |  | True |
| NotYet | reading iteratedType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading json | 2 | 2 | not in pinned table |  | True |
| NotYet | reading jsxFragmentFactoryName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading lanes | 2 | 2 | not in pinned table |  | True |
| NotYet | reading literalType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading newDirectory | 2 | 2 | not in pinned table |  | True |
| NotYet | reading newReturnType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading operandConstraint | 2 | 2 | not in pinned table |  | True |
| NotYet | reading origTypeParameter | 2 | 2 | not in pinned table |  | True |
| NotYet | reading originalCreateDirectory | 2 | 2 | not in pinned table |  | True |
| NotYet | reading ownMap | 2 | 2 | not in pinned table |  | True |
| NotYet | reading parseTreeNode | 2 | 2 | not in pinned table |  | True |
| NotYet | reading potentiallyUnusedIdentifiers | 2 | 2 | not in pinned table |  | True |
| NotYet | reading previous | 2 | 2 | not in pinned table |  | True |
| NotYet | reading prototypeProperty | 2 | 2 | not in pinned table |  | True |
| NotYet | reading sorted | 2 | 2 | not in pinned table |  | True |
| NotYet | reading sourceDiscriminantTypes | 2 | 2 | not in pinned table |  | True |
| NotYet | reading staticType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading targetProp | 2 | 2 | not in pinned table |  | True |
| NotYet | reading targets | 2 | 2 | not in pinned table |  | True |
| NotYet | reading templateArguments | 2 | 2 | not in pinned table |  | True |
| NotYet | reading transformed | 2 | 2 | not in pinned table |  | True |
| NotYet | reading typeLiteralSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading typeVariable | 2 | 2 | not in pinned table |  | True |
| NotYet | reading useDefineForClassFields | 2 | 2 | not in pinned table |  | True |
| NotYet | reading variances | 2 | 2 | not in pinned table |  | True |
| Refused | Object.create | 2 | 2 | not in pinned table |  | False |
| Refused | Object.defineProperty | 2 | 2 | not in pinned table |  | False |
| Refused | a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions &#124; ScriptTarget, onError?: ((message: string) => void) &#124; undefined, shouldCreateNewSourceFile?: boolean &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (clearTimeout would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getBuildInfo would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getCurrentDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (hasGlobalName would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setTimeout would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (substituteNode would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (toKey would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (watchDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (watchFile would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (writeFile would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view) | 2 | 2 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ArrayLiteralExpression &#124; AssignmentExpression<EqualsToken> &#124; BindingElement &#124; ElementAccessExpression &#124; ... 8 more ... &#124; VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Block seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type BreakStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] &#124; undefined where string[] is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ContinueStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticWithLocation[] &#124; undefined seen as DiagnosticRelatedInformation[] &#124; undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type FutureSourceFile &#124; SourceFile seen as Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | not in pinned table |  | False |
| Refused | a value of type Identifier[] &#124; undefined seen as ModuleExportName[] &#124; undefined, which can write ModuleExportName where Identifier is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | not in pinned table |  | False |
| Refused | a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean &#124; (() => boolean) &#124; undefined where (() => boolean) & (boolean &#124; (() => boolean) &#124; undefined) is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Node seen as Node &#124; TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Node &#124; TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ReturnStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile seen as EmitNode &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableStatement seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] seen as unknown[], which can write unknown where string is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof FlowFlags is read | 2 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 | 2 | not in pinned table |  | False |
| Refused | an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral &#124; SyntaxKind.TemplateHead &#124; SyntaxKind.TemplateMiddle &#124; SyntaxKind.TemplateTail | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword &#124; SyntaxKind.AssertKeyword | 2 | 0 | not in pinned table |  | False |
| Refused | delete | 2 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property id in NamedExports absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property packageJsonScope in Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" &#124; "impliedNodeFormat">, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Diagnostic> absent from structural source ArrayIterator<DiagnosticWithLocation>, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<JSDocTag>, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<TemplateLiteralTypeSpan>, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| NotYet | Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | not in pinned table |  | False |
| NotYet | a Map of RedirectsCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a call to an Identifier | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type "circularity" &#124; boolean | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type EmitSignature &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; Type &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; VersionPaths | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; VersionPaths &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning CacheWithRedirects<K, V> | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning IncrementalBuildInfoFileId | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning IncrementalBuildInfoFileIdListId | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning RedirectsCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning T &#124; T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning never | 1 | 1 | not in pinned table |  | False |
| NotYet | a mutable namespace export; use a module or export functions around private state | 1 | 0 | not in pinned table |  | False |
| NotYet | a namespace binding without a plain initialized name | 1 | 0 | not in pinned table |  | False |
| NotYet | a non-accessor method in an accessor literal | 1 | 1 | not in pinned table |  | False |
| NotYet | a reopened namespace or namespace merged with a runtime value; put the declarations in one namespace or use a module | 1 | 0 | not in pinned table |  | False |
| NotYet | a value of type ElementWithComputedPropertyName | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type EqualityComparer<T> &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type ExportDeclaration & { readonly exportClause: NamedExports; } | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type Identifier &#124; PrivateIdentifier &#124; __String | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFileIdListId | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFilePendingEmit | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type Map<string, SingleFileWatcher<T>> | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type ModeAwareCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type NodeArray<T> &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type PathPathComponents | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type RedirectsCacheKey &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type T &#124; T[] | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type T &#124; null &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type T &#124; readonly T[] | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type TIn &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type __String & string | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type readonly (readonly T[])[] | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type { getModifiedTime: (path: string) => Date &#124; undefined; } | 1 | 1 | not in pinned table |  | False |
| NotYet | an array of number &#124; "_" | 1 | 1 | not in pinned table |  | False |
| NotYet | apply without a dense argument literal (length, presence and argument representations must be proven) | 1 | 1 | not in pinned table |  | False |
| NotYet | overload 1 of or with a non-array implementation rest parameter | 1 | 1 | not in pinned table |  | False |
| NotYet | reading abstractSignatures | 1 | 1 | not in pinned table |  | True |
| NotYet | reading accessExpression | 1 | 1 | not in pinned table |  | True |
| NotYet | reading accessModifier | 1 | 1 | not in pinned table |  | True |
| NotYet | reading accessor | 1 | 1 | not in pinned table |  | True |
| NotYet | reading activeLabel | 1 | 1 | not in pinned table |  | True |
| NotYet | reading addUndefinedForParameter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading allSetOptions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading allowedEndings | 1 | 1 | not in pinned table |  | True |
| NotYet | reading annotationSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading augmentation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading b1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseConstraintMapper | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseConstructorType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseTypeNode | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseTypeNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading bases | 1 | 1 | not in pinned table |  | True |
| NotYet | reading best | 1 | 1 | not in pinned table |  | True |
| NotYet | reading bestMatch | 1 | 1 | not in pinned table |  | True |
| NotYet | reading binaryExpression | 1 | 1 | not in pinned table |  | True |
| NotYet | reading cachedDiagnostics | 1 | 1 | not in pinned table |  | True |
| NotYet | reading cachedResolvedSignatures | 1 | 1 | not in pinned table |  | True |
| NotYet | reading cachedResult | 1 | 1 | not in pinned table |  | True |
| NotYet | reading cachedTypes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading candidateDirectories | 1 | 1 | not in pinned table |  | True |
| NotYet | reading chain1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading checkTuples | 1 | 1 | not in pinned table |  | True |
| NotYet | reading childFieldType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading childrenNameType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading classInstanceType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading commonJSPropertyAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading compilerOptionsProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading computedPropertyName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading conditional | 1 | 1 | not in pinned table |  | True |
| NotYet | reading containingClassDecl | 1 | 1 | not in pinned table |  | True |
| NotYet | reading contextSpecifier | 1 | 1 | not in pinned table |  | True |
| NotYet | reading converters | 1 | 1 | not in pinned table |  | True |
| NotYet | reading createNodeArray | 1 | 1 | not in pinned table |  | True |
| NotYet | reading createProgramOptionsHost | 1 | 1 | not in pinned table |  | True |
| NotYet | reading currentOptions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading defaultExportSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading defaultReplaced | 1 | 1 | not in pinned table |  | True |
| NotYet | reading directory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading directoryStart | 1 | 1 | not in pinned table |  | True |
| NotYet | reading disposeScope | 1 | 1 | not in pinned table |  | True |
| NotYet | reading effectiveEnclosingContext | 1 | 1 | not in pinned table |  | True |
| NotYet | reading elementAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading emitSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading enter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading equivalentFileSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading errorBindingElement | 1 | 1 | not in pinned table |  | True |
| NotYet | reading excludedProperties | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exit | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exportDecl | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exportEqualsSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exportStars | 1 | 1 | not in pinned table |  | True |
| NotYet | reading expressions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ext | 1 | 1 | not in pinned table |  | True |
| NotYet | reading factory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading fakeSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading finalizeBoundary | 1 | 1 | not in pinned table |  | True |
| NotYet | reading forcedLookupLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading freshTypeParameter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading functionLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading globalTypingsCacheLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading grid | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasInstanceProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasSignatures | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasTransformableStatics | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ids | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ifStatement | 1 | 1 | not in pinned table |  | True |
| NotYet | reading immediateContainer | 1 | 1 | not in pinned table |  | True |
| NotYet | reading includeFileRegexes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading indexedType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading instantiated | 1 | 1 | not in pinned table |  | True |
| NotYet | reading instantiationExpressionType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading intrinsicType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isExportEquals | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isKnownProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isLengthPushOrUnshift | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isNonLocalFunctionSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading jsxChildrenPropertyName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading jsxFactoryNamespace | 1 | 1 | not in pinned table |  | True |
| NotYet | reading labeledElementDeclarations | 1 | 1 | not in pinned table |  | True |
| NotYet | reading laneCount | 1 | 1 | not in pinned table |  | True |
| NotYet | reading lastCommentLine | 1 | 1 | not in pinned table |  | True |
| NotYet | reading lastLeft | 1 | 1 | not in pinned table |  | True |
| NotYet | reading limitedConstraint | 1 | 1 | not in pinned table |  | True |
| NotYet | reading literalValue | 1 | 1 | not in pinned table |  | True |
| NotYet | reading localJsxNamespace | 1 | 1 | not in pinned table |  | True |
| NotYet | reading localProps | 1 | 1 | not in pinned table |  | True |
| NotYet | reading mapDirectory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading mappedTypeNode | 1 | 1 | not in pinned table |  | True |
| NotYet | reading missingPaths | 1 | 1 | not in pinned table |  | True |
| NotYet | reading newConstraint | 1 | 1 | not in pinned table |  | True |
| NotYet | reading noInferSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading noSupertypeReduction | 1 | 1 | not in pinned table |  | True |
| NotYet | reading numNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading o1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading oldEnclosing | 1 | 1 | not in pinned table |  | True |
| NotYet | reading oldTime | 1 | 1 | not in pinned table |  | True |
| NotYet | reading onProgramCreateComplete | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parameterSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parentDirectory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parenthesizerRules | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prevSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading printList | 1 | 1 | not in pinned table |  | True |
| NotYet | reading projectReferences | 1 | 1 | not in pinned table |  | True |
| NotYet | reading propertyAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypePropertyType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypeSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypeType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading quick | 1 | 1 | not in pinned table |  | True |
| NotYet | reading r | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rawName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rawSources | 1 | 1 | not in pinned table |  | True |
| NotYet | reading reexports | 1 | 1 | not in pinned table |  | True |
| NotYet | reading referencedFileName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading replacements | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolutionDiagnostic | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolvedModuleNames | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolver | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rest | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resultFromDts | 1 | 1 | not in pinned table |  | True |
| NotYet | reading results | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rootPathComponents | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rootResult | 1 | 1 | not in pinned table |  | True |
| NotYet | reading scanner | 1 | 1 | not in pinned table |  | True |
| NotYet | reading secondType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldResolveAlias | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldResolveFactoryReference | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformInitializers | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformInitializersUsingSet | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformThisInStaticInitializers | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldWriteNativeEvents | 1 | 1 | not in pinned table |  | True |
| NotYet | reading signatureNextType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceFileAbsolutePaths | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceOrigin | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceProp | 1 | 1 | not in pinned table |  | True |
| NotYet | reading specialPropertyAssignmentKind | 1 | 1 | not in pinned table |  | True |
| NotYet | reading stat | 1 | 1 | not in pinned table |  | True |
| NotYet | reading subcontext | 1 | 1 | not in pinned table |  | True |
| NotYet | reading suffix | 1 | 1 | not in pinned table |  | True |
| NotYet | reading swappedMode | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sym | 1 | 1 | not in pinned table |  | True |
| NotYet | reading symbolExport | 1 | 1 | not in pinned table |  | True |
| NotYet | reading symbols | 1 | 1 | not in pinned table |  | True |
| NotYet | reading syntacticBuilderResolver | 1 | 1 | not in pinned table |  | True |
| NotYet | reading syntheticArgsSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading targetOrigin | 1 | 1 | not in pinned table |  | True |
| NotYet | reading targetProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading tempSources | 1 | 1 | not in pinned table |  | True |
| NotYet | reading texts | 1 | 1 | not in pinned table |  | True |
| NotYet | reading thenFunction | 1 | 1 | not in pinned table |  | True |
| NotYet | reading thisParam | 1 | 1 | not in pinned table |  | True |
| NotYet | reading throwDiagnostic | 1 | 1 | not in pinned table |  | True |
| NotYet | reading tok | 1 | 1 | not in pinned table |  | True |
| NotYet | reading trampoline | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeClass | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeKey | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeKind | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading undefinedStrippedTarget | 1 | 1 | not in pinned table |  | True |
| NotYet | reading valueParam | 1 | 1 | not in pinned table |  | True |
| NotYet | reading variableDeclarator | 1 | 1 | not in pinned table |  | True |
| NotYet | replacing a namespace export; keep exported functions fixed and mutate private state through them | 1 | 0 | not in pinned table |  | False |
| NotYet | storing Path &#124; undefined in a field | 1 | 1 | not in pinned table |  | False |
| NotYet | typed array element type Uint16Array | 1 | 1 | not in pinned table |  | False |
| NotYet | var inside a namespace; use initialized private let or const | 1 | 0 | not in pinned table |  | False |
| Refused | &&= | 1 | 0 | not in pinned table |  | False |
| Refused | JSON.parse: its result's type can't be proven from the text | 1 | 1 | not in pinned table |  | False |
| Refused | Object.setPrototypeOf | 1 | 1 | not in pinned table |  | False |
| Refused | a constructor object escaping before static fields are initialized | 1 | 1 | not in pinned table |  | False |
| Refused | a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking NodeArray<TypeNode> &#124; undefined seen as one taking readonly TypeNode[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean &#124; undefined, options?: WatchOptions &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number &#124; undefined, options?: WatchOptions &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (add would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (base64decode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (base64encode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (clearScreen would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (compare would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLink would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (deleteFile would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emit would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (fill would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getAllDependencies would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getModifiedTime would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (log would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (onDiscoveredSymlink would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (releaseProgram would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (remove would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (repeat would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportTruncationError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setModifiedTime would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (toString would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a spread after the first field | 1 | 1 | not in pinned table |  | False |
| Refused | a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> &#124; undefined, which can write string &#124; undefined where string is read | 1 | 1 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on hostSourceFile) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (the body's true narrowing does not match null &#124; undefined) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (the predicate parameter is assigned) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ((level: LogLevel) => boolean) &#124; typeof log &#124; (() => AssertionLevel) &#124; ((level: AssertionLevel) => void) &#124; ((level: AssertionLevel) => boolean) &#124; ... 45 more ... &#124; ... seen as AnyFunction, which can write void where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression) &#124; undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (AbstractKeyword &#124; AccessorKeyword &#124; AsyncKeyword &#124; ConstKeyword &#124; DeclareKeyword &#124; Decorator &#124; ... 7 more ... &#124; StaticKeyword)[] &#124; undefined seen as Node[] &#124; undefined, which can write Node where AbstractKeyword &#124; AccessorKeyword &#124; AsyncKeyword &#124; ConstKeyword &#124; DeclareKeyword &#124; Decorator &#124; ... 7 more ... &#124; StaticKeyword is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> &#124; undefined, which can write string &#124; undefined where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName &#124; undefined; } &#124; undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic &#124; undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (value: TIn) => value is TOut seen as AnyFunction, which can write void where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; GeneratedIdentifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; GeneratedIdentifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BinaryExpression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BindingName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration &#124; FunctionDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback &#124; undefined where WriteFileCallback is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain &#124; { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... &#124; undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] &#124; undefined where DiagnosticMessageChain[] &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] &#124; undefined where DiagnosticMessageChain[] &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type EnumDeclaration &#124; ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ExportAssignment seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Expression &#124; GeneratedIdentifier seen as Expression &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Expression &#124; GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type FlowArrayMutation &#124; FlowAssignment &#124; FlowCall &#124; FlowCondition &#124; FlowLabel &#124; FlowReduceLabel &#124; FlowStart &#124; FlowUnreachable seen as FlowNode, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier seen as Expression &#124; GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier seen as GeneratedIdentifier &#124; Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Identifier &#124; PrivateIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as GeneratedIdentifier &#124; Identifier &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as Identifier &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as string &#124; GeneratedIdentifier &#124; Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier &#124; undefined seen as string &#124; ModuleExportName &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GetAccessorDeclaration &#124; SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type Identifier[] seen as ModuleExportName[] &#124; undefined, which can write ModuleExportName where Identifier is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ImportEqualsDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type InitializedVariableDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> &#124; undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, string[]> seen as InvokeMap, which can write true &#124; string[] where string[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, Set<string> &#124; undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> &#124; Set<ResolutionWithFailedLookupLocations> &#124; undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ImportsNotUsedAsValues is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, JsxEmit> seen as Map<string, string &#124; number>, which can write string &#124; number where JsxEmit is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ModuleDetectionKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ModuleDetectionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ModuleResolutionKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ModuleResolutionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, NewLineKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where NewLineKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, PollingWatchKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where PollingWatchKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, WatchDirectoryKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where WatchDirectoryKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, WatchFileKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where WatchFileKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, string> seen as Map<string, string &#124; number>, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type MethodDeclaration &#124; PropertyAssignment &#124; AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleExportName seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type Node seen as Node &#124; SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Node &#124; SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration &#124; PropertyDeclaration &#124; PropertySignature &#124; SignatureDeclaration seen as Mutable<ParameterDeclaration &#124; PropertyDeclaration &#124; PropertySignature &#124; SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration &#124; VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean &#124; (() => boolean) &#124; undefined where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Path[] seen as string[], which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type PropertyAssignment seen as Mutable<PropertyAssignment &#124; ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Readonly<BuilderState> &#124; undefined seen as BuilderState &#124; undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ReportFileInError[] seen as (ReportFileInError &#124; undefined)[], which can write ReportFileInError &#124; undefined where ReportFileInError is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ResolvedModuleFull &#124; undefined seen as { path: string; originalPath: string &#124; true; extension: string; packageId: PackageId &#124; undefined; resolvedUsingTsExtension: boolean &#124; undefined; } &#124; undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string &#124; undefined, which a write of string &#124; true would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } &#124; undefined; } &#124; undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved &#124; undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Set<Path> &#124; undefined seen as Set<string> &#124; undefined, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment &#124; ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile &#124; undefined seen as FileReasonToChainCache &#124; undefined, which can write DiagnosticMessageChain[] &#124; undefined where RedirectInfo &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement[] &#124; undefined seen as CaseClause[] &#124; undefined, which can write CaseClause where Statement is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type T[] seen as (T &#124; undefined)[], which can write T &#124; undefined where T is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ThrowStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration &#124; DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls &#124; undefined)[], which can write WatchedFileWithUnchangedPolls &#124; undefined where WatchedFileWithUnchangedPolls is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as (JSDoc &#124; JSDocTag)[], which can write JSDoc &#124; JSDocTag where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as (SourceMapRange &#124; undefined)[], which can write SourceMapRange &#124; undefined where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Expression[], which can write Expression where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as IntrinsicType[], which can write IntrinsicType where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as JsxAttributes[], which can write JsxAttributes where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ParameterDeclaration[], which can write ParameterDeclaration where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Path[], which can write Path where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as PotentiallyUnusedIdentifier[], which can write PotentiallyUnusedIdentifier where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as PropertyAssignment[], which can write PropertyAssignment where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as TransformerFactory<Bundle &#124; SourceFile>[], which can write TransformerFactory<Bundle &#124; SourceFile> where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] &#124; SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[][] seen as string[][], which can write string where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type number[] seen as string &#124; (string &#124; number)[] &#124; undefined, which can write string &#124; number where number is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly string[] &#124; undefined seen as RegExp[] &#124; undefined, which can write RegExp where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; GeneratedIdentifier &#124; Identifier seen as string &#124; ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined seen as string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; number &#124; boolean &#124; string[] &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> seen as CompilerOptionsValue, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as string &#124; (string &#124; number)[] &#124; undefined, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof CheckMode is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof EmitFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof GeneratedIdentifierFlags is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ModifierFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof NodeCheckFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof NodeFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ObjectFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof RelationComparisonResult is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ScriptKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SignatureCheckMode is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SignatureFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SnippetKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SymbolFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SyntaxKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TransformFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TypeFacts is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TypeFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; BuilderFileEmit.DtsErrors &#124; ... 6 more ... &#124; BuilderFileEmit.All; } seen as { affectedFile: Program &#124; SourceFile &#124; undefined; emitKind: BuilderFileEmit; }, which can write Program &#124; SourceFile &#124; undefined where SourceFile is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind &#124; undefined where ModuleResolutionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } &#124; { arguments: { name: string; }; range: CommentRange; } &#124; { arguments: { factory: string; }; range: CommentRange; } &#124; ... 5 more ... &#124; ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } &#124; { arguments: { name: string; }; range: CommentRange; } &#124; { arguments: { factory: string; }; range: CommentRange; } &#124; ... 5 more ... &#124; ...)[] &#124; ... 8 more ... &#124; ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] &#124; undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] &#124; undefined where never[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding &#124; undefined; value: string; }[], which can write ModuleSpecifierEnding &#124; undefined where ModuleSpecifierEnding is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] &#124; undefined; affectingLocations: string[] &#124; undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string &#124; number; minor: number; patch: number; prerelease: string &#124; readonly string[]; build: string &#124; readonly string[]; }, which can write string &#124; number where number is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { name: string &#124; undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string &#124; undefined; path: string; } is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { referencedName: StringLiteral; name: PropertyName; } &#124; { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression &#124; undefined; name: PropertyName &#124; undefined; }, which can write Expression &#124; undefined where StringLiteral is read | 1 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from InstantiableType &#124; UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.element.parent.name has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; BuilderFileEmit.DtsErrors &#124; BuilderFileEmit.DtsEmit &#124; ... 5 more ... &#124; BuilderFileEmit.All | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Connection | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword &#124; SyntaxKind.ImplementsKeyword | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword &#124; SyntaxKind.ReadonlyKeyword &#124; SyntaxKind.UniqueKeyword | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral &#124; SyntaxKind.BigIntLiteral &#124; SyntaxKind.StringLiteral &#124; SyntaxKind.JsxText &#124; SyntaxKind.JsxTextAllWhiteSpaces &#124; SyntaxKind.RegularExpressionLiteral &#124; SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown &#124; SyntaxKind.NumericLiteral &#124; SyntaxKind.BigIntLiteral &#124; SyntaxKind.StringLiteral &#124; SyntaxKind.JsxText &#124; SyntaxKind.RegularExpressionLiteral &#124; SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | arguments | 1 | 0 | not in pinned table |  | False |
| Refused | inherited library member compare read as an own field | 1 | 1 | not in pinned table |  | False |
| Refused | inherited library member repeat read as an own field | 1 | 1 | not in pinned table |  | False |
| Refused | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" &#124; "strict">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" &#124; "strict">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictBindCallApply">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictBuiltinIteratorReturn">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictFunctionTypes">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictNullChecks">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictPropertyInitialization">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "useUnknownInCatchVariables">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source ObjectType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property default in TypeParameter absent from structural source IndexedAccessType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in ComputedPropertyName absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source IntersectionType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source UnionType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property outputDts in ResolvedRefAndOutputDts absent from structural source ResolvedRefAndSource, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS &#124; ModuleKind.ESNext; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<ImportAttribute>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<WatchedFileWithUnchangedPolls &#124; undefined> absent from structural source ArrayIterator<WatchedFileWithUnchangedPolls>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in MapIterator<[string, WatchDirectoryFlags]> absent from structural source MapIterator<[string, WatchDirectoryFlags]>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[]; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property source in ResolvedRefAndSource absent from structural source ResolvedRefAndOutputDts, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property target in AnonymousType absent from structural source IntrinsicType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property typeArguments in NodeWithTypeArguments absent from structural source ArrayTypeNode, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string &#124; undefined; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string &#124; { value: string; pos: number; end: number; }; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | overload 1 of arrayFrom result U[] cannot be served by implementation result (T &#124; U)[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of convertOptionsFromJson result WatchOptions &#124; undefined cannot be served by implementation result CompilerOptions &#124; TypeAcquisition &#124; WatchOptions &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createImmediatelyInvokedArrowFunction result ImmediatelyInvokedArrowFunction cannot be served by implementation result Mutable<CallExpression> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createImmediatelyInvokedFunctionExpression result ImmediatelyInvokedFunctionExpression cannot be served by implementation result Mutable<CallExpression> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of ensureTrailingDirectorySeparator result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of filter result U[] cannot be served by implementation result readonly T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of find result U &#124; undefined cannot be served by implementation result T &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of findLast result U &#124; undefined cannot be served by implementation result T &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getDirectoryPath result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getGlobalBuiltinTypes result ObjectType[] cannot be served by implementation result Type[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSuperContainer result SuperContainer &#124; undefined cannot be served by implementation result SuperContainerOrFunctions &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction &#124; ComputedPropertyName &#124; ThisContainer | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> &#124; Extract<...> cannot be served by implementation result ClassLikeDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of mergeLexicalEnvironment result NodeArray<Statement> cannot be served by implementation result Statement[] &#124; NodeArray<Statement> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseExpectedToken parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseExpectedTokenJSDoc parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseModifiers result NodeArray<Modifier> &#124; undefined cannot be served by implementation result NodeArray<ModifierLike> &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseNamedImportsOrExports result NamedImports cannot be served by implementation result NamedImportsOrExports | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseOptionalToken parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseOptionalTokenJSDoc parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of removeTrailingDirectorySeparator result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replaceDecoratorsAndModifiers result T cannot be served by implementation result ClassDeclaration &#124; ClassExpression &#124; GetAccessorDeclaration &#124; MethodDeclaration &#124; ParameterDeclaration &#124; PropertyDeclaration &#124; SetAccessorDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replaceModifiers result T cannot be served by implementation result ArrowFunction &#124; ClassDeclaration &#124; ClassExpression &#124; ConstructorDeclaration &#124; ConstructorTypeNode &#124; ... 19 more ... &#124; VariableStatement | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replacePropertyName result T cannot be served by implementation result GetAccessorDeclaration &#124; MethodDeclaration &#124; MethodSignature &#124; PropertyAssignment &#124; PropertyDeclaration &#124; PropertySignature &#124; SetAccessorDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipOuterExpressions result T cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipParentheses result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of symbolToName result Identifier cannot be served by implementation result EntityName | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> &#124; ... 7 more ... &#124; Extract<...> cannot be served by implementation result BinaryExpression &#124; BindingElement &#124; ExportAssignment &#124; ParameterDeclaration &#124; PropertyAssignment &#124; PropertyDeclaration &#124; ShorthandPropertyAssignment &#124; VariableDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of writeTokenText result void cannot be served by implementation result number | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of getParseTreeNode result T &#124; undefined cannot be served by implementation result Node &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of makeSerializePropertySymbol result (p: Symbol, isStatic: boolean, baseType: Type &#124; undefined) => T &#124; T[] cannot be served by implementation result (p: Symbol, isStatic: boolean, baseType: Type &#124; undefined) => T &#124; (T &#124; AccessorDeclaration)[] &#124; AccessorDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of skipOuterExpressions result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 3 of getGlobalType result GenericType cannot be served by implementation result ObjectType &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | sort without a comparator | 1 | 1 | not in pinned table |  | False |

## tsc

measured on a checker-rejected entry-root program.

81 source files; 324 checker diagnostics; unit statuses {'attempted': 10551, 'split_checker_body': 147}.
Excluded nested functions: 110; gross body bytes 437999 (can overlap); union bytes 414509.
State-snapshot failure boundaries: 0.

| kind | combined | actual_lowering |
| --- | --- | --- |
| NotYet | 13881 | 13859 |
| Refused | 9393 | 2056 |
| panic | 27 | 25 |
| error | 2 | 2 |
| SkippedDependency | 5 | 5 |
| Boundary | 20457 | 20457 |

| owner | NotYet | Refused |
| --- | --- | --- |
| compiler | 9492 | 37 |
| stage3 adaptation | 131 | 5104 |
| UNOWNED | 4258 | 4252 |

### compiler

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| NotYet | a method call through a structural signature in a program with statics; use typeof the declaring class | 1832 | 1832 | compiler lesson |  | False |
| NotYet | reading node | 1091 | 1091 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a value and a value | 509 | 509 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a boolean | 314 | 314 | compiler lesson |  | False |
| NotYet | reading result | 228 | 228 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a number | 220 | 220 | compiler lesson |  | False |
| NotYet | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 189 | 189 | compiler lesson |  | False |
| NotYet | assigning to an Identifier | 166 | 166 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a value | 139 | 139 | compiler lesson |  | False |
| NotYet | a call through ?. (an optional call) | 125 | 125 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number and a boolean | 122 | 122 | compiler lesson |  | False |
| NotYet | for...of over an object | 115 | 115 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a number | 107 | 107 | compiler lesson |  | False |
| NotYet | reading type | 102 | 102 | compiler lesson |  | True |
| NotYet | a function returning T | 100 | 100 | compiler lesson |  | False |
| NotYet | a void call used as a value | 86 | 86 | compiler lesson |  | False |
| NotYet | an ElementAccessExpression | 83 | 83 | compiler lesson |  | False |
| NotYet | a function returning T &#124; undefined | 75 | 75 | compiler lesson |  | False |
| NotYet | a declaration directly in a case (wrap the case in a block) | 60 | 60 | compiler lesson |  | False |
| NotYet | reading performance | 57 | 57 | compiler lesson |  | True |
| NotYet | a value of type T | 50 | 50 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a number | 47 | 47 | compiler lesson |  | False |
| NotYet | a rest array of DiagnosticArguments | 46 | 46 | compiler lesson |  | False |
| NotYet | a call returning void &#124; undefined | 43 | 43 | compiler lesson |  | False |
| NotYet | an array of T | 43 | 43 | compiler lesson |  | False |
| NotYet | reading expression | 43 | 43 | compiler lesson |  | True |
| NotYet | reading updated | 43 | 43 | compiler lesson |  | True |
| NotYet | reading host | 41 | 41 | compiler lesson |  | True |
| NotYet | reading symbol | 40 | 40 | compiler lesson |  | True |
| Refused | a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 37 | 0 | compiler lesson | internal/lower/predicates.go / predicateProof.refused | False |
| NotYet | reading name | 36 | 36 | compiler lesson |  | True |
| NotYet | this outside a method | 34 | 34 | compiler lesson |  | False |
| NotYet | reading statement | 30 | 30 | compiler lesson |  | True |
| NotYet | reading container | 28 | 28 | compiler lesson |  | True |
| NotYet | reading clone | 27 | 27 | compiler lesson |  | True |
| NotYet | a function value returning union of differently held members | 26 | 26 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a number | 25 | 25 | compiler lesson |  | False |
| NotYet | reading resolved | 25 | 25 | compiler lesson |  | True |
| NotYet | reading declaration | 24 | 24 | compiler lesson |  | True |
| NotYet | a field of type string &#124; NodeArray<JSDocComment> &#124; undefined | 23 | 23 | compiler lesson |  | False |
| NotYet | reading compilerOptions | 23 | 23 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a string and a string | 22 | 22 | compiler lesson |  | False |
| NotYet | a value of type SolutionBuilderState<T> | 21 | 21 | compiler lesson |  | False |
| NotYet | new an Identifier | 21 | 21 | compiler lesson |  | False |
| NotYet | reading sourceFile | 21 | 21 | compiler lesson |  | True |
| NotYet | a generic function as a value | 20 | 20 | compiler lesson |  | False |
| NotYet | a value of type unknown | 20 | 20 | compiler lesson |  | False |
| NotYet | reading options | 20 | 20 | compiler lesson |  | True |
| NotYet | a value of type ResolvedConfigFilePath | 18 | 18 | compiler lesson |  | False |
| NotYet | a value of type T["kind"] | 18 | 18 | compiler lesson |  | False |
| NotYet | reading c | 18 | 18 | compiler lesson |  | True |
| NotYet | reading target | 18 | 18 | compiler lesson |  | True |
| NotYet | reading types | 18 | 18 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a value | 17 | 17 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a number | 17 | 17 | compiler lesson |  | False |
| NotYet | reading expr | 17 | 17 | compiler lesson |  | True |
| NotYet | replacing a represented method at runtime | 17 | 17 | compiler lesson |  | False |
| NotYet | a value of type object | 16 | 16 | compiler lesson |  | False |
| NotYet | optional chaining to .size on a value | 15 | 15 | compiler lesson |  | False |
| NotYet | reading file | 15 | 15 | compiler lesson |  | True |
| NotYet | reading modifiers | 15 | 15 | compiler lesson |  | True |
| NotYet | reading visited | 15 | 15 | compiler lesson |  | True |
| NotYet | reading diag | 14 | 14 | compiler lesson |  | True |
| NotYet | reading parameter | 14 | 14 | compiler lesson |  | True |
| NotYet | reading array | 13 | 13 | compiler lesson |  | True |
| NotYet | reading block | 13 | 13 | compiler lesson |  | True |
| NotYet | reading sig | 13 | 13 | compiler lesson |  | True |
| NotYet | a function returning undefined | 12 | 12 | compiler lesson |  | False |
| NotYet | a value of type ResolvedConfigFileName | 12 | 12 | compiler lesson |  | False |
| NotYet | a value of type T &#124; undefined | 12 | 12 | compiler lesson |  | False |
| NotYet | reading cache | 12 | 12 | compiler lesson |  | True |
| NotYet | reading child | 11 | 11 | compiler lesson |  | True |
| NotYet | reading diagnostics | 11 | 11 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a number | 10 | 10 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a boolean | 10 | 10 | compiler lesson |  | False |
| NotYet | a function returning U &#124; undefined | 10 | 10 | compiler lesson |  | False |
| NotYet | for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 10 | 10 | compiler lesson |  | False |
| NotYet | reading exception | 10 | 10 | compiler lesson |  | True |
| NotYet | reading existing | 10 | 10 | compiler lesson |  | True |
| NotYet | reading firstDecorator | 10 | 10 | compiler lesson |  | True |
| NotYet | reading flowType | 10 | 10 | compiler lesson |  | True |
| NotYet | reading index | 10 | 10 | compiler lesson |  | True |
| NotYet | reading objectFlags | 10 | 10 | compiler lesson |  | True |
| NotYet | reading reference | 10 | 10 | compiler lesson |  | True |
| NotYet | reading specifier | 10 | 10 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a string | 9 | 9 | compiler lesson |  | False |
| NotYet | a SpreadElement | 9 | 9 | compiler lesson |  | False |
| NotYet | an optional chain longer than one step | 9 | 9 | compiler lesson |  | False |
| NotYet | assigning a field of a value | 9 | 9 | compiler lesson |  | False |
| NotYet | reading iterationTypes | 9 | 9 | compiler lesson |  | True |
| NotYet | reading left | 9 | 9 | compiler lesson |  | True |
| NotYet | reading links | 9 | 9 | compiler lesson |  | True |
| NotYet | reading parent | 9 | 9 | compiler lesson |  | True |
| NotYet | reading parsed | 9 | 9 | compiler lesson |  | True |
| NotYet | reading temp | 9 | 9 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a number and a boolean &#124; undefined | 8 | 8 | compiler lesson |  | False |
| NotYet | a value of type ((node: Node) => boolean) &#124; undefined | 8 | 8 | compiler lesson |  | False |
| NotYet | a value of type K | 8 | 8 | compiler lesson |  | False |
| NotYet | a value of type readonly T[] &#124; undefined | 8 | 8 | compiler lesson |  | False |
| NotYet | new a ParenthesizedExpression | 8 | 8 | compiler lesson |  | False |
| NotYet | reading decl | 8 | 8 | compiler lesson |  | True |
| NotYet | reading directoryWatcher | 8 | 8 | compiler lesson |  | True |
| NotYet | reading related | 8 | 8 | compiler lesson |  | True |
| NotYet | reading statements | 8 | 8 | compiler lesson |  | True |
| NotYet | reading text | 8 | 8 | compiler lesson |  | True |
| NotYet | reading value | 8 | 8 | compiler lesson |  | True |
| NotYet | regex replacement other than a string | 8 | 8 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a boolean &#124; undefined | 7 | 7 | compiler lesson |  | False |
| NotYet | a function returning T[] | 7 | 7 | compiler lesson |  | False |
| NotYet | a rest parameter outside a nongeneric named function | 7 | 7 | compiler lesson |  | False |
| NotYet | a value of type BindableStaticNameExpression | 7 | 7 | compiler lesson |  | False |
| NotYet | a value of type NonNullable<T> | 7 | 7 | compiler lesson |  | False |
| NotYet | assigning to a NonNullExpression | 7 | 7 | compiler lesson |  | False |
| NotYet | reading autoGenerate | 7 | 7 | compiler lesson |  | True |
| NotYet | reading buildOrder | 7 | 7 | compiler lesson |  | True |
| NotYet | reading candidate | 7 | 7 | compiler lesson |  | True |
| NotYet | reading current | 7 | 7 | compiler lesson |  | True |
| NotYet | reading errorNode | 7 | 7 | compiler lesson |  | True |
| NotYet | reading identifier | 7 | 7 | compiler lesson |  | True |
| NotYet | reading includes | 7 | 7 | compiler lesson |  | True |
| NotYet | reading t | 7 | 7 | compiler lesson |  | True |
| NotYet | reading targetType | 7 | 7 | compiler lesson |  | True |
| NotYet | reading token | 7 | 7 | compiler lesson |  | True |
| NotYet | reading typeNode | 7 | 7 | compiler lesson |  | True |
| NotYet | reading watcher | 7 | 7 | compiler lesson |  | True |
| NotYet | ?.[] on a value | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a string | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a boolean | 6 | 6 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a string | 6 | 6 | compiler lesson |  | False |
| NotYet | a case whose type differs from the switch's | 6 | 6 | compiler lesson |  | False |
| NotYet | a value of type HasJSDoc &#124; undefined | 6 | 6 | compiler lesson |  | False |
| NotYet | a value of type TOuterState | 6 | 6 | compiler lesson |  | False |
| NotYet | an array of U | 6 | 6 | compiler lesson |  | False |
| NotYet | an array of boolean &#124; undefined | 6 | 6 | compiler lesson |  | False |
| NotYet | reading body | 6 | 6 | compiler lesson |  | True |
| NotYet | reading bundle | 6 | 6 | compiler lesson |  | True |
| NotYet | reading classType | 6 | 6 | compiler lesson |  | True |
| NotYet | reading constraint | 6 | 6 | compiler lesson |  | True |
| NotYet | reading e | 6 | 6 | compiler lesson |  | True |
| NotYet | reading elem | 6 | 6 | compiler lesson |  | True |
| NotYet | reading emitNode | 6 | 6 | compiler lesson |  | True |
| NotYet | reading flags | 6 | 6 | compiler lesson |  | True |
| NotYet | reading id | 6 | 6 | compiler lesson |  | True |
| NotYet | reading innerExpression | 6 | 6 | compiler lesson |  | True |
| NotYet | reading lexicallyScopedSymbol | 6 | 6 | compiler lesson |  | True |
| NotYet | reading location | 6 | 6 | compiler lesson |  | True |
| NotYet | reading parseNode | 6 | 6 | compiler lesson |  | True |
| NotYet | reading queue | 6 | 6 | compiler lesson |  | True |
| NotYet | reading reduced | 6 | 6 | compiler lesson |  | True |
| NotYet | reading res | 6 | 6 | compiler lesson |  | True |
| NotYet | reading scope | 6 | 6 | compiler lesson |  | True |
| NotYet | reading ts | 6 | 6 | compiler lesson |  | True |
| NotYet | reading typeParameters | 6 | 6 | compiler lesson |  | True |
| NotYet | reading valueDeclaration | 6 | 6 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a string and a value | 5 | 5 | compiler lesson |  | False |
| NotYet | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type HasJSDoc | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type readonly T[] | 5 | 5 | compiler lesson |  | False |
| NotYet | a value of type string &#124; null &#124; undefined | 5 | 5 | compiler lesson |  | False |
| NotYet | an array of ResolvedConfigFileName | 5 | 5 | compiler lesson |  | False |
| NotYet | reading args | 5 | 5 | compiler lesson |  | True |
| NotYet | reading assignedName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading buildOptions | 5 | 5 | compiler lesson |  | True |
| NotYet | reading cached | 5 | 5 | compiler lesson |  | True |
| NotYet | reading directoryExists | 5 | 5 | compiler lesson |  | True |
| NotYet | reading emittedExpression | 5 | 5 | compiler lesson |  | True |
| NotYet | reading exportSpecifiers | 5 | 5 | compiler lesson |  | True |
| NotYet | reading externalHelpersImportDeclaration | 5 | 5 | compiler lesson |  | True |
| NotYet | reading getter | 5 | 5 | compiler lesson |  | True |
| NotYet | reading indexInfo | 5 | 5 | compiler lesson |  | True |
| NotYet | reading indexType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading info | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isAmbient | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isAsync | 5 | 5 | compiler lesson |  | True |
| NotYet | reading isGenerator | 5 | 5 | compiler lesson |  | True |
| NotYet | reading list | 5 | 5 | compiler lesson |  | True |
| NotYet | reading moduleSymbol | 5 | 5 | compiler lesson |  | True |
| NotYet | reading objectType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading path | 5 | 5 | compiler lesson |  | True |
| NotYet | reading pos | 5 | 5 | compiler lesson |  | True |
| NotYet | reading propType | 5 | 5 | compiler lesson |  | True |
| NotYet | reading propertyName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading source | 5 | 5 | compiler lesson |  | True |
| NotYet | reading start | 5 | 5 | compiler lesson |  | True |
| NotYet | reading targetParam | 5 | 5 | compiler lesson |  | True |
| NotYet | reading typeName | 5 | 5 | compiler lesson |  | True |
| NotYet | reading v | 5 | 5 | compiler lesson |  | True |
| NotYet | reading yieldType | 5 | 5 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a value | 4 | 4 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a boolean | 4 | 4 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a value | 4 | 4 | compiler lesson |  | False |
| NotYet | a computed field name | 4 | 4 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 4 | 4 | compiler lesson |  | False |
| NotYet | a field of type NodeArray<ParameterDeclaration> &#124; readonly JSDocParameterTag[] | 4 | 4 | compiler lesson |  | False |
| NotYet | a for...of destructuring an object | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning CompilerOptionsValue | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning U | 4 | 4 | compiler lesson |  | False |
| NotYet | a function returning readonly T[] &#124; undefined | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type (ConstructorDeclaration & { body: Block; }) &#124; undefined | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type CompilerOptionsValue | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type ImmediatelyInvokedArrowFunction | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type T &#124; Program | 4 | 4 | compiler lesson |  | False |
| NotYet | a value of type object &#124; undefined | 4 | 4 | compiler lesson |  | False |
| NotYet | assigning an element of a value | 4 | 4 | compiler lesson |  | False |
| NotYet | assigning to an ObjectLiteralExpression | 4 | 4 | compiler lesson |  | False |
| NotYet | reading argument | 4 | 4 | compiler lesson |  | True |
| NotYet | reading asteriskToken | 4 | 4 | compiler lesson |  | True |
| NotYet | reading bindings | 4 | 4 | compiler lesson |  | True |
| NotYet | reading cacheKey | 4 | 4 | compiler lesson |  | True |
| NotYet | reading chain | 4 | 4 | compiler lesson |  | True |
| NotYet | reading clause | 4 | 4 | compiler lesson |  | True |
| NotYet | reading commonSourceDirectory | 4 | 4 | compiler lesson |  | True |
| NotYet | reading config | 4 | 4 | compiler lesson |  | True |
| NotYet | reading contextualType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading cooked | 4 | 4 | compiler lesson |  | True |
| NotYet | reading descriptorName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading element | 4 | 4 | compiler lesson |  | True |
| NotYet | reading emitResult | 4 | 4 | compiler lesson |  | True |
| NotYet | reading excludeRegex | 4 | 4 | compiler lesson |  | True |
| NotYet | reading exportStatement | 4 | 4 | compiler lesson |  | True |
| NotYet | reading exportedName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading extensionGroup | 4 | 4 | compiler lesson |  | True |
| NotYet | reading hasDefaultClause | 4 | 4 | compiler lesson |  | True |
| NotYet | reading importDeclaration | 4 | 4 | compiler lesson |  | True |
| NotYet | reading initializer | 4 | 4 | compiler lesson |  | True |
| NotYet | reading jsxFactorySymbol | 4 | 4 | compiler lesson |  | True |
| NotYet | reading key | 4 | 4 | compiler lesson |  | True |
| NotYet | reading kind | 4 | 4 | compiler lesson |  | True |
| NotYet | reading map | 4 | 4 | compiler lesson |  | True |
| NotYet | reading mapping | 4 | 4 | compiler lesson |  | True |
| NotYet | reading nameType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading nodeModulesFolderExists | 4 | 4 | compiler lesson |  | True |
| NotYet | reading normalized | 4 | 4 | compiler lesson |  | True |
| NotYet | reading openParenPosition | 4 | 4 | compiler lesson |  | True |
| NotYet | reading original | 4 | 4 | compiler lesson |  | True |
| NotYet | reading output | 4 | 4 | compiler lesson |  | True |
| NotYet | reading propertyOriginalNode | 4 | 4 | compiler lesson |  | True |
| NotYet | reading props | 4 | 4 | compiler lesson |  | True |
| NotYet | reading rawText | 4 | 4 | compiler lesson |  | True |
| NotYet | reading refPath | 4 | 4 | compiler lesson |  | True |
| NotYet | reading referencedName | 4 | 4 | compiler lesson |  | True |
| NotYet | reading restParameter | 4 | 4 | compiler lesson |  | True |
| NotYet | reading restType | 4 | 4 | compiler lesson |  | True |
| NotYet | reading seen | 4 | 4 | compiler lesson |  | True |
| NotYet | reading setter | 4 | 4 | compiler lesson |  | True |
| NotYet | reading signature | 4 | 4 | compiler lesson |  | True |
| NotYet | reading sourceIndex | 4 | 4 | compiler lesson |  | True |
| NotYet | reading typeArguments | 4 | 4 | compiler lesson |  | True |
| NotYet | reading varStatement | 4 | 4 | compiler lesson |  | True |
| NotYet | reading watchCompilerHost | 4 | 4 | compiler lesson |  | True |
| NotYet | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 3 | 3 | compiler lesson |  | False |
| NotYet | Object.entries on a shape not proven by a plain literal or its const binding | 3 | 3 | compiler lesson |  | False |
| NotYet | RegExp with a nonconstant pattern | 3 | 3 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a boolean | 3 | 3 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a union of differently held members | 3 | 3 | compiler lesson |  | False |
| NotYet | a ConditionalExpression as a statement | 3 | 3 | compiler lesson |  | False |
| NotYet | a Map of false &#124; MutableFileSystemEntries | 3 | 3 | compiler lesson |  | False |
| NotYet | a PrefixUnaryExpression on a number | 3 | 3 | compiler lesson |  | False |
| NotYet | a destructured name that isn't plain | 3 | 3 | compiler lesson |  | False |
| NotYet | a field holding union of differently held members | 3 | 3 | compiler lesson |  | False |
| NotYet | a field of type string &#124; number &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning NodeArray<T> | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning ResolvedConfigFileName | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning T[] &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a function returning object | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type BindableAccessExpression | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type BindableObjectDefinePropertyCall | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type Children &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type EndOfFileToken | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type InitializedVariableDeclaration | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type WrappedExpression<AnonymousFunctionDefinition> | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type X | 3 | 3 | compiler lesson |  | False |
| NotYet | a value of type false &#124; TypeOnlyAliasDeclaration &#124; undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | an array of NonNullable<T> | 3 | 3 | compiler lesson |  | False |
| NotYet | an array of undefined | 3 | 3 | compiler lesson |  | False |
| NotYet | lastIndexOf with these arguments | 3 | 3 | compiler lesson |  | False |
| NotYet | reading addUndefined | 3 | 3 | compiler lesson |  | True |
| NotYet | reading awaitedType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading baseTypes | 3 | 3 | compiler lesson |  | True |
| NotYet | reading buildInfo | 3 | 3 | compiler lesson |  | True |
| NotYet | reading callSignatures | 3 | 3 | compiler lesson |  | True |
| NotYet | reading callee | 3 | 3 | compiler lesson |  | True |
| NotYet | reading configFile | 3 | 3 | compiler lesson |  | True |
| NotYet | reading constructor | 3 | 3 | compiler lesson |  | True |
| NotYet | reading declaredType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading diagnosticStart | 3 | 3 | compiler lesson |  | True |
| NotYet | reading diff | 3 | 3 | compiler lesson |  | True |
| NotYet | reading dupFile | 3 | 3 | compiler lesson |  | True |
| NotYet | reading emitSuperHelpers | 3 | 3 | compiler lesson |  | True |
| NotYet | reading enclosingDeclaration | 3 | 3 | compiler lesson |  | True |
| NotYet | reading encodeURI | 3 | 3 | compiler lesson |  | True |
| NotYet | reading entityName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading evaluated | 3 | 3 | compiler lesson |  | True |
| NotYet | reading fakeScope | 3 | 3 | compiler lesson |  | True |
| NotYet | reading fileName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading filesSpecs | 3 | 3 | compiler lesson |  | True |
| NotYet | reading filtered | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstAccessorWithDecorators | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstDecl | 3 | 3 | compiler lesson |  | True |
| NotYet | reading firstDeclaration | 3 | 3 | compiler lesson |  | True |
| NotYet | reading functionName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading getCommonSourceDirectory | 3 | 3 | compiler lesson |  | True |
| NotYet | reading graphNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading hasRestParameter | 3 | 3 | compiler lesson |  | True |
| NotYet | reading hostNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading inferredProp | 3 | 3 | compiler lesson |  | True |
| NotYet | reading introducesError | 3 | 3 | compiler lesson |  | True |
| NotYet | reading isImmediatelyInvoked | 3 | 3 | compiler lesson |  | True |
| NotYet | reading isSimilarNode | 3 | 3 | compiler lesson |  | True |
| NotYet | reading jsFilePath | 3 | 3 | compiler lesson |  | True |
| NotYet | reading jsdocAliasDecl | 3 | 3 | compiler lesson |  | True |
| NotYet | reading keyPropertyName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading lastDecorator | 3 | 3 | compiler lesson |  | True |
| NotYet | reading lastError | 3 | 3 | compiler lesson |  | True |
| NotYet | reading lastParam | 3 | 3 | compiler lesson |  | True |
| NotYet | reading length | 3 | 3 | compiler lesson |  | True |
| NotYet | reading loadPackageJsonMainState | 3 | 3 | compiler lesson |  | True |
| NotYet | reading mappedType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading method | 3 | 3 | compiler lesson |  | True |
| NotYet | reading methodType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading moveModifiers | 3 | 3 | compiler lesson |  | True |
| NotYet | reading named | 3 | 3 | compiler lesson |  | True |
| NotYet | reading oldEmitKind | 3 | 3 | compiler lesson |  | True |
| NotYet | reading oldOptions | 3 | 3 | compiler lesson |  | True |
| NotYet | reading operand | 3 | 3 | compiler lesson |  | True |
| NotYet | reading outerTypeParameters | 3 | 3 | compiler lesson |  | True |
| NotYet | reading parameters | 3 | 3 | compiler lesson |  | True |
| NotYet | reading parentType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading prop | 3 | 3 | compiler lesson |  | True |
| NotYet | reading property | 3 | 3 | compiler lesson |  | True |
| NotYet | reading qualifiedName | 3 | 3 | compiler lesson |  | True |
| NotYet | reading questionToken | 3 | 3 | compiler lesson |  | True |
| NotYet | reading range | 3 | 3 | compiler lesson |  | True |
| NotYet | reading resolution | 3 | 3 | compiler lesson |  | True |
| NotYet | reading resolvedRequire | 3 | 3 | compiler lesson |  | True |
| NotYet | reading returnStatement | 3 | 3 | compiler lesson |  | True |
| NotYet | reading rhsValue | 3 | 3 | compiler lesson |  | True |
| NotYet | reading root | 3 | 3 | compiler lesson |  | True |
| NotYet | reading sourceEmitHelpers | 3 | 3 | compiler lesson |  | True |
| NotYet | reading state | 3 | 3 | compiler lesson |  | True |
| NotYet | reading suggestion | 3 | 3 | compiler lesson |  | True |
| NotYet | reading thisType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading tupleType | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeArgumentTypes | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeCopy | 3 | 3 | compiler lesson |  | True |
| NotYet | reading typeSymbol | 3 | 3 | compiler lesson |  | True |
| NotYet | a BinaryExpression with a boolean &#124; undefined and a boolean | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a number &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a string | 2 | 2 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a value and a number &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a Map of HostFileInfo | 2 | 2 | compiler lesson |  | False |
| NotYet | a Map of VisitResult<ExportAssignment &#124; LateVisibilityPaintedStatement &#124; undefined> | 2 | 2 | compiler lesson |  | False |
| NotYet | a comparator that doesn't take two elements and return a number | 2 | 2 | compiler lesson |  | False |
| NotYet | a destructured name held otherwise than its field | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type AnyBuildOrder &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type boolean &#124; (() => boolean) &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a field of type false &#124; string[] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning NonNullable<T> | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning PackageJson[K] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning SortedReadonlyArray<T> | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning V | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning WatchFactory<X, Y>[T] | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning readonly T[] | 2 | 2 | compiler lesson |  | False |
| NotYet | a function returning void &#124; "skip" | 2 | 2 | compiler lesson |  | False |
| NotYet | a number &#124; undefined argument to slice | 2 | 2 | compiler lesson |  | False |
| NotYet | a number &#124; undefined argument to substring | 2 | 2 | compiler lesson |  | False |
| NotYet | a rest array of (string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined)[] | 2 | 2 | compiler lesson |  | False |
| NotYet | a rest parameter other than an array | 2 | 2 | compiler lesson |  | False |
| NotYet | a tagged template other than the intrinsic String.raw | 2 | 2 | compiler lesson |  | False |
| NotYet | a template interpolating an object, an array, a map, a function or undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a tuple element of type string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type AnonymousFunctionDefinition | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type CompilerHost & ReadBuildProgramHost | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type K &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type LeftHandSideExpression & Identifier | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ParameterPropertyDeclaration | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type SourceFile | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type T[] | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type ThisCapturingVariableDeclaration | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type U | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type U &#124; readonly U[] &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type U &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type V &#124; undefined | 2 | 2 | compiler lesson |  | False |
| NotYet | a value of type false &#124; RegExpExecArray &#124; null | 2 | 2 | compiler lesson |  | False |
| NotYet | an array of Child | 2 | 2 | compiler lesson |  | False |
| NotYet | an array of V | 2 | 2 | compiler lesson |  | False |
| NotYet | destructuring a value | 2 | 2 | compiler lesson |  | False |
| NotYet | destructuring anything but a tuple into [names] | 2 | 2 | compiler lesson |  | False |
| NotYet | indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 | 2 | compiler lesson |  | False |
| NotYet | passing union of differently held members to a function value | 2 | 2 | compiler lesson |  | False |
| NotYet | push with other than one value | 2 | 2 | compiler lesson |  | False |
| NotYet | reading aComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading aParts | 2 | 2 | compiler lesson |  | True |
| NotYet | reading accessorDeclarations | 2 | 2 | compiler lesson |  | True |
| NotYet | reading allAccessors | 2 | 2 | compiler lesson |  | True |
| NotYet | reading allowStructuralFallback | 2 | 2 | compiler lesson |  | True |
| NotYet | reading antecedents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading arg | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assignClassAliasInStaticBlock | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assignment | 2 | 2 | compiler lesson |  | True |
| NotYet | reading assumeInitialized | 2 | 2 | compiler lesson |  | True |
| NotYet | reading baseClassType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading baseObjectType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading basePath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading cacheAssignment | 2 | 2 | compiler lesson |  | True |
| NotYet | reading cachedPackageJson | 2 | 2 | compiler lesson |  | True |
| NotYet | reading call | 2 | 2 | compiler lesson |  | True |
| NotYet | reading callbackToAdd | 2 | 2 | compiler lesson |  | True |
| NotYet | reading capturedLeft | 2 | 2 | compiler lesson |  | True |
| NotYet | reading classDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading cloned | 2 | 2 | compiler lesson |  | True |
| NotYet | reading columns | 2 | 2 | compiler lesson |  | True |
| NotYet | reading commentRange | 2 | 2 | compiler lesson |  | True |
| NotYet | reading comments | 2 | 2 | compiler lesson |  | True |
| NotYet | reading compilerHost | 2 | 2 | compiler lesson |  | True |
| NotYet | reading constructorDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading containingFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading contextualAwaitedType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading curCategory | 2 | 2 | compiler lesson |  | True |
| NotYet | reading currentNode | 2 | 2 | compiler lesson |  | True |
| NotYet | reading data | 2 | 2 | compiler lesson |  | True |
| NotYet | reading declBlocked | 2 | 2 | compiler lesson |  | True |
| NotYet | reading declarations | 2 | 2 | compiler lesson |  | True |
| NotYet | reading decorator | 2 | 2 | compiler lesson |  | True |
| NotYet | reading defaultIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading defaultType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading diagnostic | 2 | 2 | compiler lesson |  | True |
| NotYet | reading diagnosticWithLocation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading directoryPath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading elementTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading emitFlags | 2 | 2 | compiler lesson |  | True |
| NotYet | reading emittedOperand | 2 | 2 | compiler lesson |  | True |
| NotYet | reading enclosingClass | 2 | 2 | compiler lesson |  | True |
| NotYet | reading endLabel | 2 | 2 | compiler lesson |  | True |
| NotYet | reading equalsToken | 2 | 2 | compiler lesson |  | True |
| NotYet | reading errorInfo | 2 | 2 | compiler lesson |  | True |
| NotYet | reading errorRecord | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportEquals | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exportStarFunction | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exported | 2 | 2 | compiler lesson |  | True |
| NotYet | reading exports | 2 | 2 | compiler lesson |  | True |
| NotYet | reading extraInitializersName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading firstArgument | 2 | 2 | compiler lesson |  | True |
| NotYet | reading firstParameterIsThis | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fixedEndLength | 2 | 2 | compiler lesson |  | True |
| NotYet | reading flatDiagnostics | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fromCache | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fromComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading fullName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading func | 2 | 2 | compiler lesson |  | True |
| NotYet | reading generatorYieldType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading getCanonicalFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading getModifiedTime | 2 | 2 | compiler lesson |  | True |
| NotYet | reading gutterWidth | 2 | 2 | compiler lesson |  | True |
| NotYet | reading hasDefault | 2 | 2 | compiler lesson |  | True |
| NotYet | reading helper | 2 | 2 | compiler lesson |  | True |
| NotYet | reading heritageClause | 2 | 2 | compiler lesson |  | True |
| NotYet | reading i | 2 | 2 | compiler lesson |  | True |
| NotYet | reading immediate | 2 | 2 | compiler lesson |  | True |
| NotYet | reading implementsTypeNodes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading indexInfos | 2 | 2 | compiler lesson |  | True |
| NotYet | reading initialType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading innerModuleSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading instantiatedBaseType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading instantiation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading intrinsicElementsType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading invalidatedProject | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isDeferredMappedIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isFinite | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isJSDoc | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isOptionalChain | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isSetonlyAccessor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading isUsed | 2 | 2 | compiler lesson |  | True |
| NotYet | reading jsdocParameters | 2 | 2 | compiler lesson |  | True |
| NotYet | reading label | 2 | 2 | compiler lesson |  | True |
| NotYet | reading last | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lastElement | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lastParamVariadicType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading leadingNewlines | 2 | 2 | compiler lesson |  | True |
| NotYet | reading leftType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading len | 2 | 2 | compiler lesson |  | True |
| NotYet | reading line | 2 | 2 | compiler lesson |  | True |
| NotYet | reading lineNumber | 2 | 2 | compiler lesson |  | True |
| NotYet | reading linesBeforeDot | 2 | 2 | compiler lesson |  | True |
| NotYet | reading linkType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading local | 2 | 2 | compiler lesson |  | True |
| NotYet | reading mapped | 2 | 2 | compiler lesson |  | True |
| NotYet | reading match | 2 | 2 | compiler lesson |  | True |
| NotYet | reading maxErrors | 2 | 2 | compiler lesson |  | True |
| NotYet | reading maybeParameters | 2 | 2 | compiler lesson |  | True |
| NotYet | reading members | 2 | 2 | compiler lesson |  | True |
| NotYet | reading methodSignatures | 2 | 2 | compiler lesson |  | True |
| NotYet | reading min | 2 | 2 | compiler lesson |  | True |
| NotYet | reading minor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading missing | 2 | 2 | compiler lesson |  | True |
| NotYet | reading modifier | 2 | 2 | compiler lesson |  | True |
| NotYet | reading mustBeRemoved | 2 | 2 | compiler lesson |  | True |
| NotYet | reading names | 2 | 2 | compiler lesson |  | True |
| NotYet | reading namespaceDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading needsParens | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newAliasSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newParametersArray | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newParsedCommandLine | 2 | 2 | compiler lesson |  | True |
| NotYet | reading newSignature | 2 | 2 | compiler lesson |  | True |
| NotYet | reading nextKey | 2 | 2 | compiler lesson |  | True |
| NotYet | reading noTruncation | 2 | 2 | compiler lesson |  | True |
| NotYet | reading nodeId | 2 | 2 | compiler lesson |  | True |
| NotYet | reading numTypeArguments | 2 | 2 | compiler lesson |  | True |
| NotYet | reading ok | 2 | 2 | compiler lesson |  | True |
| NotYet | reading omitKeyType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading openBracePosition | 2 | 2 | compiler lesson |  | True |
| NotYet | reading originalClassDecl | 2 | 2 | compiler lesson |  | True |
| NotYet | reading other | 2 | 2 | compiler lesson |  | True |
| NotYet | reading otherAccessor | 2 | 2 | compiler lesson |  | True |
| NotYet | reading override | 2 | 2 | compiler lesson |  | True |
| NotYet | reading ownKey | 2 | 2 | compiler lesson |  | True |
| NotYet | reading paramSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parentSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parenthesizerRule | 2 | 2 | compiler lesson |  | True |
| NotYet | reading parsedCommandLine | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pathAndExtension | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pattern | 2 | 2 | compiler lesson |  | True |
| NotYet | reading patterns | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pendingDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading possiblyOutOfBoundsType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading pragma | 2 | 2 | compiler lesson |  | True |
| NotYet | reading predicate | 2 | 2 | compiler lesson |  | True |
| NotYet | reading primaryTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading programDiagnosticsInFile | 2 | 2 | compiler lesson |  | True |
| NotYet | reading propName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading questionDotToken | 2 | 2 | compiler lesson |  | True |
| NotYet | reading react | 2 | 2 | compiler lesson |  | True |
| NotYet | reading real | 2 | 2 | compiler lesson |  | True |
| NotYet | reading reducedTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading relatedInfo | 2 | 2 | compiler lesson |  | True |
| NotYet | reading relativeFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading remove | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolutions | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedFileName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedProject | 2 | 2 | compiler lesson |  | True |
| NotYet | reading resolvedTypeSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading restIdent | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnMethod | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading returnTypeNode | 2 | 2 | compiler lesson |  | True |
| NotYet | reading right | 2 | 2 | compiler lesson |  | True |
| NotYet | reading s | 2 | 2 | compiler lesson |  | True |
| NotYet | reading savedPreserveSourceNewlines | 2 | 2 | compiler lesson |  | True |
| NotYet | reading semanticDiagnostics | 2 | 2 | compiler lesson |  | True |
| NotYet | reading shorterParamType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading shortest | 2 | 2 | compiler lesson |  | True |
| NotYet | reading signatureDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFilePath | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFileWithAddedExtension | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceFiles | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceRoot | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceSymbolFile | 2 | 2 | compiler lesson |  | True |
| NotYet | reading sourceTypes | 2 | 2 | compiler lesson |  | True |
| NotYet | reading spread | 2 | 2 | compiler lesson |  | True |
| NotYet | reading str | 2 | 2 | compiler lesson |  | True |
| NotYet | reading substitute | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetIndex | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetReturnType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading targetSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading terminalWidth | 2 | 2 | compiler lesson |  | True |
| NotYet | reading testedSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading thisContainer | 2 | 2 | compiler lesson |  | True |
| NotYet | reading timerToUpdateChildWatches | 2 | 2 | compiler lesson |  | True |
| NotYet | reading toComponents | 2 | 2 | compiler lesson |  | True |
| NotYet | reading toWatch | 2 | 2 | compiler lesson |  | True |
| NotYet | reading transform | 2 | 2 | compiler lesson |  | True |
| NotYet | reading trueType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeOnlyDeclaration | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeOnlyDeclarationIsExportStar | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeParameter | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typeParams | 2 | 2 | compiler lesson |  | True |
| NotYet | reading typesVersions | 2 | 2 | compiler lesson |  | True |
| NotYet | reading valueSymbol | 2 | 2 | compiler lesson |  | True |
| NotYet | reading valueType | 2 | 2 | compiler lesson |  | True |
| NotYet | reading varDecl | 2 | 2 | compiler lesson |  | True |
| NotYet | reading variable | 2 | 2 | compiler lesson |  | True |
| NotYet | reading versionPaths | 2 | 2 | compiler lesson |  | True |
| NotYet | reading visibleDefaultBinding | 2 | 2 | compiler lesson |  | True |
| NotYet | reading visitedAccessorName | 2 | 2 | compiler lesson |  | True |
| NotYet | reading widened | 2 | 2 | compiler lesson |  | True |
| NotYet | reading yieldedType | 2 | 2 | compiler lesson |  | True |
| NotYet | spreading an array of other elements | 2 | 2 | compiler lesson |  | False |
| NotYet | storing true &#124; Node &#124; undefined in a field | 2 | 2 | compiler lesson |  | False |
| NotYet | .length on a value | 1 | 1 | compiler lesson |  | False |
| NotYet | ?. to a number, which would be number &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | JSON.stringify a union containing containers without runtime element metadata | 1 | 1 | compiler lesson |  | False |
| NotYet | Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a boolean and a number &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number and a union of differently held members | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a number &#124; undefined and a string | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a string and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a BinaryExpression with a union of differently held members and a union of differently held members | 1 | 1 | compiler lesson |  | False |
| NotYet | a ClassExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a DeleteExpression as a statement | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of CompilerOptionsValue | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of ResolvedConfigFilePath | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of T | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map of string &#124; number | 1 | 1 | compiler lesson |  | False |
| NotYet | a Map whose key and value types aren't known | 1 | 1 | compiler lesson |  | False |
| NotYet | a PostfixUnaryExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 1 | 1 | compiler lesson |  | False |
| NotYet | a YieldExpression as a statement | 1 | 1 | compiler lesson |  | False |
| NotYet | a base that isn't a declared class | 1 | 1 | compiler lesson |  | False |
| NotYet | a call returning T | 1 | 1 | compiler lesson |  | False |
| NotYet | a case that isn't a constant | 1 | 1 | compiler lesson |  | False |
| NotYet | a destructured parameter beside a parameter with a default | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "list" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type "boolean" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number> | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type 0 &#124; boolean &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type boolean &#124; (() => boolean) | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type string &#124; false | 1 | 1 | compiler lesson |  | False |
| NotYet | a field of type string &#124; false &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ((...args: A) => R) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning () => T | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (ConstructorDeclaration & { body: Block; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (arg: A) => T | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning (arg: T) => boolean | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning A | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning AnyValidImportOrReExport | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning AnyValidImportOrReExport &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning BuildInvalidedProject<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning BuildInvalidedProject<T> &#124; UpdateOutputFileStampsProject | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning CapturedThis | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassExpression &#124; ImmediatelyInvokedArrowFunction | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ClassThisAssignmentBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning Declaration & HasModifiers | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning EvaluatorResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning HasJSDoc &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ImmediatelyInvokedArrowFunction | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning InferenceContext &#124; (T & undefined) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning InvalidatedProject<T> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MemberName &#124; (Expression & (NumericLiteral &#124; StringLiteralLike)) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MissingList<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ModeAwareCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ModuleOrTypeReferenceResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning MultiMap<K, V> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning NodeArray<NonNullable<T>> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning NonRelativeNameResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning PerDirectoryResolutionCache<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning Queue<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning R | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ResolvedConfigFilePath | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning ReusableDiagnosticMessageChain | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning SolutionBuilder<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning SyntheticSuper | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; EmptyStatement &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; Identifier | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; NumericLiteral &#124; StringLiteral &#124; BooleanLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; StringLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T &#124; readonly T[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T1 & T2 | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TEntry &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TOut | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TOut &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TPrivateEntry &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TResult | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning T[][] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TransformationResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TypeMapper &#124; (T & undefined) | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning TypeOnlyAliasDeclaration &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning U[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning U[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning V &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning V[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning VisitResult<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning WatchCompilerHostOfConfigFile<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning WatchCompilerHostOfFilesAndCompilerOptions<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning object &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly Resolution[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly U[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning readonly U[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning string &#124; object | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning unknown | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning void &#124; number &#124; Symbol | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { [P in K as `${P}`]?: T[]; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { modifiers: NodeArray<Modifier> &#124; undefined; referencedName: Expression &#124; undefined; name: PropertyName; initializersName: Identifier &#124; undefined; descriptorName: Identifier &#124; undefined; thisArg: Identifier &#124; undefined; extraInitializersName?: never; } &#124; ... | 1 | 1 | compiler lesson |  | False |
| NotYet | a function returning { readonly min: number; readonly max: number; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a function value taking CreateSourceFileOptions &#124; ScriptTarget | 1 | 1 | compiler lesson |  | False |
| NotYet | a parameter that isn't a plain name | 1 | 1 | compiler lesson |  | False |
| NotYet | a rest array of T[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a spread that adds a field the source doesn't have | 1 | 1 | compiler lesson |  | False |
| NotYet | a tuple literal leaving out an element of type ModuleSpecifierEnding | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type "" &#124; ResolvedConfigFileName &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (AmbientModuleDeclaration & { name: StringLiteral; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) &#124; (... & ... 1 more ... & BinaryExpression) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (ModuleDeclaration & { name: StringLiteral; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (VariableDeclaration & { name: Identifier; }) &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (element: Node) => "quit" &#124; boolean | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (s: string) => void | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type (value: T) => void | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AccessExpression &#124; RequireOrImportCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AccessorDeclaration & { readonly name: BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type AliasDeclarationNode | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ArrowFunction &#124; BinaryExpression &#124; BindingElement &#124; Block &#124; BreakStatement &#124; CallSignatureDeclaration &#124; ... 64 more ... &#124; EndOfFileToken | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindablePropertyAssignmentExpression &#124; PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindableStaticAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CallExpression &#124; BindableStaticAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Child | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassElement &#124; ParameterPropertyDeclaration | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassNamedEvaluationHelperBlock | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type CustomTransformerFactory &#124; TransformerFactory<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type EntityNameExpression &#124; (LeftHandSideExpression & BindableStaticNameExpression) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type IncludeTypeSpaceImports | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type JSDocImportTag &#124; CanHaveModuleSpecifier | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<Path, ModeAwareCache<T>> &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Map<string, [K, V[]]> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type MapLike<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NamedDeclaration & { name: DeclarationName; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral &#124; StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NodeArray<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NonNullable<K> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type NonNullable<U> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrimitiveLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrivateEnvironment<TData, TEntry> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PrivateIdentifierInExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAccessExpression &#124; SyntheticSuper | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type PropertyDeclaration &#124; ParameterPropertyDeclaration | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ReplaceableIndexedAccessType | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type RequireOrImportCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ResolvedConfigFileName &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Set<K> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type SortedArray<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type Source | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type SourceFileOrString | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type T &#124; T[] &#124; readonly T[] &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type T1 | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TData | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TEntry | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TKind | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TKind &#124; Token<TKind> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TNode | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TransformedSuperCall | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type UnaryExpression & (BigIntLiteral &#124; NumericLiteral) | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type UnaryExpression & NumericLiteral | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type V | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclaration & { name: Identifier; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type VariableDeclarationList & { _usingBrand: void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type boolean &#124; V &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly Child[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] &#124; readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type readonly K[] | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type string &#124; null | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type string &#124; object &#124; undefined | 1 | 1 | compiler lesson |  | False |
| NotYet | a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of CanonicalKey | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of TState | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of object | 1 | 1 | compiler lesson |  | False |
| NotYet | an array of unknown | 1 | 1 | compiler lesson |  | False |
| NotYet | destructuring a string | 1 | 1 | compiler lesson |  | False |
| NotYet | for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | compiler lesson |  | False |
| NotYet | incrementing a NonNullExpression | 1 | 1 | compiler lesson |  | False |
| NotYet | iterating a value | 1 | 1 | compiler lesson |  | False |
| NotYet | new Map from something that isn't [key, value] pairs | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToMultiMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of arrayToNumericMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of createBinaryExpressionTrampoline with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of createToken with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachAncestorDirectory with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachLeadingCommentRange with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of forEachTrailingCommentRange with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of getOriginalNode with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of mutateMap with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of mutateMapSkippingNewValues with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of resolveTypeReferenceDirectiveNamesReusingOldState with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of setSerializerContextAnd with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 1 of sortAndDeduplicate with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 2 of arrayFrom with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | overload 3 of group with additional implementation type parameters | 1 | 1 | compiler lesson |  | False |
| NotYet | push on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 | 1 | compiler lesson |  | False |
| NotYet | reading absoluteSourceFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading accessibleSymbolsFromExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading accessorType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading actualFileName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading add | 1 | 1 | compiler lesson |  | True |
| NotYet | reading affectedSourceFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading afterImportPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading afterImportTagPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading aliasDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading allComponentComputedNamesSerializable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alreadyTransformed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alternateResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading alternateResultMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ambientModuleDeclare | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ancestor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ancestorFacts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading applicableByArity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading arrayType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading arrowExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading awaitToken | 1 | 1 | compiler lesson |  | True |
| NotYet | reading awaitedLeftType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading base64SourceMapText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseCandidates | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseIndexedAccess | 1 | 1 | compiler lesson |  | True |
| NotYet | reading baseType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading bindingElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading bindingList | 1 | 1 | compiler lesson |  | True |
| NotYet | reading blockedByExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading build | 1 | 1 | compiler lesson |  | True |
| NotYet | reading buildArray | 1 | 1 | compiler lesson |  | True |
| NotYet | reading buildOrderFromState | 1 | 1 | compiler lesson |  | True |
| NotYet | reading builderProgram | 1 | 1 | compiler lesson |  | True |
| NotYet | reading byteOrderMarkIndicator | 1 | 1 | compiler lesson |  | True |
| NotYet | reading callArgument | 1 | 1 | compiler lesson |  | True |
| NotYet | reading callTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canSuggestTypeof | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canUseBreakOrContinue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading candidateExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canonical | 1 | 1 | compiler lesson |  | True |
| NotYet | reading canonicalFileName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading changed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading check | 1 | 1 | compiler lesson |  | True |
| NotYet | reading checkBody | 1 | 1 | compiler lesson |  | True |
| NotYet | reading childComponents | 1 | 1 | compiler lesson |  | True |
| NotYet | reading classDeclarations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading className | 1 | 1 | compiler lesson |  | True |
| NotYet | reading classThis | 1 | 1 | compiler lesson |  | True |
| NotYet | reading closingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading collidingSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commentEnd | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commentText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading commonResolved | 1 | 1 | compiler lesson |  | True |
| NotYet | reading compareResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading comparer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading components | 1 | 1 | compiler lesson |  | True |
| NotYet | reading conditionPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading configFileText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constantValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constraintNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorFunction | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorLikeName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading constructorTypeCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containerObjectType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containingDirectoryPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containingMethod | 1 | 1 | compiler lesson |  | True |
| NotYet | reading containingSourceFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading content | 1 | 1 | compiler lesson |  | True |
| NotYet | reading contextType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentDetachedCommentInfo | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentDirectory | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentGlobalDiagnostics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentNamespace | 1 | 1 | compiler lesson |  | True |
| NotYet | reading currentWriterIndentSpacing | 1 | 1 | compiler lesson |  | True |
| NotYet | reading customTransformer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading d | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading declarationTransform | 1 | 1 | compiler lesson |  | True |
| NotYet | reading decorators | 1 | 1 | compiler lesson |  | True |
| NotYet | reading defaultDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading deprecatedTag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading diagnosticMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading diags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dir | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dirPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading directlyRelated | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminant | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminantCombinations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading discriminantType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading dist | 1 | 1 | compiler lesson |  | True |
| NotYet | reading doneType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading downleveledImport | 1 | 1 | compiler lesson |  | True |
| NotYet | reading effectiveExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading elementFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading elementType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitAsSingleStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitExplicitInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitFileKey | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitSourceMaps | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emitTrailingComma | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emittedAsTopLevel | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emittedCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading emptyIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enclosingBlockScopeContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enclosingContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading end | 1 | 1 | compiler lesson |  | True |
| NotYet | reading entry | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enumResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading enumStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading envVarStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errorMessage | 1 | 1 | compiler lesson |  | True |
| NotYet | reading errorSpan | 1 | 1 | compiler lesson |  | True |
| NotYet | reading everyClauseChecks | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exclamationToken | 1 | 1 | compiler lesson |  | True |
| NotYet | reading excludeRe | 1 | 1 | compiler lesson |  | True |
| NotYet | reading excludeSpecs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingPending | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingProp | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading existingTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exitStatus | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expando | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportClause | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportContainer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading exportedNamesStorageRef | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expressionPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading expressionResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extended | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extendedConstraint | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extendsType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extensionless | 1 | 1 | compiler lesson |  | True |
| NotYet | reading extensions | 1 | 1 | compiler lesson |  | True |
| NotYet | reading externalHelpersModuleName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading externalHelpersModuleReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading facts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading failed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading failedSignatureDeclarations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading falseSubtype | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fileSystemEntryExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fileToErrorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filesForEmit | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filesInError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading filteredTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | reading finished | 1 | 1 | compiler lesson |  | True |
| NotYet | reading first | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstAccessor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstInterfaceDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstNonzeroSegment | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstThisParameterOfUnionSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading firstVariableMatch | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fixed | 1 | 1 | compiler lesson |  | True |
| NotYet | reading following | 1 | 1 | compiler lesson |  | True |
| NotYet | reading forInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading forStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading format | 1 | 1 | compiler lesson |  | True |
| NotYet | reading fromNameType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatedName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatorFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading generatorInstantiation | 1 | 1 | compiler lesson |  | True |
| NotYet | reading genericDiag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading getAccessorType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading getFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading getSourceFileWithCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading globalCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading grandParent | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasEmptyObject | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasExistingReasonToReportErrorOn | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasJSDocFunctionType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasPrivateModifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hasSyntheticDefault | 1 | 1 | compiler lesson |  | True |
| NotYet | reading headerPadding | 1 | 1 | compiler lesson |  | True |
| NotYet | reading helpers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading hostSourceFileInfo | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iife | 1 | 1 | compiler lesson |  | True |
| NotYet | reading implDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading impliedNodeFormat | 1 | 1 | compiler lesson |  | True |
| NotYet | reading importClause | 1 | 1 | compiler lesson |  | True |
| NotYet | reading importDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading imports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inAmbientContextOrInterface | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inTupleContext | 1 | 1 | compiler lesson |  | True |
| NotYet | reading includeRe | 1 | 1 | compiler lesson |  | True |
| NotYet | reading includeSpecs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading indexedAccessType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inferences | 1 | 1 | compiler lesson |  | True |
| NotYet | reading init | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initialLocationForSecondaryLookup | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializerStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializerWithoutParens | 1 | 1 | compiler lesson |  | True |
| NotYet | reading initializersName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading inlinable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading insertIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading instantiatedTemplateType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading instantiations | 1 | 1 | compiler lesson |  | True |
| NotYet | reading internalFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading intrinsicAttribs | 1 | 1 | compiler lesson |  | True |
| NotYet | reading intrinsics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading invalidElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading invokedExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isAnonymous | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isArrowFunctionInJsx | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCallToReadHelper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCallbackTag | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isCapturedInFunction | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isClassWithConstructorReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isConfigIdentical | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isDerivedClass | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isDosStyle | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isEitherEnum | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isEmpty | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isExternalImportAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isIllegalExportDefaultInCJS | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isInExternalModule | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isLeftNaN | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isMarkdownOrJSDocLink | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isOptional | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isOverload | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isPropertyName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isReservedWord | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isRest | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isStaticMethodSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isUntypedSignatureInJSFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading isVoidPromiseError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading issuedDiagnostic | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iterator | 1 | 1 | compiler lesson |  | True |
| NotYet | reading iteratorValueStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsDoc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsDocType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsxFragPragma | 1 | 1 | compiler lesson |  | True |
| NotYet | reading jsxSpecific | 1 | 1 | compiler lesson |  | True |
| NotYet | reading keyAttr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading labeledElementDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastChild | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastJSDocParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastModifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastPart | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastSpan | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lastStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leadingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftSpread | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftThisArg | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftmost | 1 | 1 | compiler lesson |  | True |
| NotYet | reading leftmostExpressionKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading lineText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading linesAfterDot | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literal | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literalText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading literals | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localCheckDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localIndexDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading localSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mainExport | 1 | 1 | compiler lesson |  | True |
| NotYet | reading major | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mappedSource | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mapper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mappings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading markAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading matching | 1 | 1 | compiler lesson |  | True |
| NotYet | reading max | 1 | 1 | compiler lesson |  | True |
| NotYet | reading maxLength | 1 | 1 | compiler lesson |  | True |
| NotYet | reading maxNonRestParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading meaning | 1 | 1 | compiler lesson |  | True |
| NotYet | reading member | 1 | 1 | compiler lesson |  | True |
| NotYet | reading memberProps | 1 | 1 | compiler lesson |  | True |
| NotYet | reading merged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading messageChain | 1 | 1 | compiler lesson |  | True |
| NotYet | reading metadataReference | 1 | 1 | compiler lesson |  | True |
| NotYet | reading minArgumentCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading missingNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mixinCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mixinFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading mod | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleBlock | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleFileToTry | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moduleStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading moreThanOneRealChildren | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nameStr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nameText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading namedBindings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needCheckInitializer | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needSyncEval | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsModifierPreservingWrapper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading needsOutParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading negative | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newElements | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newItem | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newParam | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newParams | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | reading newValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading next | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeConstructors | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeContextFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeModulesAtTypesExists | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nodeName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading normalizedElements | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ns | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nsIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading nullishSemantics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading numParameters | 1 | 1 | compiler lesson |  | True |
| NotYet | reading numTypeParameters | 1 | 1 | compiler lesson |  | True |
| NotYet | reading objectLiterals | 1 | 1 | compiler lesson |  | True |
| NotYet | reading objectProperties | 1 | 1 | compiler lesson |  | True |
| NotYet | reading offset | 1 | 1 | compiler lesson |  | True |
| NotYet | reading oldSignature | 1 | 1 | compiler lesson |  | True |
| NotYet | reading oldSourceFiles | 1 | 1 | compiler lesson |  | True |
| NotYet | reading onlyRecordFailuresForIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading onlyRecordFailuresForPackageFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading openBracketPosition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading operandPrecedence | 1 | 1 | compiler lesson |  | True |
| NotYet | reading operatorToken | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionalDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionsOfCurCategory | 1 | 1 | compiler lesson |  | True |
| NotYet | reading optionsType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalClass | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalModuleSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading originalReadFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading otherFiles | 1 | 1 | compiler lesson |  | True |
| NotYet | reading otherOption | 1 | 1 | compiler lesson |  | True |
| NotYet | reading outputDir | 1 | 1 | compiler lesson |  | True |
| NotYet | reading overloadSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ownKeys | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ownOutputFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading ownedTags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageFileResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageJsonMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading packageRootPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading param | 1 | 1 | compiler lesson |  | True |
| NotYet | reading paramCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parameterDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parameterIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parameterNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parametersWithPropertyAssignments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parentComponents | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parentDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading parts | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pathComponents | 1 | 1 | compiler lesson |  | True |
| NotYet | reading peerDependencies | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pendingKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading pollScheduled | 1 | 1 | compiler lesson |  | True |
| NotYet | reading possibleOption | 1 | 1 | compiler lesson |  | True |
| NotYet | reading postSuper | 1 | 1 | compiler lesson |  | True |
| NotYet | reading precedingLineBreak | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prerelease | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prereleaseArray | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prevNodeIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading previousDuration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading previousGlobalDiagnostics | 1 | 1 | compiler lesson |  | True |
| NotYet | reading primaryDeclaration | 1 | 1 | compiler lesson |  | True |
| NotYet | reading program | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prologueStatementCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading promise | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propContext | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading propertyAssignmentType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading proto | 1 | 1 | compiler lesson |  | True |
| NotYet | reading prototype | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reachable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reactExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading readExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading readFileWithCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading realDeclarationPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading recursionIdentity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading recursiveInnerModule | 1 | 1 | compiler lesson |  | True |
| NotYet | reading reducedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading referencedMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading relativePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading relevantTypeParameterConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading remainingPaths | 1 | 1 | compiler lesson |  | True |
| NotYet | reading removeUndefined | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rename | 1 | 1 | compiler lesson |  | True |
| NotYet | reading requiresAddingUndefined | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolutionsChanged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedFromFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedMethodReturnType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedTypeReferenceDirective | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resolvedValueSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading restElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading restParameterSymbols | 1 | 1 | compiler lesson |  | True |
| NotYet | reading resultType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading returnOrPromisedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rightExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rootExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading rootSymbol | 1 | 1 | compiler lesson |  | True |
| NotYet | reading runtimeImportSpecifier | 1 | 1 | compiler lesson |  | True |
| NotYet | reading savedInStrictMode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading seenSymbols | 1 | 1 | compiler lesson |  | True |
| NotYet | reading segments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading semanticDiagnosticsPerFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading separatingLineTerminatorCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading serializedName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setFunc | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setProp | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setReadFileCache | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setterModifiers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading setterType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sharedLength | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldConvertCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldEmitDetachedComment | 1 | 1 | compiler lesson |  | True |
| NotYet | reading shouldEmitDotDot | 1 | 1 | compiler lesson |  | True |
| NotYet | reading signatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading signaturesWithCorrectTypeArgumentArity | 1 | 1 | compiler lesson |  | True |
| NotYet | reading singleLine | 1 | 1 | compiler lesson |  | True |
| NotYet | reading singleQuote | 1 | 1 | compiler lesson |  | True |
| NotYet | reading skipBindingPatterns | 1 | 1 | compiler lesson |  | True |
| NotYet | reading skipped | 1 | 1 | compiler lesson |  | True |
| NotYet | reading snippetElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sortedIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceFileNoExtension | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceIsJSConstructor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMap | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapFilePath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapRange | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMapUrlPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceMappings | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceSignature | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceSignatures | 1 | 1 | compiler lesson |  | True |
| NotYet | reading sourceValue | 1 | 1 | compiler lesson |  | True |
| NotYet | reading space | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spacesToEmit | 1 | 1 | compiler lesson |  | True |
| NotYet | reading specifierType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading specifiers | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spreadElement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading spreadIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading startPos | 1 | 1 | compiler lesson |  | True |
| NotYet | reading startsOnNewLine | 1 | 1 | compiler lesson |  | True |
| NotYet | reading stateVariable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading statementExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading statementOffset | 1 | 1 | compiler lesson |  | True |
| NotYet | reading staticBlocks | 1 | 1 | compiler lesson |  | True |
| NotYet | reading strName | 1 | 1 | compiler lesson |  | True |
| NotYet | reading subsequentNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading substituteConstraints | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superCallShouldBeRootLevel | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superPath | 1 | 1 | compiler lesson |  | True |
| NotYet | reading superStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading system | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tagExpression | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetCheckType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetDeclarationKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetFlags | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetHasStringIndex | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetIsJSConstructor | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetReturn | 1 | 1 | compiler lesson |  | True |
| NotYet | reading targetSymbolFile | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tempVar | 1 | 1 | compiler lesson |  | True |
| NotYet | reading templateType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading templates | 1 | 1 | compiler lesson |  | True |
| NotYet | reading testedNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisAccess | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisArgument | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisNodeOrAnySubNodesHasError | 1 | 1 | compiler lesson |  | True |
| NotYet | reading thisParameters | 1 | 1 | compiler lesson |  | True |
| NotYet | reading timerToInvalidateFailedLookupResolutions | 1 | 1 | compiler lesson |  | True |
| NotYet | reading timerToUpdateProgram | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tokenSourceMapRanges | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tokenText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tp | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tracker | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trailingComments | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trailingNewlines | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tripleSlash | 1 | 1 | compiler lesson |  | True |
| NotYet | reading trueCondition | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tryStatement | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tsPriority | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tsconfigTime | 1 | 1 | compiler lesson |  | True |
| NotYet | reading tupleTarget | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeAlias | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeArgCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeArgument | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeLiteralNode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeOfArrayLiteral | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeOfObjectLiteral | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeParameterCount | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typePredicateVariable | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeReferenceResolutionsChanged | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typeRoots | 1 | 1 | compiler lesson |  | True |
| NotYet | reading typescriptVersion | 1 | 1 | compiler lesson |  | True |
| NotYet | reading uniqueFilled | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unmatched | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwidenedType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwrappedExpr | 1 | 1 | compiler lesson |  | True |
| NotYet | reading unwrappedExprType | 1 | 1 | compiler lesson |  | True |
| NotYet | reading updatedText | 1 | 1 | compiler lesson |  | True |
| NotYet | reading usageMode | 1 | 1 | compiler lesson |  | True |
| NotYet | reading useCaseSensitiveFileNames | 1 | 1 | compiler lesson |  | True |
| NotYet | reading useStrictDirective | 1 | 1 | compiler lesson |  | True |
| NotYet | reading val | 1 | 1 | compiler lesson |  | True |
| NotYet | reading valid | 1 | 1 | compiler lesson |  | True |
| NotYet | reading validatedFilesSpecBeforeSubstitution | 1 | 1 | compiler lesson |  | True |
| NotYet | reading valueDecl | 1 | 1 | compiler lesson |  | True |
| NotYet | reading values | 1 | 1 | compiler lesson |  | True |
| NotYet | reading variableList | 1 | 1 | compiler lesson |  | True |
| NotYet | reading verbatimFromExports | 1 | 1 | compiler lesson |  | True |
| NotYet | reading version | 1 | 1 | compiler lesson |  | True |
| NotYet | reading visibilityResult | 1 | 1 | compiler lesson |  | True |
| NotYet | reading watchDirectoryKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading watchFileKind | 1 | 1 | compiler lesson |  | True |
| NotYet | reading widenedTypes | 1 | 1 | compiler lesson |  | True |
| NotYet | slice with an index that isn't a number | 1 | 1 | compiler lesson |  | False |
| NotYet | storing any in a field | 1 | 1 | compiler lesson |  | False |
| NotYet | storing false &#124; Type in a field | 1 | 1 | compiler lesson |  | False |
| NotYet | storing string &#124; number in a field | 1 | 1 | compiler lesson |  | False |

### stage3 adaptation

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| Refused | a cast the runtime can't check | 3646 | 1713 | adaptation | internal/lower/cast_proof.go / lowering.castProof | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 414 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| NotYet | a value of type any | 91 | 91 | adaptation |  | False |
| Refused | a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | 90 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where SourceFile is read | 59 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile seen as SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 35 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TupleTypeReference seen as TypeReference, which can write GenericType where TupleType is read | 25 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| NotYet | an array of any | 20 | 20 | adaptation |  | False |
| Refused | a type predicate whose return is not proven (branch is not a trusted parameter check) | 20 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type FlowNode seen as FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 19 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (liftToBlock would lose its object, and this with it) | 18 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type ResolvedType seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 16 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 15 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | yield (generators) | 15 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| NotYet | a function returning any | 14 | 14 | adaptation |  | False |
| Refused | a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as BaseType[], which can write BaseType where never is read | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 14 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as string[], which can write string where never is read | 13 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | the void operator | 13 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a generator function | 12 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionForDisallowedComma would lose its object, and this with it) | 12 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Node[] seen as unknown[], which can write unknown where Node is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Signature[], which can write Signature where never is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Symbol[], which can write Symbol where never is read | 12 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type NodeBuilderContext seen as SyntacticTypeNodeBuilderContext, which can write Required<Pick<SymbolTracker, "reportInferenceFallback">> where SymbolTrackerImpl is read | 11 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 11 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | in | 11 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Symbol &#124; undefined seen as Type &#124; undefined, which can write TypeFlags where SymbolFlags is read | 10 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Expression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as IndexInfo[], which can write IndexInfo where never is read | 9 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (realpath would lose its object, and this with it) | 8 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on type) | 8 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type never[] seen as Type[], which can write Type where never is read | 8 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a method read as a value (parenthesizeLeftSideOfAccess would lose its object, and this with it) | 7 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Declaration[] seen as Node[], which can write Node where Declaration is read | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowNode seen as FlowNode[] &#124; FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (parenthesizeOperandOfPrefixUnary would lose its object, and this with it) | 6 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (readFile would lose its object, and this with it) | 6 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on n) | 6 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type Expression seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type FlowNode &#124; undefined seen as false &#124; FlowNode &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeNode &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.TypeKeyword | 6 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| NotYet | a VoidExpression as a statement | 5 | 5 | adaptation |  | False |
| Refused | a method read as a value (trace would lose its object, and this with it) | 5 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile &#124; undefined where undefined is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowArrayMutation &#124; FlowAssignment seen as FlowNode, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 4 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol &#124; undefined seen as Declaration &#124; undefined, which can write number &#124; undefined where number is read | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ModifierFlags.None | 5 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking readonly Modifier[] &#124; undefined seen as one taking readonly ModifierLike[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getSourceFile would lose its object, and this with it) | 4 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type Declaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<GeneratedIdentifier> seen as Identifier, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type NodeArray<Statement> seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SyntheticSuper seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type never[] seen as Node[], which can write Node where never is read | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 4 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (trackSymbol would lose its object, and this with it) | 3 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on d) | 3 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type Diagnostic seen as Diagnostic, which can write SourceFile &#124; undefined where SourceFile is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GenericType &#124; ResolvedType seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Signature &#124; undefined seen as Type &#124; undefined, which can write TypeFlags where SignatureFlags is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Symbol &#124; undefined, which can write SymbolFlags where TypeFlags is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as TypeNode &#124; undefined, which can write number &#124; undefined where number is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Declaration[], which can write Declaration where never is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as TypeParameter[], which can write TypeParameter where never is read | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExpressionWithTypeArguments | 3 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking ((node: Node) => VisitResult<Node>) &#124; undefined seen as one taking ((node: Node) => VisitResult<Node &#124; undefined>) &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking JSDocTypeExpression &#124; undefined seen as one taking JSDocTypeExpression &#124; JSDocTypeLiteral &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Node seen as one taking [node: Node] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Visitor seen as one taking Visitor<TIn, Node &#124; undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeBranchOfConditionalExpression would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypesOfIntersectionType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypesOfUnionType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeNonArrayTypeOfPostfixType would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 2 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a value of type (identifierOrPrivateName: Identifier &#124; PrivateIdentifier) => string seen as (name: GeneratedIdentifier &#124; GeneratedPrivateIdentifier) => string, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type (readonly [() => IntrinsicType, __String])[] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where readonly [() => IntrinsicType, __String] is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BindingElement[] seen as unknown[], which can write unknown where BindingElement is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BlockLike seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as Expression &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassStaticBlockDeclaration &#124; PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type DiagnosticWithLocation &#124; undefined seen as Diagnostic &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type FlowCall seen as FlowNode, which can write BinaryExpression &#124; CallExpression where CallExpression is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type LeftHandSideExpression & GeneratedIdentifier seen as Expression, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type MethodDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as NodeLinks &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ParameterDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol seen as Declaration &#124; undefined, which can write number &#124; undefined where number is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type false &#124; FlowNode &#124; undefined seen as false &#124; FlowAssignment &#124; FlowLabel &#124; FlowReduceLabel &#124; FlowStart &#124; FlowSwitchClause &#124; undefined, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (Identifier &#124; StringLiteral)[], which can write Identifier &#124; StringLiteral where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as DiagnosticWithLocation[], which can write DiagnosticWithLocation where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as VariableDeclaration[], which can write VariableDeclaration where never is read | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Ternary | 2 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| NotYet | a call returning any | 1 | 1 | adaptation |  | False |
| Refused | a function taking BinaryOperatorToken seen as one taking BinaryOperatorToken &#124; BinaryOperator (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking Expression seen as one taking Expression &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking NodeArray<T> seen as one taking NodeArray<T> &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking TypeFacts.None seen as one taking number (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking TypeNode seen as one taking Node (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> &#124; undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode] &#124; ... seen as one taking ConstructorTypeNode (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: Node] seen as one taking BindingElement &#124; OmittedExpression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [node: Node] seen as one taking Declaration (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking [typeParameters: readonly TypeParameterDeclaration[] &#124; undefined, parameters: readonly ParameterDeclaration[], type: TypeNode] &#124; [modifiers: readonly Modifier[] &#124; undefined, typeParameters: ... &#124; undefined, parameters: ..., type: TypeNode] seen as one taking readonly Modifier[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking boolean seen as one taking boolean &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking readonly T[] &#124; undefined seen as one taking readonly T[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a function taking string seen as one taking string &#124; MemberName (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a method read as a value (cloneNode would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createComma would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createExpressionStatement would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createTemplateMiddle would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (createTemplateTail would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (getSymlinkCache would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (nonEscapingWrite would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeCheckTypeOfConditionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConciseBodyOfArrowFunction would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConditionOfConditionalExpression would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypeOfIntersectionType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeConstituentTypeOfUnionType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeElementTypeOfTupleType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfComputedPropertyName would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfExportDefault would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfExpressionStatement would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExpressionOfNew would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeExtendsTypeOfConditionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeLeadingTypeArgument would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfPostfixUnary would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfReadonlyTypeOperator would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeOperandOfTypeOperator would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (parenthesizeTypeOfOptionalType would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a method read as a value (throwIfCancellationRequested would lose its object, and this with it) | 1 | 0 | adaptation | internal/lower/refusals.go / lowering.refuse | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on arg) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on func) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on m) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on t) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on tagName) | 1 | 0 | adaptation | internal/lower/predicates.go / predicateProof.refused | False |
| Refused | a value of type (ClassDeclaration &#124; EnumDeclaration &#124; ExportAssignment &#124; ExportDeclaration &#124; FunctionDeclaration &#124; ... 5 more ... &#124; VariableStatement)[] seen as Statement[], which can write Statement where ClassDeclaration &#124; EnumDeclaration &#124; ExportAssignment &#124; ExportDeclaration &#124; FunctionDeclaration &#124; ... 5 more ... &#124; VariableStatement is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type (node: CommentRange) => boolean seen as (value: SynthesizedComment) => boolean, which can write number where -1 is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined, prefix?: string &#124; GeneratedNamePart &#124; undefined, suffix?: string &#124; undefined) => GeneratedIdentifier seen as { (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined): Identifier; (recordTempVariable: ((node: Identifier) => void) &#124; undefined, reservedInNestedScopes?: boolean &#124; undefined, prefix?: string &#124; ... 1 more ... &#124; undefined, suffix?: string &#124; undefined): Identifi..., whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type (resolution: ResolvedModuleWithFailedLookupLocations) => ResolvedModuleFull &#124; undefined seen as (oldResolution: ResolvedModuleWithFailedLookupLocations) => ResolutionWithResolvedFileName &#124; undefined, which can write string &#124; undefined where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType &#124; DeferredTypeReference seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AnonymousType &#124; GenericType seen as AnonymousType, which can write AnonymousType &#124; undefined where GenericType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type AssertClause seen as Mutable<AssertClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AssertEntry seen as Mutable<AssertEntry>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AutoAccessorPropertyDeclaration &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type AwaitedTypeInstantiation &#124; Type seen as Type, which can write Symbol &#124; undefined where Symbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; JsxNamespacedName &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral seen as Type, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; JsxNamespacedName &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Block seen as Mutable<Block>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Block &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type BreakStatement seen as Mutable<BreakStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CallExpression &#124; undefined seen as NodeArray<Expression> &#124; undefined, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CallSignatureDeclaration seen as Mutable<CallSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as PrimaryExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CapturedThis seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type CaseBlock seen as Mutable<CaseBlock>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Children seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ClassDeclaration &#124; ClassExpression &#124; InferTypeNode &#124; InterfaceDeclaration &#124; JSDocCallbackTag &#124; ... 4 more ... &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ConstructSignatureDeclaration seen as Mutable<ConstructSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ConstructorTypeNode seen as Mutable<ConstructorTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ContinueStatement seen as Mutable<ContinueStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Declaration[] seen as (ArrayTypeNode &#124; Declaration &#124; NodeWithTypeArguments &#124; TupleTypeNode)[], which can write ArrayTypeNode &#124; Declaration &#124; NodeWithTypeArguments &#124; TupleTypeNode where Declaration is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as Diagnostic &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type DiagnosticWithLocation[] &#124; undefined seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type EntityNameExpression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type EqualsGreaterThanToken seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ExportDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Expression &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ExpressionStatement seen as Mutable<ExpressionStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type FunctionTypeNode seen as Mutable<FunctionTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GeneratedPrivateIdentifier seen as Node &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type GetAccessorDeclaration &#124; MethodDeclaration &#124; PropertyAssignment &#124; SetAccessorDeclaration &#124; ShorthandPropertyAssignment &#124; SpreadAssignment &#124; undefined seen as Type &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of TypeFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Identifier & GeneratedIdentifier & { readonly escapedText: { __escapedIdentifier: void; } & "__this"; } seen as BindingName, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Identifier seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Identifier &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportAttribute seen as Mutable<ImportAttribute>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportAttributes seen as Mutable<ImportAttributes>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportClause seen as Mutable<ImportClause>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportDeclaration seen as Mutable<ImportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportDeclaration seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ImportTypeNode seen as Mutable<ImportTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type IndexInfo &#124; undefined seen as IndexSignatureDeclaration &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type IntrinsicType[] seen as TypeParameter[], which can write TypeParameter where IntrinsicType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IntrinsicType[] &#124; undefined seen as TypeParameter[] &#124; undefined, which can write TypeParameter where IntrinsicType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type IterationTypes[] seen as (IterationTypes &#124; undefined)[], which can write IterationTypes &#124; undefined where IterationTypes is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type JSDocAugmentsTag seen as Mutable<JSDocAugmentsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocCallbackTag seen as Mutable<JSDocCallbackTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocFunctionType seen as Mutable<JSDocFunctionType>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocImplementsTag seen as Mutable<JSDocImplementsTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocImportTag seen as Mutable<JSDocImportTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLink seen as Mutable<JSDocLink>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLinkCode seen as Mutable<JSDocLinkCode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocLinkPlain seen as Mutable<JSDocLinkPlain>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocNameReference seen as Mutable<JSDocNameReference>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocOverloadTag seen as Mutable<JSDocOverloadTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocParameterTag seen as Mutable<JSDocParameterTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocPropertyTag seen as Mutable<JSDocPropertyTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSeeTag seen as Mutable<JSDocSeeTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSignature seen as Mutable<JSDocSignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocSignature &#124; SignatureDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTemplateTag seen as Mutable<JSDocTemplateTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocText seen as Mutable<JSDocText>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeExpression seen as Mutable<JSDocTypeExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeExpression &#124; undefined seen as Signature &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SignatureFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypeLiteral seen as Mutable<JSDocTypeLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocTypedefTag seen as Mutable<JSDocTypedefTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type JSDocUnknownTag seen as Mutable<JSDocUnknownTag>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Map<string, [VariableDeclarationList, VariableDeclaration[]]> seen as Map<string, [CatchClause &#124; VariableDeclarationList, VariableDeclaration[]]>, which can write [CatchClause &#124; VariableDeclarationList, VariableDeclaration[]] where [VariableDeclarationList, VariableDeclaration[]] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type MappedTypeNode seen as Mutable<MappedTypeNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type MemberName &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration &#124; SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<NoSubstitutionTemplateLiteral> seen as Mutable<TemplateLiteralLikeNode>, which can write SyntaxKind where SyntaxKind.NoSubstitutionTemplateLiteral is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode &#124; undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode &#124; undefined; } can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamedImports seen as Mutable<NamedImports>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamespaceExport seen as Mutable<NamespaceExport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NamespaceImport seen as Mutable<NamespaceImport>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as EmitNode &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Node &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type NumberLiteralType seen as LiteralType, which can write string &#124; number &#124; PseudoBigInt where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type ParameterDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyAccessExpression &#124; SyntheticSuper seen as LeftHandSideExpression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyAccessExpression &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type PropertyName &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ResolvedType &#124; TypeReference seen as ObjectType, which can write SymbolTable &#124; undefined where SymbolTable is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type ReturnStatement seen as Mutable<ReturnStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Set<__String> seen as Set<__String &#124; undefined>, which can write __String &#124; undefined where __String is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile seen as Mutable<ModuleDeclaration &#124; SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile seen as SourceFileLike &#124; undefined, which can write readonly number[] &#124; undefined where readonly number[] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SourceFile &#124; undefined seen as NodeLinks &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceFile &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SourceMapSource seen as SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Statement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type StringLiteral seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type StringLiteralType seen as LiteralType, which can write string &#124; number &#124; PseudoBigInt where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type StringLiteralType[] seen as Type[], which can write Type where StringLiteralType is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SuperExpression seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SuperExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type SwitchStatement seen as Mutable<SwitchStatement>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type Symbol seen as Type &#124; undefined, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol seen as Type, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as GenericType &#124; undefined, which can write TypeFlags where SymbolFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as Node &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol &#124; undefined seen as SourceFile &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol[] seen as (Symbol &#124; undefined)[], which can write Symbol &#124; undefined where Symbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Symbol[] seen as Symbol[], which can write Symbol where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type SyntheticSuper seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ClassDeclaration &#124; ClassExpression &#124; GetAccessorDeclaration &#124; MethodDeclaration &#124; ParameterDeclaration &#124; PropertyDeclaration &#124; SetAccessorDeclaration can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint MethodDeclaration &#124; MethodSignature &#124; PropertyAssignment &#124; PropertyDeclaration &#124; PropertySignature &#124; AccessorDeclaration can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type T seen as T, a type parameter whose constraint Node &#124; undefined can be written, so it can write what T can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type ThisCapturingVariableDeclaration seen as VariableDeclaration, which can write EmitNode &#124; undefined where EmitNode & { autoGenerate: AutoGenerateInfo; } is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TransientSymbol[] seen as Symbol[], which can write Symbol where TransientSymbol is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type seen as TypeNode, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Expression &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as JSDocTypeExpression &#124; undefined, which can write number &#124; undefined where number is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type &#124; undefined seen as Signature &#124; undefined, which can write SignatureFlags where TypeFlags is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeOperatorNode seen as Mutable<TypeOperatorNode>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeParameterDeclaration seen as Mutable<TypeParameterDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeParameterDeclaration &#124; undefined seen as Symbol &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type TypeVariable[] seen as (Type &#124; undefined)[], which can write Type &#124; undefined where TypeVariable is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type Type[] seen as (Type &#124; undefined)[], which can write Type &#124; undefined where Type is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type VariableDeclarationList seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type YieldExpression seen as Mutable<YieldExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | a value of type [() => Type, __String][] seen as (readonly [() => Type, __String])[], which can write readonly [() => Type, __String] where [() => Type, __String] is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypedefTag)[], which can write JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypedefTag where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as (readonly (Identifier &#124; StringLiteral)[])[], which can write readonly (Identifier &#124; StringLiteral)[] where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as BindingElement[], which can write BindingElement where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as FlowNode[], which can write FlowNode where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as JSDocImportTag[], which can write JSDocImportTag where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as SymbolTable[], which can write SymbolTable where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as Symbol[] &#124; undefined, which can write Symbol where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type never[] seen as VarianceFlags[], which can write VarianceFlags where never is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type readonly TypeParameter[] &#124; undefined seen as TypeParameterDeclaration[] &#124; undefined, which can write TypeParameterDeclaration where TypeParameter is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type string[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where string is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type { introducesError: boolean; node: Identifier &#124; PropertyAccessEntityNameExpression; sym?: never; } &#124; { introducesError: boolean; node: Identifier &#124; PropertyAccessEntityNameExpression; sym: Symbol &#124; undefined; } seen as { introducesError: boolean; node: LeftHandSideExpression; }, which can write LeftHandSideExpression where Identifier &#124; PropertyAccessEntityNameExpression is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | a value of type { noInferenceFallback?: boolean &#124; undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile &#124; undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node &#124; undefined where ModuleDeclaration is read | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal / lowering.refuseElementWidening | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ElementFlags.Optional | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot NodeFlags.Const | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Block | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ConditionalType | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.DotDotDotToken | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExportAssignment | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportAttributes | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.MethodSignature &#124; SyntaxKind.MethodDeclaration &#124; SyntaxKind.Constructor &#124; SyntaxKind.GetAccessor &#124; SyntaxKind.SetAccessor &#124; SyntaxKind.CallSignature &#124; SyntaxKind.ConstructSignature &#124; ... 6 more ... &#124; SyntaxKind.JSDocFunctionType | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Parameter | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.QuestionToken | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot TempFlags | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Ternary.False | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot VarianceFlags.Contravariant | 1 | 0 | adaptation | internal/lower/invariance.go / lowering.wideningRefusal | False |

### UNOWNED

| kind | reason | count | actual_lowering | disposition | constructor | context_sensitive |
| --- | --- | --- | --- | --- | --- | --- |
| NotYet | a function inside a function (a closure) | 3238 | 3238 | not in pinned table |  | False |
| Refused | an object refinement using an open numeric enum as a literal tag | 2776 | 0 | not in pinned table |  | False |
| NotYet | a value of type __String | 173 | 173 | not in pinned table |  | False |
| Refused | var | 168 | 168 | not in pinned table |  | False |
| NotYet | a value of type Path | 118 | 118 | not in pinned table |  | False |
| Refused | &#124;&#124;= | 105 | 0 | not in pinned table |  | False |
| NotYet | an array of never | 74 | 74 | not in pinned table |  | False |
| NotYet | a BinaryExpression as a statement | 72 | 72 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (there is no body proving this parameter) | 71 | 0 | not in pinned table |  | False |
| Refused | optional property id in Node absent from structural source never, which can hide fields | 55 | 0 | not in pinned table |  | False |
| Refused | the comma operator | 47 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 24 | 10 | not in pinned table |  | False |
| NotYet | a value of type __String &#124; undefined | 23 | 23 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on kind) | 22 | 0 | not in pinned table |  | False |
| Refused | a method in object destructuring | 21 | 21 | not in pinned table |  | False |
| Refused | optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 21 | 0 | not in pinned table |  | False |
| NotYet | a call to a PropertyAccessExpression | 20 | 20 | not in pinned table |  | False |
| NotYet | a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 17 | 17 | not in pinned table |  | False |
| Refused | optional property id in ArrayBindingPattern absent from structural source never, which can hide fields | 17 | 0 | not in pinned table |  | False |
| NotYet | a function returning __String &#124; undefined | 16 | 16 | not in pinned table |  | False |
| NotYet | a function returning __String | 15 | 15 | not in pinned table |  | False |
| NotYet | a boolean &#124; undefined variable a function value captures | 14 | 14 | not in pinned table |  | False |
| Refused | a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 0 | not in pinned table |  | False |
| NotYet | a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 13 | 13 | not in pinned table |  | False |
| Refused | a definite assignment assertion ! | 12 | 0 | not in pinned table |  | False |
| Refused | a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 12 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 12 | 0 | not in pinned table |  | False |
| NotYet | a function returning Path | 11 | 11 | not in pinned table |  | False |
| Refused | an index signature | 11 | 0 | not in pinned table |  | False |
| NotYet | a field of type string &#124; DiagnosticMessageChain | 10 | 10 | not in pinned table |  | False |
| NotYet | an enum inside a function or block; declare it at module scope | 10 | 0 | not in pinned table |  | False |
| Refused | a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program &#124; undefined where Program is read | 10 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, never[]> seen as Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>, which can write string[] where never[] is read | 10 | 0 | not in pinned table |  | False |
| Refused | optional property id in BigIntLiteral absent from structural source never, which can hide fields | 10 | 0 | not in pinned table |  | False |
| NotYet | a field of type true &#124; Node &#124; undefined | 9 | 9 | not in pinned table |  | False |
| Refused | a value of type BuilderProgramState seen as ReusableBuilderProgramState, which can write Map<Path, readonly Diagnostic[] &#124; readonly ReusableDiagnostic[]> where Map<Path, readonly Diagnostic[]> is read | 9 | 0 | not in pinned table |  | False |
| NotYet | reading origin | 7 | 7 | not in pinned table |  | True |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 7 | 0 | not in pinned table |  | False |
| Refused | optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 7 | 0 | not in pinned table |  | False |
| Refused | optional property id in Identifier absent from structural source never, which can hide fields | 7 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source IntrinsicType, which can hide fields | 7 | 0 | not in pinned table |  | False |
| NotYet | reading lateSymbol | 6 | 6 | not in pinned table |  | True |
| NotYet | reading params | 6 | 6 | not in pinned table |  | True |
| Refused | a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (fileExists would lose its object, and this with it) | 6 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on token) | 6 | 0 | not in pinned table |  | False |
| Refused | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 6 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 6 | 0 | not in pinned table |  | False |
| Refused | optional property id in Expression absent from structural source never, which can hide fields | 6 | 0 | not in pinned table |  | False |
| NotYet | an overloaded function as a value | 5 | 5 | not in pinned table |  | False |
| NotYet | reading containingDirectory | 5 | 5 | not in pinned table |  | True |
| NotYet | reading context | 5 | 5 | not in pinned table |  | True |
| Refused | a method read as a value (directoryExists would lose its object, and this with it) | 5 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getCanonicalFileName would lose its object, and this with it) | 5 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerOptionsValue seen as TsConfigSourceFile &#124; CompilerOptionsValue, which can write string &#124; number where string is read | 5 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as CompilerOptionsValue, which can write string &#124; number where string is read | 5 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CompilerOptions &#124; BuilderFileEmit | 5 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source BuildOptions, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property reportsUnnecessary in Diagnostic absent from structural source DiagnosticRelatedInformation, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[] &#124; undefined; }, which can hide fields | 5 | 0 | not in pinned table |  | False |
| Refused | optional property version in PackageJson absent from structural source PackageJsonPathFields, which can hide fields | 5 | 0 | not in pinned table |  | False |
| NotYet | a namespace object used as a value; use qualified members or named module imports | 4 | 0 | not in pinned table |  | False |
| NotYet | a value of type string &#124; (void & { __escapedIdentifier: void; }) &#124; (string & { __escapedIdentifier: void; }) | 4 | 4 | not in pinned table |  | False |
| NotYet | reading cachedChain | 4 | 4 | not in pinned table |  | True |
| NotYet | reading childrenTargetType | 4 | 4 | not in pinned table |  | True |
| NotYet | reading filePath | 4 | 4 | not in pinned table |  | True |
| NotYet | reading getCurrentDirectory | 4 | 4 | not in pinned table |  | True |
| NotYet | reading oldProgram | 4 | 4 | not in pinned table |  | True |
| NotYet | reading regularType | 4 | 4 | not in pinned table |  | True |
| NotYet | reading resolvedModule | 4 | 4 | not in pinned table |  | True |
| NotYet | reading restParamSymbol | 4 | 4 | not in pinned table |  | True |
| NotYet | reading reversed | 4 | 4 | not in pinned table |  | True |
| NotYet | reading typeSet | 4 | 4 | not in pinned table |  | True |
| Refused | Object.defineProperties | 4 | 4 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on expr) | 4 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on f) | 4 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | 4 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind | 4 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 4 | 0 | not in pinned table |  | False |
| NotYet | a union of differently held members variable a function value captures | 3 | 3 | not in pinned table |  | False |
| NotYet | a value of type Identifier &#124; __String | 3 | 3 | not in pinned table |  | False |
| NotYet | reading arrayLiteral | 3 | 3 | not in pinned table |  | True |
| NotYet | reading connectors | 3 | 3 | not in pinned table |  | True |
| NotYet | reading decorationStatements | 3 | 3 | not in pinned table |  | True |
| NotYet | reading factory | 3 | 3 | not in pinned table |  | True |
| NotYet | reading inference | 3 | 3 | not in pinned table |  | True |
| NotYet | reading languageVersion | 3 | 3 | not in pinned table |  | True |
| NotYet | reading lines | 3 | 3 | not in pinned table |  | True |
| NotYet | reading metaPropertySymbol | 3 | 3 | not in pinned table |  | True |
| NotYet | reading newBaseType | 3 | 3 | not in pinned table |  | True |
| NotYet | reading newSymbol | 3 | 3 | not in pinned table |  | True |
| NotYet | reading regularNew | 3 | 3 | not in pinned table |  | True |
| NotYet | reading rootNames | 3 | 3 | not in pinned table |  | True |
| NotYet | reading sources | 3 | 3 | not in pinned table |  | True |
| NotYet | reading spreadType | 3 | 3 | not in pinned table |  | True |
| NotYet | reading targetPropertySymbol | 3 | 3 | not in pinned table |  | True |
| Refused | a filter callback that doesn't return a boolean | 3 | 3 | not in pinned table |  | False |
| Refused | a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getDirectories would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getSemanticDiagnosticsOfNextAffectedFile would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (now would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 | 0 | not in pinned table |  | False |
| Refused | a method read off its object, which loses its this when called (unbound-method) | 3 | 3 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on block) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on element) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on entry) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on info) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on member) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on tag) | 3 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on value) | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type FileReference seen as { pos: number &#124; undefined; end: number &#124; undefined; }, which can write number &#124; undefined where number is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type FutureSourceFile &#124; SourceFile seen as Pick<SourceFile, "fileName" &#124; "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier &#124; TextRange seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier &#124; undefined seen as Identifier &#124; TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, string[] &#124; never[]> seen as Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>, which can write string where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile seen as SourceFile &#124; SourceFileLike, which can write readonly number[] &#124; undefined where readonly number[] is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type TsConfigSourceFile &#124; CompilerOptionsValue seen as string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined, which can write string &#124; number where string is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type YieldExpression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as string[] &#124; never[], which can write string where never is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from StructuredType to ObjectType: optional field members has no proven compatible presence/type | 3 | 0 | not in pinned table |  | False |
| Refused | optional property id in FalseLiteral absent from structural source never, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property id in LiteralExpression & StringLiteral absent from structural source never, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property omitTrailingSemicolon in PrinterOptions absent from structural source CompilerOptions, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property rawText in TemplateLiteralLikeNode absent from structural source NumericLiteral, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property resolutionMode in FileReference absent from structural source { preserve: true; }, which can hide fields | 3 | 0 | not in pinned table |  | False |
| Refused | optional property return in ArrayIterator<Statement> absent from structural source ArrayIterator<ExpressionStatement>, which can hide fields | 3 | 0 | not in pinned table |  | False |
| NotYet | a field of type string &#124; number &#124; boolean &#124; DiagnosticMessage &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a function returning ModeAwareCacheKey | 2 | 2 | not in pinned table |  | False |
| NotYet | a function returning Path &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | 2 | 2 | not in pinned table |  | False |
| NotYet | a namespace member other than a function, type, initialized binding or nested namespace | 2 | 0 | not in pinned table |  | False |
| NotYet | a value of type Canonicalized | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFileId | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type Path &#124; undefined | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type RedirectsCacheKey | 2 | 2 | not in pinned table |  | False |
| NotYet | a value of type never | 2 | 2 | not in pinned table |  | False |
| NotYet | an array of Path | 2 | 2 | not in pinned table |  | False |
| NotYet | incrementing an Identifier | 2 | 2 | not in pinned table |  | False |
| NotYet | reading _createProgramOptions | 2 | 2 | not in pinned table |  | True |
| NotYet | reading attributesType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading baseSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading checkFlags | 2 | 2 | not in pinned table |  | True |
| NotYet | reading checkTypeDeferred | 2 | 2 | not in pinned table |  | True |
| NotYet | reading childPropName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading childrenPropName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading classLikeDeclaration | 2 | 2 | not in pinned table |  | True |
| NotYet | reading derived | 2 | 2 | not in pinned table |  | True |
| NotYet | reading emitSkipped | 2 | 2 | not in pinned table |  | True |
| NotYet | reading escapedText | 2 | 2 | not in pinned table |  | True |
| NotYet | reading fileInfos | 2 | 2 | not in pinned table |  | True |
| NotYet | reading firstType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading flattenContext | 2 | 2 | not in pinned table |  | True |
| NotYet | reading freshType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading hasArguments | 2 | 2 | not in pinned table |  | True |
| NotYet | reading hooks | 2 | 2 | not in pinned table |  | True |
| NotYet | reading indent | 2 | 2 | not in pinned table |  | True |
| NotYet | reading indexSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading isCallExpression | 2 | 2 | not in pinned table |  | True |
| NotYet | reading isZero | 2 | 2 | not in pinned table |  | True |
| NotYet | reading iteratedType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading json | 2 | 2 | not in pinned table |  | True |
| NotYet | reading jsxFragmentFactoryName | 2 | 2 | not in pinned table |  | True |
| NotYet | reading lanes | 2 | 2 | not in pinned table |  | True |
| NotYet | reading literalType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading newDirectory | 2 | 2 | not in pinned table |  | True |
| NotYet | reading newReturnType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading operandConstraint | 2 | 2 | not in pinned table |  | True |
| NotYet | reading operation | 2 | 2 | not in pinned table |  | True |
| NotYet | reading origTypeParameter | 2 | 2 | not in pinned table |  | True |
| NotYet | reading ownMap | 2 | 2 | not in pinned table |  | True |
| NotYet | reading potentiallyUnusedIdentifiers | 2 | 2 | not in pinned table |  | True |
| NotYet | reading prevStatement | 2 | 2 | not in pinned table |  | True |
| NotYet | reading previous | 2 | 2 | not in pinned table |  | True |
| NotYet | reading prototypeProperty | 2 | 2 | not in pinned table |  | True |
| NotYet | reading results | 2 | 2 | not in pinned table |  | True |
| NotYet | reading staticType | 2 | 2 | not in pinned table |  | True |
| NotYet | reading targetProp | 2 | 2 | not in pinned table |  | True |
| NotYet | reading targets | 2 | 2 | not in pinned table |  | True |
| NotYet | reading templateArguments | 2 | 2 | not in pinned table |  | True |
| NotYet | reading transformed | 2 | 2 | not in pinned table |  | True |
| NotYet | reading typeLiteralSymbol | 2 | 2 | not in pinned table |  | True |
| NotYet | reading typeVariable | 2 | 2 | not in pinned table |  | True |
| NotYet | reading useDefineForClassFields | 2 | 2 | not in pinned table |  | True |
| NotYet | reading variances | 2 | 2 | not in pinned table |  | True |
| NotYet | replacing a namespace export; keep exported functions fixed and mutate private state through them | 2 | 0 | not in pinned table |  | False |
| Refused | Object.create | 2 | 2 | not in pinned table |  | False |
| Refused | Object.defineProperty | 2 | 2 | not in pinned table |  | False |
| Refused | a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions &#124; ScriptTarget, onError?: ((message: string) => void) &#124; undefined, shouldCreateNewSourceFile?: boolean &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (clearTimeout would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitNodeWithNotification would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getBuildInfo would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getCurrentDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (hasGlobalName would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (isEmitNotificationEnabled would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setTimeout would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (substituteNode would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (toKey would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (watchDirectory would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (watchFile would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (writeFile would lose its object, and this with it) | 2 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on declaration) | 2 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on predicate) | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ArrayLiteralExpression &#124; AssignmentExpression<EqualsToken> &#124; BindingElement &#124; ElementAccessExpression &#124; ... 8 more ... &#124; VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Block seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type BreakStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] &#124; undefined where string[] is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ContinueStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticWithLocation[] &#124; undefined seen as DiagnosticRelatedInformation[] &#124; undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type FutureSourceFile &#124; SourceFile seen as Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | not in pinned table |  | False |
| Refused | a value of type Identifier[] &#124; undefined seen as ModuleExportName[] &#124; undefined, which can write ModuleExportName where Identifier is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | not in pinned table |  | False |
| Refused | a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleSpecifierResolutionHost & ModuleResolutionHost seen as ModuleResolutionHost, which can write boolean &#124; (() => boolean) &#124; undefined where (() => boolean) & (boolean &#124; (() => boolean) &#124; undefined) is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Node seen as Node &#124; TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Node &#124; TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type NumericLiteral seen as Mutable<LiteralExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type ReturnStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile seen as EmitNode &#124; undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | not in pinned table |  | False |
| Refused | a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableStatement seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] &#124; undefined, which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile &#124; undefined where SourceFile is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] seen as unknown[], which can write unknown where string is read | 2 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"FlowFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof FlowFlags is read | 2 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 | 2 | not in pinned table |  | False |
| Refused | an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.FunctionExpression | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NoSubstitutionTemplateLiteral &#124; SyntaxKind.TemplateHead &#124; SyntaxKind.TemplateMiddle &#124; SyntaxKind.TemplateTail | 2 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.WithKeyword &#124; SyntaxKind.AssertKeyword | 2 | 0 | not in pinned table |  | False |
| Refused | delete | 2 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source {}, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source InterfaceType, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property equalsToken in ShorthandPropertyAssignment absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property id in NamedExports absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in GetAccessorDeclaration absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property packageJsonScope in Pick<SourceFile, "fileName" &#124; "impliedNodeFormat" &#124; "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" &#124; "impliedNodeFormat">, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property templateFlags in NoSubstitutionTemplateLiteral absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| Refused | optional property textSourceNode in StringLiteral absent from structural source never, which can hide fields | 2 | 0 | not in pinned table |  | False |
| NotYet | .hasTrailingComma on a value | 1 | 1 | not in pinned table |  | False |
| NotYet | Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | not in pinned table |  | False |
| NotYet | Object.keys on a shape not proven by a plain literal or its const binding | 1 | 1 | not in pinned table |  | False |
| NotYet | a Map of RedirectsCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a call to an Identifier | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type "circularity" &#124; boolean | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type EmitSignature &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; Type &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; VersionPaths | 1 | 1 | not in pinned table |  | False |
| NotYet | a field of type false &#124; VersionPaths &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning CacheWithRedirects<K, V> | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning IncrementalBuildInfoFileId | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning IncrementalBuildInfoFileIdListId | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning RedirectsCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a function returning never | 1 | 1 | not in pinned table |  | False |
| NotYet | a mutable namespace export; use a module or export functions around private state | 1 | 0 | not in pinned table |  | False |
| NotYet | a namespace binding without a plain initialized name | 1 | 0 | not in pinned table |  | False |
| NotYet | a non-accessor method in an accessor literal | 1 | 1 | not in pinned table |  | False |
| NotYet | a reopened namespace or namespace merged with a runtime value; put the declarations in one namespace or use a module | 1 | 0 | not in pinned table |  | False |
| NotYet | a value of type A | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type Identifier &#124; PrivateIdentifier &#124; __String | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFileIdListId | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type IncrementalBuildInfoFilePendingEmit | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type ModeAwareCacheKey | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type PathPathComponents | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type RedirectsCacheKey &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type T &#124; null &#124; undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type __String & string | 1 | 1 | not in pinned table |  | False |
| NotYet | a value of type undefined | 1 | 1 | not in pinned table |  | False |
| NotYet | apply without a dense argument literal (length, presence and argument representations must be proven) | 1 | 1 | not in pinned table |  | False |
| NotYet | reading accessExpression | 1 | 1 | not in pinned table |  | True |
| NotYet | reading accessModifier | 1 | 1 | not in pinned table |  | True |
| NotYet | reading activeLabel | 1 | 1 | not in pinned table |  | True |
| NotYet | reading addUndefinedForParameter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading allSetOptions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading allowedEndings | 1 | 1 | not in pinned table |  | True |
| NotYet | reading annotationSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading b1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseConstructorType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseTypeNode | 1 | 1 | not in pinned table |  | True |
| NotYet | reading baseTypeNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading bases | 1 | 1 | not in pinned table |  | True |
| NotYet | reading best | 1 | 1 | not in pinned table |  | True |
| NotYet | reading binaryExpression | 1 | 1 | not in pinned table |  | True |
| NotYet | reading cachedDiagnostics | 1 | 1 | not in pinned table |  | True |
| NotYet | reading candidateDirectories | 1 | 1 | not in pinned table |  | True |
| NotYet | reading chain1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading checkTuples | 1 | 1 | not in pinned table |  | True |
| NotYet | reading childFieldType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading childrenNameType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading classInstanceType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading commonJSPropertyAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading compilerOptionsProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading computedPropertyName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading conditional | 1 | 1 | not in pinned table |  | True |
| NotYet | reading connector | 1 | 1 | not in pinned table |  | True |
| NotYet | reading containers | 1 | 1 | not in pinned table |  | True |
| NotYet | reading containingClassDecl | 1 | 1 | not in pinned table |  | True |
| NotYet | reading contextSpecifier | 1 | 1 | not in pinned table |  | True |
| NotYet | reading converters | 1 | 1 | not in pinned table |  | True |
| NotYet | reading createNodeArray | 1 | 1 | not in pinned table |  | True |
| NotYet | reading createProgramOptionsHost | 1 | 1 | not in pinned table |  | True |
| NotYet | reading currentOptions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading defaultExportSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading directory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading directoryStart | 1 | 1 | not in pinned table |  | True |
| NotYet | reading disposeScope | 1 | 1 | not in pinned table |  | True |
| NotYet | reading elementAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading emitSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading enter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading errorBindingElement | 1 | 1 | not in pinned table |  | True |
| NotYet | reading excludedProperties | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exit | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exportEqualsSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading exportStars | 1 | 1 | not in pinned table |  | True |
| NotYet | reading expressions | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ext | 1 | 1 | not in pinned table |  | True |
| NotYet | reading fakeSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading finalizeBoundary | 1 | 1 | not in pinned table |  | True |
| NotYet | reading firstRelevantLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading forcedLookupLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading freshTypeParameter | 1 | 1 | not in pinned table |  | True |
| NotYet | reading functionLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading globalTypingsCacheLocation | 1 | 1 | not in pinned table |  | True |
| NotYet | reading grid | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasInstanceProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasSignatures | 1 | 1 | not in pinned table |  | True |
| NotYet | reading hasTransformableStatics | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ids | 1 | 1 | not in pinned table |  | True |
| NotYet | reading ifStatement | 1 | 1 | not in pinned table |  | True |
| NotYet | reading immediateContainer | 1 | 1 | not in pinned table |  | True |
| NotYet | reading includeFileRegexes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading indexedType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading instantiated | 1 | 1 | not in pinned table |  | True |
| NotYet | reading instantiationExpressionType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading intrinsicType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isExportEquals | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isKnownProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading isLengthPushOrUnshift | 1 | 1 | not in pinned table |  | True |
| NotYet | reading jsxChildrenPropertyName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading jsxFactoryNamespace | 1 | 1 | not in pinned table |  | True |
| NotYet | reading labeledElementDeclarations | 1 | 1 | not in pinned table |  | True |
| NotYet | reading lastLeft | 1 | 1 | not in pinned table |  | True |
| NotYet | reading limitedConstraint | 1 | 1 | not in pinned table |  | True |
| NotYet | reading literalValue | 1 | 1 | not in pinned table |  | True |
| NotYet | reading localJsxNamespace | 1 | 1 | not in pinned table |  | True |
| NotYet | reading mapDirectory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading missingPaths | 1 | 1 | not in pinned table |  | True |
| NotYet | reading newConstraint | 1 | 1 | not in pinned table |  | True |
| NotYet | reading noInferSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading numNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading o1 | 1 | 1 | not in pinned table |  | True |
| NotYet | reading oldTime | 1 | 1 | not in pinned table |  | True |
| NotYet | reading onProgramCreateComplete | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parameterSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parentDirectory | 1 | 1 | not in pinned table |  | True |
| NotYet | reading parenthesizerRules | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prevSignature | 1 | 1 | not in pinned table |  | True |
| NotYet | reading printList | 1 | 1 | not in pinned table |  | True |
| NotYet | reading projectReferences | 1 | 1 | not in pinned table |  | True |
| NotYet | reading propertyAccess | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypePropertyType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypeSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading prototypeType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading quick | 1 | 1 | not in pinned table |  | True |
| NotYet | reading r | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rawName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rawSources | 1 | 1 | not in pinned table |  | True |
| NotYet | reading referencedFileName | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolutionDiagnostic | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolvedModuleNames | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resolver | 1 | 1 | not in pinned table |  | True |
| NotYet | reading resultFromDts | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rootPathComponents | 1 | 1 | not in pinned table |  | True |
| NotYet | reading rootResult | 1 | 1 | not in pinned table |  | True |
| NotYet | reading scanner | 1 | 1 | not in pinned table |  | True |
| NotYet | reading secondType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldResolveFactoryReference | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformInitializers | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformInitializersUsingSet | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldTransformThisInStaticInitializers | 1 | 1 | not in pinned table |  | True |
| NotYet | reading shouldWriteNativeEvents | 1 | 1 | not in pinned table |  | True |
| NotYet | reading signatureNextType | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceDiscriminantTypes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceFileAbsolutePaths | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceOrigin | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sourceProp | 1 | 1 | not in pinned table |  | True |
| NotYet | reading specialPropertyAssignmentKind | 1 | 1 | not in pinned table |  | True |
| NotYet | reading stat | 1 | 1 | not in pinned table |  | True |
| NotYet | reading suffix | 1 | 1 | not in pinned table |  | True |
| NotYet | reading sym | 1 | 1 | not in pinned table |  | True |
| NotYet | reading symbolExport | 1 | 1 | not in pinned table |  | True |
| NotYet | reading symbols | 1 | 1 | not in pinned table |  | True |
| NotYet | reading syntacticBuilderResolver | 1 | 1 | not in pinned table |  | True |
| NotYet | reading syntheticArgsSymbol | 1 | 1 | not in pinned table |  | True |
| NotYet | reading targetOrigin | 1 | 1 | not in pinned table |  | True |
| NotYet | reading targetProperty | 1 | 1 | not in pinned table |  | True |
| NotYet | reading tempSources | 1 | 1 | not in pinned table |  | True |
| NotYet | reading texts | 1 | 1 | not in pinned table |  | True |
| NotYet | reading thenFunction | 1 | 1 | not in pinned table |  | True |
| NotYet | reading thisParam | 1 | 1 | not in pinned table |  | True |
| NotYet | reading throwDiagnostic | 1 | 1 | not in pinned table |  | True |
| NotYet | reading tok | 1 | 1 | not in pinned table |  | True |
| NotYet | reading trampoline | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeKey | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeKind | 1 | 1 | not in pinned table |  | True |
| NotYet | reading typeNodes | 1 | 1 | not in pinned table |  | True |
| NotYet | reading undefinedStrippedTarget | 1 | 1 | not in pinned table |  | True |
| NotYet | reading valueParam | 1 | 1 | not in pinned table |  | True |
| NotYet | reading variableDeclarator | 1 | 1 | not in pinned table |  | True |
| NotYet | storing Path &#124; undefined in a field | 1 | 1 | not in pinned table |  | False |
| NotYet | typed array element type Uint16Array | 1 | 1 | not in pinned table |  | False |
| NotYet | var inside a namespace; use initialized private let or const | 1 | 0 | not in pinned table |  | False |
| Refused | &&= | 1 | 0 | not in pinned table |  | False |
| Refused | JSON.parse: its result's type can't be proven from the text | 1 | 1 | not in pinned table |  | False |
| Refused | Object.entries | 1 | 1 | not in pinned table |  | False |
| Refused | Object.setPrototypeOf | 1 | 1 | not in pinned table |  | False |
| Refused | a constructor object escaping before static fields are initialized | 1 | 1 | not in pinned table |  | False |
| Refused | a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking NodeArray<TypeNode> &#124; undefined seen as one taking readonly TypeNode[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean &#124; undefined, options?: WatchOptions &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number &#124; undefined, options?: WatchOptions &#124; undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) &#124; undefined, sourceFiles?: readonly SourceFile[] &#124; undefined, data?: WriteFileCallbackData &#124; undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] &#124; undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (add would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (base64decode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (base64encode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (clearScreen would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (compare would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLink would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (deleteFile would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emit would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitBuildInfo would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (emitNextAffectedFile would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (fill would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getAllDependencies would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getDeclarationDiagnostics would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getMemoryUsage would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getModifiedTime would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (getSemanticDiagnostics would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (hasChangedEmitSignature would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (log would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (onDiscoveredSymlink would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (releaseProgram would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (remove would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (repeat would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (reportTruncationError would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setBlocking would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (setModifiedTime would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (toString would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 | 0 | not in pinned table |  | False |
| Refused | a spread after the first field | 1 | 1 | not in pinned table |  | False |
| Refused | a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> &#124; undefined, which can write string &#124; undefined where string is read | 1 | 1 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (asserts cond needs a boolean parameter) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (false return can still contain LateBoundDeclaration) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on array) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on buildOrder) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on diagnostic) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on file) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on hostSourceFile) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on location) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on option) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on options) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on p) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on program) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on sourceFile) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on symbol) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on x) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (the body's true narrowing does not match null &#124; undefined) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (the predicate parameter is assigned) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to MappedPosition, not SourceMappedPosition) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to Mapping, not SourceMapping) | 1 | 0 | not in pinned table |  | False |
| Refused | a type predicate whose return is not proven (true return narrows to ReusableBuilderProgramState, not BuilderProgramStateWithDefinedProgram) | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ((level: LogLevel) => boolean) &#124; typeof log &#124; (() => AssertionLevel) &#124; ((level: AssertionLevel) => void) &#124; ((level: AssertionLevel) => boolean) &#124; ... 45 more ... &#124; ... seen as AnyFunction, which can write void where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ((node: PrivateIdentifierPropertyDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression) &#124; undefined seen as ((node: PropertyDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type () => { diagnosticMessage: DiagnosticMessage; errorNode: ExportAssignment; } seen as GetSymbolAccessibilityDiagnostic, which can write Node where ExportAssignment is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (AbstractKeyword &#124; AccessorKeyword &#124; AsyncKeyword &#124; ConstKeyword &#124; DeclareKeyword &#124; Decorator &#124; ... 7 more ... &#124; StaticKeyword)[] &#124; undefined seen as Node[] &#124; undefined, which can write Node where AbstractKeyword &#124; AccessorKeyword &#124; AsyncKeyword &#124; ConstKeyword &#124; DeclareKeyword &#124; Decorator &#124; ... 7 more ... &#124; StaticKeyword is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> &#124; undefined, which can write string &#124; undefined where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierGetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: GetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierMethodDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: MethodDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (node: PrivateIdentifierSetAccessorDeclaration, modifiers: ModifiersArray &#124; undefined) => ObjectLiteralExpression seen as ((node: SetAccessorDeclaration & { readonly name: PrivateIdentifier; }, modifiers: ModifiersArray &#124; undefined) => Expression) &#124; undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than PrivateIdentifier, which a write of PrivateIdentifier would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName &#124; undefined; } &#124; undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic &#124; undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type (value: TIn) => value is TOut seen as AnyFunction, which can write void where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; GeneratedIdentifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BigIntLiteral &#124; ComputedPropertyName &#124; GeneratedIdentifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BinaryExpression seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type BindingName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ClassDeclaration &#124; FunctionDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback &#124; undefined where WriteFileCallback is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain &#124; { messageText: string; category: DiagnosticCategory; code: number; repopulateInfo?: () => RepopulateDiagnosticChainInfo; canonicalHead?: CanonicalDiagnostic; next: ... &#124; undefined; } seen as ReusableDiagnosticMessageChain, which can write ReusableDiagnosticMessageChain[] &#124; undefined where DiagnosticMessageChain[] &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain where DiagnosticMessageChain is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type DiagnosticMessageChain[] seen as ReusableDiagnosticMessageChain[], which can write ReusableDiagnosticMessageChain[] &#124; undefined where DiagnosticMessageChain[] &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type EnumDeclaration &#124; ModuleDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ExportAssignment seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Expression &#124; GeneratedIdentifier seen as Expression &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Expression &#124; GeneratedIdentifier seen as Expression, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type FlowArrayMutation &#124; FlowAssignment &#124; FlowCall &#124; FlowCondition &#124; FlowLabel &#124; FlowReduceLabel &#124; FlowStart &#124; FlowUnreachable seen as FlowNode, which can write BindingElement &#124; Expression &#124; VariableDeclaration where BinaryExpression &#124; CallExpression is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier seen as Expression &#124; GeneratedIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier seen as GeneratedIdentifier &#124; Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier seen as GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; GeneratedPrivateIdentifier &#124; Identifier &#124; PrivateIdentifier seen as Identifier &#124; PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as GeneratedIdentifier &#124; Identifier &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as Identifier &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as string &#124; BindingName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier seen as string &#124; GeneratedIdentifier &#124; Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GeneratedIdentifier &#124; Identifier &#124; undefined seen as string &#124; ModuleExportName &#124; undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type GetAccessorDeclaration &#124; SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type Identifier[] seen as ModuleExportName[] &#124; undefined, which can write ModuleExportName where Identifier is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ImportEqualsDeclaration seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type InitializedVariableDeclaration seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, Diagnostic[]> seen as Map<Path, readonly Diagnostic[]> &#124; undefined, which can write readonly Diagnostic[] where Diagnostic[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, DirectoryWatchesOfFailedLookup> seen as Map<string, DirectoryWatchesOfFailedLookup>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<Path, string[]> seen as InvokeMap, which can write true &#124; string[] where string[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, Set<string> &#124; undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, CachedResolvedModuleWithFailedLookupLocations> seen as Map<string, ResolutionWithFailedLookupLocations> &#124; Set<ResolutionWithFailedLookupLocations> &#124; undefined, which can write ResolutionWithFailedLookupLocations where CachedResolvedModuleWithFailedLookupLocations is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ImportsNotUsedAsValues is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, JsxEmit> seen as Map<string, string &#124; number>, which can write string &#124; number where JsxEmit is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, Map<string, string[]> &#124; Map<string, never[]> &#124; Map<string, string[] &#124; never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ModuleDetectionKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ModuleDetectionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, ModuleResolutionKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where ModuleResolutionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, NewLineKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where NewLineKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, PollingWatchKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where PollingWatchKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, WatchDirectoryKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where WatchDirectoryKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, WatchFileKind> seen as "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map<string, string &#124; number>, which can write string &#124; number where WatchFileKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Map<string, string> seen as Map<string, string &#124; number>, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type MethodDeclaration &#124; PropertyAssignment &#124; AccessorDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleExportName seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleName seen as SourceMapRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type Node seen as Node &#124; SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Node &#124; SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration &#124; PropertyDeclaration &#124; PropertySignature &#124; SignatureDeclaration seen as Mutable<ParameterDeclaration &#124; PropertyDeclaration &#124; PropertySignature &#124; SignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type ParameterDeclaration &#124; VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean &#124; (() => boolean) &#124; undefined where boolean is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Path[] seen as string[], which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type PropertyAssignment seen as Mutable<PropertyAssignment &#124; ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Readonly<BuilderState> &#124; undefined seen as BuilderState &#124; undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ReferencedFile & { kind: FileIncludeKind.LibReferenceDirective; } seen as ReferencedFile, which can write ReferencedFileKind where FileIncludeKind.LibReferenceDirective is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ReportFileInError[] seen as (ReportFileInError &#124; undefined)[], which can write ReportFileInError &#124; undefined where ReportFileInError is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ResolvedModuleFull &#124; undefined seen as { path: string; originalPath: string &#124; true; extension: string; packageId: PackageId &#124; undefined; resolvedUsingTsExtension: boolean &#124; undefined; } &#124; undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string &#124; undefined, which a write of string &#124; true would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } &#124; undefined; } &#124; undefined, whose readonly field value becomes writable: a readonly field may hold something narrower than Resolved &#124; undefined, which a write of { resolved: Resolved; isExternalLibraryImport: true; } &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Set<Path> &#124; undefined seen as Set<string> &#124; undefined, which can write string where Path is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment &#124; ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type SourceFile &#124; undefined seen as FileReasonToChainCache &#124; undefined, which can write DiagnosticMessageChain[] &#124; undefined where RedirectInfo &#124; undefined is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement[] seen as Node[], which can write Node where Statement is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type Statement[] &#124; undefined seen as CaseClause[] &#124; undefined, which can write CaseClause where Statement is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type StringLiteral seen as Mutable<StringLiteral>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type T[] seen as (T &#124; undefined)[], which can write T &#124; undefined where T is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type ThrowStatement seen as TextRange &#124; undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration &#124; DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableDeclaration[] seen as VariableStatement[], which can write VariableStatement where VariableDeclaration is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls &#124; undefined)[], which can write WatchedFileWithUnchangedPolls &#124; undefined where WatchedFileWithUnchangedPolls is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as (JSDoc &#124; JSDocTag)[], which can write JSDoc &#124; JSDocTag where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as (SourceMapRange &#124; undefined)[], which can write SourceMapRange &#124; undefined where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Expression[], which can write Expression where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Identifier[], which can write Identifier where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as IntrinsicType[], which can write IntrinsicType where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as JsxAttributes[], which can write JsxAttributes where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ParameterDeclaration[], which can write ParameterDeclaration where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as Path[], which can write Path where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as PotentiallyUnusedIdentifier[], which can write PotentiallyUnusedIdentifier where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as PropertyAssignment[], which can write PropertyAssignment where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as ResolvedProjectReference[], which can write ResolvedProjectReference where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as SourceMappedPosition[], which can write SourceMappedPosition where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] seen as TransformerFactory<Bundle &#124; SourceFile>[], which can write TransformerFactory<Bundle &#124; SourceFile> where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[] &#124; SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type never[][] seen as string[][], which can write string where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type number[] seen as string &#124; (string &#124; number)[] &#124; undefined, which can write string &#124; number where number is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type readonly string[] &#124; undefined seen as RegExp[] &#124; undefined, which can write RegExp where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; GeneratedIdentifier &#124; Identifier seen as string &#124; ModuleExportName, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode &#124; undefined would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined seen as string &#124; number &#124; boolean &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> &#124; TsConfigSourceFile &#124; null &#124; undefined, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string &#124; number &#124; boolean &#124; string[] &#124; PluginImport[] &#124; ProjectReference[] &#124; (string &#124; number)[] &#124; MapLike<string[]> seen as CompilerOptionsValue, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type string[] seen as string &#124; (string &#124; number)[] &#124; undefined, which can write string &#124; number where string is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"CheckMode", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof CheckMode is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"EmitFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof EmitFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"GeneratedIdentifierFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof GeneratedIdentifierFlags is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ModifierFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ModifierFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeCheckFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof NodeCheckFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"NodeFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof NodeFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ObjectFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ObjectFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"RelationComparisonResult", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof RelationComparisonResult is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"ScriptKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof ScriptKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureCheckMode", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SignatureCheckMode is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SignatureFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SignatureFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SnippetKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SnippetKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SymbolFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SymbolFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"SyntaxKind", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof SyntaxKind is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TransformFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TransformFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFacts", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TypeFacts is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type typeof import("/workspace/stage3-scoreboard-tmp/main-adapted/src/compiler/_namespaces/ts") seen as Record<"TypeFlags", Record<string, string &#124; number>>, which can write Record<string, string &#124; number> where typeof TypeFlags is read | 1 | 1 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; BuilderFileEmit.DtsErrors &#124; ... 6 more ... &#124; BuilderFileEmit.All; } seen as { affectedFile: Program &#124; SourceFile &#124; undefined; emitKind: BuilderFileEmit; }, which can write Program &#124; SourceFile &#124; undefined where SourceFile is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind &#124; undefined where ModuleResolutionKind is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } &#124; { arguments: { name: string; }; range: CommentRange; } &#124; { arguments: { factory: string; }; range: CommentRange; } &#124; ... 5 more ... &#124; ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } &#124; { arguments: { name: string; }; range: CommentRange; } &#124; { arguments: { factory: string; }; range: CommentRange; } &#124; ... 5 more ... &#124; ...)[] &#124; ... 8 more ... &#124; ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] &#124; undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] &#124; undefined where never[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { ending: ModuleSpecifierEnding; value: string; }[] seen as { ending: ModuleSpecifierEnding &#124; undefined; value: string; }[], which can write ModuleSpecifierEnding &#124; undefined where ModuleSpecifierEnding is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] &#124; undefined; affectingLocations: string[] &#124; undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string &#124; number; minor: number; patch: number; prerelease: string &#124; readonly string[]; build: string &#124; readonly string[]; }, which can write string &#124; number where number is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { name: string &#124; undefined; path: string; }[] seen as AmdDependency[], which can write AmdDependency where { name: string &#124; undefined; path: string; } is read | 1 | 0 | not in pinned table |  | False |
| Refused | a value of type { referencedName: StringLiteral; name: PropertyName; } &#124; { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression &#124; undefined; name: PropertyName &#124; undefined; }, which can write Expression &#124; undefined where StringLiteral is read | 1 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 | 0 | not in pinned table |  | False |
| Refused | an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from "" to T: the source is not assignable to the target | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from DiagnosticRelatedInformation to Diagnostic: optional field reportsUnnecessary has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from InstantiableType &#124; UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from NodeArray<BindingElement> &#124; NodeArray<Expression> &#124; NodeArray<ArrayBindingElement> to NodeArray<Node>: optional field concat.element.propertyName has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from NodeArray<ModifierLike> & readonly Decorator[] to NodeArray<Decorator>: optional field concat.element.parent.name has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven relation from StructuredType to AnonymousType: optional field target has no proven compatible presence/type | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 1 | 1 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuildStep.Done | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; BuilderFileEmit.DtsErrors &#124; BuilderFileEmit.DtsEmit &#124; ... 5 more ... &#124; BuilderFileEmit.All | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot BuilderFileEmit.None | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CharacterCodes.plus | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot CharacterCodes.slash | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot Connection | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot EmitOnly.Dts | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot ModuleKind.ESNext | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot NodeFlags.NestedNamespace | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot PunctuationOrKeywordSyntaxKind | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ClassStaticBlockDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ExtendsKeyword &#124; SyntaxKind.ImplementsKeyword | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ImportDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.KeyOfKeyword &#124; SyntaxKind.ReadonlyKeyword &#124; SyntaxKind.UniqueKeyword | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceExport | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NamespaceImport | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.NumericLiteral &#124; SyntaxKind.BigIntLiteral &#124; SyntaxKind.StringLiteral &#124; SyntaxKind.JsxText &#124; SyntaxKind.JsxTextAllWhiteSpaces &#124; SyntaxKind.RegularExpressionLiteral &#124; SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.ObjectLiteralExpression | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.StringLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Unknown &#124; SyntaxKind.NumericLiteral &#124; SyntaxKind.BigIntLiteral &#124; SyntaxKind.StringLiteral &#124; SyntaxKind.JsxText &#124; SyntaxKind.RegularExpressionLiteral &#124; SyntaxKind.NoSubstitutionTemplateLiteral | 1 | 0 | not in pinned table |  | False |
| Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 | 0 | not in pinned table |  | False |
| Refused | arguments | 1 | 0 | not in pinned table |  | False |
| Refused | inherited library member compare read as an own field | 1 | 1 | not in pinned table |  | False |
| Refused | inherited library member repeat read as an own field | 1 | 1 | not in pinned table |  | False |
| Refused | optional property Low in Partial<Levels> absent from structural source {}, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property _propertyAccessExpressionLikeQualifiedNameBrand in PropertyAccessEntityNameExpression absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source OptionsBase, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitAny" &#124; "strict">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "noImplicitThis" &#124; "strict">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictBindCallApply">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictBuiltinIteratorReturn">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictFunctionTypes">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictNullChecks">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "strictPropertyInitialization">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source Pick<CompilerOptions, "strict" &#124; "useUnknownInCatchVariables">, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property all in CompilerOptions absent from structural source TypeAcquisition, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property constraint in TypeParameter absent from structural source ObjectType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property createProgram in IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram> absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property default in TypeParameter absent from structural source IndexedAccessType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in BinaryExpression absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in ComputedPropertyName absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in JSDocThisTag absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property id in PrivateIdentifier absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property isInvalidated in CachedResolvedModuleWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedModuleWithFailedLookupLocations, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property jsDocCache in JSDocArray absent from structural source JSDoc[], which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property localSymbol in Declaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source IntersectionType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property members in ObjectType absent from structural source UnionType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in ExportAssignment absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in ParameterDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in PropertyDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property modifiers in SetAccessorDeclaration absent from structural source never, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property outputDts in ResolvedRefAndOutputDts absent from structural source ResolvedRefAndSource, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property packageRootPath in { moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; } absent from structural source { moduleFileToTry: string; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property preserve in FileReference absent from structural source { resolutionMode: ModuleKind.CommonJS &#124; ModuleKind.ESNext; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property reportsUnnecessary in ReusableDiagnostic absent from structural source ReusableDiagnosticRelatedInformation, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in MapIterator<BuilderFileEmit> absent from structural source MapIterator<BuilderFileEmit>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in MapIterator<[string, IntrinsicTypeKind]> absent from structural source MapIterator<[string, IntrinsicTypeKind]>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in MapIterator<[string, TypeFacts]> absent from structural source MapIterator<[string, TypeFacts.TypeofNEString &#124; TypeFacts.TypeofNENumber &#124; TypeFacts.TypeofNEBigInt &#124; TypeFacts.TypeofNEBoolean &#124; TypeFacts.TypeofNESymbol &#124; TypeFacts.TypeofNEObject &#124; TypeFacts.TypeofNEFunction &#124; TypeFacts.NEUndefined]>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property return in MapIterator<[string, WatchDirectoryFlags]> absent from structural source MapIterator<[string, WatchDirectoryFlags]>, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property skipLogging in ErrorOutputContainer absent from structural source { errors?: Diagnostic[]; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property source in ResolvedRefAndSource absent from structural source ResolvedRefAndOutputDts, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property target in AnonymousType absent from structural source IntrinsicType, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property typeArguments in NodeWithTypeArguments absent from structural source ArrayTypeNode, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string &#124; undefined; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property types in { types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ... absent from structural source { [index: string]: string &#124; { value: string; pos: number; end: number; }; }, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | optional property watchFile in WatchOptions absent from structural source {}, which can hide fields | 1 | 0 | not in pinned table |  | False |
| Refused | overload 1 of arrayFrom result U[] cannot be served by implementation result (T &#124; U)[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of cloneNode type parameter T cannot satisfy implementation parameter T | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of combine result T[] &#124; undefined cannot be served by implementation result T &#124; T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of compact parameter array cannot be served by implementation parameter array | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of convertOptionsFromJson result WatchOptions &#124; undefined cannot be served by implementation result CompilerOptions &#124; TypeAcquisition &#124; WatchOptions &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createImmediatelyInvokedArrowFunction result ImmediatelyInvokedArrowFunction cannot be served by implementation result Mutable<CallExpression> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of createImmediatelyInvokedFunctionExpression result ImmediatelyInvokedFunctionExpression cannot be served by implementation result Mutable<CallExpression> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of ensureTrailingDirectorySeparator result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of evaluate result EvaluatorResult<string &#124; undefined> cannot be served by implementation result EvaluatorResult<string &#124; number &#124; undefined> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of filter result U[] cannot be served by implementation result readonly T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of find result U &#124; undefined cannot be served by implementation result T &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of findLast result U &#124; undefined cannot be served by implementation result T &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getDirectoryPath result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getGlobalBuiltinTypes result ObjectType[] cannot be served by implementation result Type[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSuperContainer result SuperContainer &#124; undefined cannot be served by implementation result SuperContainerOrFunctions &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction &#124; ComputedPropertyName &#124; ThisContainer | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of groupBy type parameter U cannot satisfy implementation parameter K | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract<ClassDeclaration, Pick<...>> &#124; Extract<...> cannot be served by implementation result ClassLikeDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of mergeLexicalEnvironment result NodeArray<Statement> cannot be served by implementation result Statement[] &#124; NodeArray<Statement> | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of or type parameter P cannot satisfy implementation parameter T | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseExpectedToken parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseExpectedTokenJSDoc parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseExpectedTokenJSDoc result Token<TKind> cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseModifiers result NodeArray<Modifier> &#124; undefined cannot be served by implementation result NodeArray<ModifierLike> &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseNamedImportsOrExports result NamedImports cannot be served by implementation result NamedImportsOrExports | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseOptionalToken parameter t cannot be served by implementation parameter t | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of parseOptionalTokenJSDoc result Token<TKind> cannot be served by implementation result Node &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of removeTrailingDirectorySeparator result Path cannot be served by implementation result string | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replaceDecoratorsAndModifiers result T cannot be served by implementation result ClassDeclaration &#124; ClassExpression &#124; GetAccessorDeclaration &#124; MethodDeclaration &#124; ParameterDeclaration &#124; PropertyDeclaration &#124; SetAccessorDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replaceModifiers result T cannot be served by implementation result ArrowFunction &#124; ClassDeclaration &#124; ClassExpression &#124; ConstructorDeclaration &#124; ConstructorTypeNode &#124; ... 19 more ... &#124; VariableStatement | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of replacePropertyName result T cannot be served by implementation result GetAccessorDeclaration &#124; MethodDeclaration &#124; MethodSignature &#124; PropertyAssignment &#124; PropertyDeclaration &#124; PropertySignature &#124; SetAccessorDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of sameFlatMap result T[] cannot be served by implementation result readonly T[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of singleOrMany result T &#124; T[] cannot be served by implementation result T &#124; readonly T[] &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipOuterExpressions result T cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipParentheses result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of symbolToName result Identifier cannot be served by implementation result EntityName | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of transformNamedEvaluation result Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> &#124; ... 7 more ... &#124; Extract<...> cannot be served by implementation result BinaryExpression &#124; BindingElement &#124; ExportAssignment &#124; ParameterDeclaration &#124; PropertyAssignment &#124; PropertyDeclaration &#124; ShorthandPropertyAssignment &#124; VariableDeclaration | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | not in pinned table |  | False |
| Refused | overload 1 of writeTokenText result void cannot be served by implementation result number | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of getParseTreeNode result T &#124; undefined cannot be served by implementation result Node &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of group parameter resultSelector cannot be served by implementation parameter resultSelector | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of skipOuterExpressions result Expression cannot be served by implementation result Node | 1 | 1 | not in pinned table |  | False |
| Refused | overload 2 of toArray parameter value cannot be served by implementation parameter value | 1 | 1 | not in pinned table |  | False |
| Refused | overload 3 of getGlobalType result GenericType cannot be served by implementation result ObjectType &#124; undefined | 1 | 1 | not in pinned table |  | False |
| Refused | sort without a comparator | 1 | 1 | not in pinned table |  | False |

Limits: checker-diagnosed nested bodies remain excluded. Failed compounds may hide children, and signature/prologue failures may hide a whole body; boundaries retain each observed span. Isolated reading-X findings can be context-sensitive after rollback. Generic units lack invented specializations. Final module order, ownership and backends are not measured. Count reductions do not prove that every removed site compiles.
