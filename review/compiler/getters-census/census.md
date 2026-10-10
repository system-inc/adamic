# Getter and setter census

No main miscompile was observed. **No, all 84 do not lower in these faithful reduced contexts:** 80 declarations are blocked loudly, and four are represented by two passing shape fixtures. A third passing fixture isolates the common binary getter after expanding its memoization helper. This is a census of reductions, not a claim that the complete upstream compiler compiles.

Adamic baseline: `72ad75effe2dbcc0deedc3e7e34a8d5e29670eac`, current origin/main at branch creation. TypeScript pin: `050880ce59e30b356b686bd3144efe24f875ebc8`. TypeScript API: 6.0.3. Node: 24.19.0. clang: 20.1.8. Go: 1.27.1. `nproc`: 5, cgroup quota 4 CPUs.

The AST walk found **84 accessor declarations, 82 get and 2 set**. There are 82 distinct accessor properties: 80 getter-only properties and two getter/setter pairs. The claimed total differs by zero. The file breakdown is 72 in factory/nodeFactory.ts, 3 in sourcemap.ts, 3 in checker.ts, **4 declarations forming 2 pairs** in transformer.ts, and 1 each in core.ts and transformers/utilities.ts. All except the IdentifierNameMap class getter are in object literals.

The sparse clone hit its 90-second hard limit. The pinned archive succeeded instead, with SHA256 `a0d93407d5797dc26e03badda0fdd6915e465538fa31f0e30d949d08fa64dce8`. `ast-census.cjs` walks createProgram source files with ts.forEachChild and isGetAccessorDeclaration/isSetAccessorDeclaration. No grep determines the census. The checker supplies complete return signatures and contextual interfaces. To reproduce after installing the locked stage3/api dependencies and extracting the pinned archive:

```sh
node review/compiler/getters-census/ast-census.cjs   "$PWD/stage3/api/node_modules/typescript" /path/to/TypeScript-050880ce /tmp/sites.json
```

A rerun produced byte-identical sites.json (`cmp`, exit 0). Shape counts sum to 84. `sites.json` preserves bodies, owner member kinds, exact checker return types, and every matching named property read/write found in src/compiler and src/tsc. Matching uses property declarations as well as symbols, so instantiated generic interface receivers are included.

Observed callers use direct property syntax, generally through NodeFactory, IterationTypes, MappingsDecoder, TransformationContext or Set interfaces; the class getter has IdentifierNameMap callers. No matched receiver in this pinned scope is a union. The class fixture additionally exercises an interface and a structural union containing a plain field. A zero reference count means no matching named caller in the scanned scope, not an unused public API. Set.size candidates include other implementations of the same interface, so candidate callers are not proven calls to createSet. Computed dynamic keys, external callers and runtime patchers are not proven absent. No named writes to getter-only properties were found. The two hook properties have 12 and 10 named writes respectively; paired declarations share the same reference set.

Reductions preserve accessor placement, supporting cache/closure behavior and ordinary methods sharing the literal. Large node types become small Node records; comments become strings instead of NodeArray unions; SyntaxKind constants become numbers. Generic memoizeOne retains a callback, cache lookup, cache write and lazy factory function. This deliberately retains its captured-callback refusal rather than silently removing the blocker. The lazy-object reduction retains a callback referring forward to the factory it accompanies. Debug.fail becomes a module function that throws Error, retaining the typed throwing getter. These are conservative sound reductions, not copied cohere sources.

Observed stops (the exact diagnostic text after the fixture location):

| Code | Diagnostic |
|---|---|
| R1 | stage 0 can't lower a non-accessor method in an accessor literal yet |
| R2 | Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| R3 | stage 0 can't lower reading factory yet |
| R4 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |

Each site row below inherits its shape witness's observed verdict. This says whether the reduced site context lowers on main. It does not attribute every stop to the descriptor machinery: R2 is in the helper's captured callback, R3 in its forward capture, and R4 in the cast. The inline binary control demonstrates that its accessor itself lowers when the helper blocker is removed by reduction.

| Site | Kind/name | Shape | Exact return/input type | Named callers / writes | Lowers on main | Refusal or failure |
|---|---|---|---|---|---|---|
| src/compiler/core.ts:1707 | get size | closure_number_methods | `number` | 18 / 0 | no | R1 |
| src/compiler/factory/nodeFactory.ts:514 | get parenthesizer | lazy_object | `import("../types.js").ParenthesizerRules` | 1 / 0 | no | R3 |
| src/compiler/factory/nodeFactory.ts:517 | get converters | lazy_object | `import("../types.js").NodeConverters` | 10 / 0 | no | R3 |
| src/compiler/factory/nodeFactory.ts:802 | get createJSDocAllType | primary_function | `() => Mutable<JSDocAllType>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:805 | get createJSDocUnknownType | primary_function | `() => Mutable<JSDocUnknownType>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:808 | get createJSDocNonNullableType | optional_boolean_function | `(type: TypeNode, postfix?: boolean) => JSDocNonNullableType` | 2 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:811 | get updateJSDocNonNullableType | update_node_function | `(node: JSDocNonNullableType, type: TypeNode) => JSDocNonNullableType` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:814 | get createJSDocNullableType | optional_boolean_function | `(type: TypeNode, postfix?: boolean) => JSDocNullableType` | 2 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:817 | get updateJSDocNullableType | update_node_function | `(node: JSDocNullableType, type: TypeNode) => JSDocNullableType` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:820 | get createJSDocOptionalType | unary_node_function | `(type: TypeNode) => JSDocOptionalType` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:823 | get updateJSDocOptionalType | update_node_function | `(node: JSDocOptionalType, type: TypeNode) => JSDocOptionalType` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:826 | get createJSDocVariadicType | unary_node_function | `(type: TypeNode) => JSDocVariadicType` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:829 | get updateJSDocVariadicType | update_node_function | `(node: JSDocVariadicType, type: TypeNode) => JSDocVariadicType` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:832 | get createJSDocNamepathType | unary_node_function | `(type: TypeNode) => JSDocNamepathType` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:835 | get updateJSDocNamepathType | update_node_function | `(node: JSDocNamepathType, type: TypeNode) => JSDocNamepathType` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:877 | get createJSDocTypeTag | optional_type_function | `(tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => Mutable<JSDocTypeTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:880 | get updateJSDocTypeTag | update_type_function | `(node: JSDocTypeTag, tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => JSDocTypeTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:883 | get createJSDocReturnTag | optional_type_function | `(tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => Mutable<JSDocReturnTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:886 | get updateJSDocReturnTag | update_type_function | `(node: JSDocReturnTag, tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => JSDocReturnTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:889 | get createJSDocThisTag | optional_type_function | `(tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => Mutable<JSDocThisTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:892 | get updateJSDocThisTag | update_type_function | `(node: JSDocThisTag, tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => JSDocThisTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:895 | get createJSDocAuthorTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocAuthorTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:898 | get updateJSDocAuthorTag | update_comment_function | `(node: JSDocAuthorTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocAuthorTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:901 | get createJSDocClassTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocClassTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:904 | get updateJSDocClassTag | update_comment_function | `(node: JSDocClassTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocClassTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:907 | get createJSDocPublicTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocPublicTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:910 | get updateJSDocPublicTag | update_comment_function | `(node: JSDocPublicTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocPublicTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:913 | get createJSDocPrivateTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocPrivateTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:916 | get updateJSDocPrivateTag | update_comment_function | `(node: JSDocPrivateTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocPrivateTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:919 | get createJSDocProtectedTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocProtectedTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:922 | get updateJSDocProtectedTag | update_comment_function | `(node: JSDocProtectedTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocProtectedTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:925 | get createJSDocReadonlyTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocReadonlyTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:928 | get updateJSDocReadonlyTag | update_comment_function | `(node: JSDocReadonlyTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocReadonlyTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:931 | get createJSDocOverrideTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocOverrideTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:934 | get updateJSDocOverrideTag | update_comment_function | `(node: JSDocOverrideTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocOverrideTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:937 | get createJSDocDeprecatedTag | optional_comment_function | `(tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => Mutable<JSDocDeprecatedTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:940 | get updateJSDocDeprecatedTag | update_comment_function | `(node: JSDocDeprecatedTag, tagName: Identifier \| undefined, comment?: NodeArray<JSDocComment>) => JSDocDeprecatedTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:943 | get createJSDocThrowsTag | optional_type_function | `(tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => Mutable<JSDocThrowsTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:946 | get updateJSDocThrowsTag | update_type_function | `(node: JSDocThrowsTag, tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => JSDocThrowsTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:949 | get createJSDocSatisfiesTag | optional_type_function | `(tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => Mutable<JSDocSatisfiesTag>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:952 | get updateJSDocSatisfiesTag | update_type_function | `(node: JSDocSatisfiesTag, tagName: Identifier \| undefined, typeExpression?: JSDocTypeExpression, comment?: NodeArray<JSDocComment>) => JSDocSatisfiesTag` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1022 | get createComma | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 14 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1025 | get createAssignment | overloaded_function | `{ (left: ObjectLiteralExpression \| ArrayLiteralExpression, right: Expression): import("../types.js").DestructuringAssignment; (left: Expression, right: Expression): import("../types.js").AssignmentExpression<import("../types.js").EqualsToken>; }` | 143 / 0 | no | R4 |
| src/compiler/factory/nodeFactory.ts:1028 | get createLogicalOr | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 5 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1031 | get createLogicalAnd | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 15 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1034 | get createBitwiseOr | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1037 | get createBitwiseXor | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1040 | get createBitwiseAnd | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1043 | get createStrictEquality | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 5 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1046 | get createStrictInequality | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 9 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1049 | get createEquality | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1052 | get createInequality | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1055 | get createLessThan | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 3 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1058 | get createLessThanEquals | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1061 | get createGreaterThan | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1064 | get createGreaterThanEquals | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1067 | get createLeftShift | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1070 | get createRightShift | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1073 | get createUnsignedRightShift | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1076 | get createAdd | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1079 | get createSubtract | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 1 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1082 | get createMultiply | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1085 | get createDivide | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1088 | get createModulo | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1091 | get createExponent | binary_function | `(left: Expression, right: Expression) => Mutable<BinaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1094 | get createPrefixPlus | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1097 | get createPrefixMinus | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1100 | get createPrefixIncrement | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1103 | get createPrefixDecrement | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1106 | get createBitwiseNot | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1109 | get createLogicalNot | unary_function | `(operand: Expression) => Mutable<PrefixUnaryExpression>` | 8 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1112 | get createPostfixIncrement | unary_function | `(operand: Expression) => Mutable<PostfixUnaryExpression>` | 3 / 0 | no | R2 |
| src/compiler/factory/nodeFactory.ts:1115 | get createPostfixDecrement | unary_function | `(operand: Expression) => Mutable<PostfixUnaryExpression>` | 0 / 0 | no | R2 |
| src/compiler/checker.ts:2187 | get yieldType | throwing_object | `Type` | 5 / 0 | yes | agrees |
| src/compiler/checker.ts:2190 | get returnType | throwing_object | `Type` | 2 / 0 | yes | agrees |
| src/compiler/checker.ts:2193 | get nextType | throwing_object | `Type` | 4 / 0 | yes | agrees |
| src/compiler/sourcemap.ts:474 | get pos | closure_number_next | `number` | 0 / 0 | no | R1 |
| src/compiler/sourcemap.ts:477 | get error | closure_union_next | `string \| undefined` | 2 / 0 | no | R1 |
| src/compiler/sourcemap.ts:480 | get state | snapshot_next | `Required<Mapping>` | 0 / 0 | no | R1 |
| src/compiler/transformers/utilities.ts:392 | get size | class_map | `number` | 2 / 0 | yes | agrees |
| src/compiler/transformer.ts:295 | get onSubstituteNode | callback_pair_methods | `(hint: EmitHint, node: Node) => Node` | 25 / 12 | no | R1 |
| src/compiler/transformer.ts:298 | set onSubstituteNode | callback_pair_methods | `(hint: EmitHint, node: Node) => Node` | 25 / 12 | no | R1 |
| src/compiler/transformer.ts:303 | get onEmitNode | higher_callback_pair_methods | `(hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void` | 21 / 10 | no | R1 |
| src/compiler/transformer.ts:306 | set onEmitNode | higher_callback_pair_methods | `(hint: EmitHint, node: Node, emitCallback: (hint: EmitHint, node: Node) => void) => void` | 21 / 10 | no | R1 |

Every fixture is under internal/oracle/testdata and begins with its real source file:line. All observations have exit 0 and empty stderr on Node. For passing fixtures, JavaScript backend, release native, sanitized native and LeakSanitizer agree, with no leak report. For refused fixtures those backends are not reached. The following outputs use literal \n to denote newline bytes.

| Shape | Declarations | Placement and behavior | Fixture | Node stdout | JavaScript / release / sanitized | Result |
|---|---:|---|---|---|---|---|
| closure_number_methods | 1 | Returned literal; getter only; mutable closure number; sibling add/has/clear methods; Set interface reads. | getters_census_closure_number_methods.a | `0\n1\n` | not reached | R1 |
| lazy_object | 2 | Local factory literal returned as NodeFactory; getter only; memoized rules/converters object whose callbacks refer forward to factory; no accessor writes. | getters_census_lazy_object.a | `constructed\ninitialize\n9\n10\n` | not reached | R3 |
| primary_function | 2 | Local returned factory literal; getter only; memoizeOne cache; zero-argument function returning a node. | getters_census_primary_function.a | `constructed\ninitialize 7\n7\n7\ninitialize 7\n7\n` | not reached | R2 |
| optional_boolean_function | 2 | Local returned factory literal; getter only; memoizeOne cache; node input and optional boolean; node result. | getters_census_optional_boolean_function.a | `constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n` | not reached | R2 |
| update_node_function | 5 | Local returned factory literal; getter only; memoizeOne cache; existing node and type node inputs; node result. | getters_census_update_node_function.a | `constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n` | not reached | R2 |
| unary_node_function | 3 | Local returned factory literal; getter only; memoizeOne cache; type node input; node result. | getters_census_unary_node_function.a | `constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n` | not reached | R2 |
| optional_type_function | 5 | Local returned factory literal; getter only; memoizeOne cache; optional name/type/comment inputs; node result. | getters_census_optional_type_function.a | `constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n` | not reached | R2 |
| update_type_function | 5 | Local returned factory literal; getter only; memoizeOne cache; existing node plus optional name/type/comment inputs; node result. | getters_census_update_type_function.a | `constructed\ninitialize 7\n15\n15\ninitialize 7\n15\n` | not reached | R2 |
| optional_comment_function | 8 | Local returned factory literal; getter only; memoizeOne cache; name and optional comment inputs; node result. | getters_census_optional_comment_function.a | `constructed\ninitialize 7\n10\n10\ninitialize 7\n10\n` | not reached | R2 |
| update_comment_function | 8 | Local returned factory literal; getter only; memoizeOne cache; existing node, name and optional comment inputs; node result. | getters_census_update_comment_function.a | `constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n` | not reached | R2 |
| binary_function | 23 | Local returned factory literal; getter only; memoizeOne cache; two expression inputs; function returning a node. | getters_census_binary_function.a | `constructed\ninitialize 7\n12\n12\ninitialize 7\n12\n` | not reached | R2 |
| overloaded_function | 1 | Local returned factory literal; getter only; memoizeOne cache; function cast to an overloaded interface property. | getters_census_overloaded_function.a | `10\n` | not reached | R4 |
| unary_function | 8 | Local returned factory literal; getter only; memoizeOne cache; one expression input; function returning a node. | getters_census_unary_function.a | `constructed\ninitialize 7\n9\n9\ninitialize 7\n9\n` | not reached | R2 |
| throwing_object | 3 | Local sentinel literal; getter only; typed Type return via Debug.fail (never); no reads from this or mutable closure. | getters_census_throwing_object.a | `constructed\nNot supported\n` | same stdout, exit 0, empty stderr | agrees |
| closure_number_next | 1 | Returned decoder literal; getter only; mutable closure number; sibling next method; MappingsDecoder interface. | getters_census_closure_number_next.a | `0\n1\n` | not reached | R1 |
| closure_union_next | 1 | Returned decoder literal; getter only; mutable closure string \| undefined; sibling next method; MappingsDecoder interface. | getters_census_closure_union_next.a | `undefined\nbad mapping\n` | not reached | R1 |
| snapshot_next | 1 | Returned decoder literal; getter only; call constructs a fresh Required<Mapping> snapshot from mutable closure counters; sibling next method. | getters_census_snapshot_next.a | `0/1\n` | not reached | R1 |
| class_map | 1 | Class getter only; number read from this._map.size; map is written by set/delete/clear; direct class property reads. | getters_census_class_map.a | `0/0/0\n1/1/1\n9\n` | same stdout, exit 0, empty stderr | agrees |
| callback_pair_methods | 2 | Local context literal; paired get/set; mutable closure callback (hint,node) => Node; initialization assertions and sibling method; interface reads/writes. | getters_census_callback_pair_methods.a | `3\n13\ndone\ninitialized\n` | not reached | R1 |
| higher_callback_pair_methods | 2 | Local context literal; paired get/set; mutable closure callback taking an emit callback, returning void; assertions and sibling method; interface reads/writes. | getters_census_higher_callback_pair_methods.a | `3\n13\ndone\ninitialized\n` | not reached | R1 |
| binary_inline | supplemental | Returned literal; memoizeOne expanded to a named cache worker; getter returns a closure and stays lazy, cache per factory. | getters_census_binary_inline.a | `constructed\ninitialize 7\n12\n16\ninitialize 7\n10\n` | same stdout, exit 0, empty stderr | agrees |

Refused witnesses are held by individual top-level TestGettersCensus functions checking both the exact message and the typed Refused/NotYet error. NotYet witnesses also register as non-lowering fixtures in TestNativeAgreesWithNode. Refused witnesses cannot use that list because its existing false branch only accepts NotYet; they stay outside the passing set and have dedicated tests. No miscompile required a t.Skip.

The mutant in eager-getter.patch temporarily changes literal accessor lowering: while constructing the most common shape's createComma literal, it calls the getter and stores that returned closure in a private field. It retains the normal getter descriptor afterward. This is enough to break lazy initialization observably. run-mutant.py restores the compiler even on failure. The exact-selector TestNativeAgreesWithNode run exits 1 because stdout differs, with successful compilation, exit 0, empty stderr on every backend, and no sanitizer or clang-warning failure:

```text
Node:       constructed\ninitialize 7\n12\n16\ninitialize 7\n10\n
JavaScript: initialize 7\nconstructed\n12\n16\ninitialize 7\n10\n
sanitized:  initialize 7\nconstructed\n12\n16\ninitialize 7\n10\n
```

The release comparison also passes against sanitized; the failure is the source oracle comparison. Production source is restored byte for byte, with no lowering change delivered. The original memoizeOne-backed binary witness is loudly refused before accessor execution, so the supplemental expanded-helper witness supplies the behavioral mutant check.

Setup attempt one (`timeout 180 bash cloud/setup.sh`) exited 124 in build-cache warming: go ready 0.068s, Node ready 0.080s, clang ready 0.476s, markdown ready 1.590s, submodules ready 23.758s, shared cache ready 30.790s. Retry (`ADAMIC_GOCACHE_OFF=1 timeout 240 bash cloud/setup.sh`) also exited 124 during go list exports: Node ready 0.027s, Go ready 0.028s, markdown ready 0.070s, submodules ready 0.074s, clang ready 0.164s, shared cache off 0.166s. Workaround: source the generated /workspace/adamic-tools/env.sh with the optional shared cache disabled, build only the focused test dependencies. Those builds and all focused tests succeed. The first counts run found missing @types/node 25.3.3 for existing host fixtures; npm ci --prefix stage3/api installed the locked dependencies and its retry is recorded separately.

Commands run from the repository root, with source /workspace/adamic-tools/env.sh before Go commands. Each test's full output went to its named log:

```sh
# Initial probe, all shapes marked lowering to expose their actual diagnostics (exit 1).
timeout 180 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^getters_census_[a-z_]+\.a$' -count=1 -timeout 90s -parallel 4 -v
# Final focused checks, uncached (exit 0, 2.163s).
ADAMIC_GATE_UNCACHED=1 timeout 90 go test ./internal/oracle -run '^TestGettersCensus|^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^getters_census_[a-z_]+\.a$' -count=1 -timeout 90s -parallel 4 -v
# Mutant runner (exit 0 because the intended test fails).
timeout 150 python3 review/compiler/getters-census/run-mutant.py
# Counts and repository guard.
timeout 90 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts
timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
```

Final new top-level leaf durations, all below 60 seconds including their fixture loading/build/run work:

| Test | Seconds |
|---|---:|
| TestGettersCensusClosureNumberMethods | 0.28 |
| TestGettersCensusOverloadedFunction | 0.29 |
| TestGettersCensusOptionalTypeFunction | 0.31 |
| TestGettersCensusUpdateCommentFunction | 0.22 |
| TestGettersCensusBinaryFunction | 0.21 |
| TestGettersCensusOptionalCommentFunction | 0.20 |
| TestGettersCensusClosureNumberNext | 0.18 |
| TestGettersCensusClassMap | 0.69 |
| TestGettersCensusUpdateTypeFunction | 0.21 |
| TestGettersCensusSnapshotNext | 0.23 |
| TestGettersCensusClosureUnionNext | 0.20 |
| TestGettersCensusUnaryFunction | 0.20 |
| TestGettersCensusOptionalBooleanFunction | 0.22 |
| TestGettersCensusUnaryNodeFunction | 0.20 |
| TestGettersCensusPrimaryFunction | 0.20 |
| TestGettersCensusThrowingObject | 0.63 |
| TestGettersCensusUpdateNodeFunction | 0.18 |
| TestGettersCensusBinaryInline | 0.71 |
| TestGettersCensusHigherCallbackPairMethods | 0.23 |
| TestGettersCensusLazyObject | 0.22 |
| TestGettersCensusCallbackPairMethods | 0.22 |

Evidence: ast.log, sites.json, shapes.json, refusals.json, node-outputs.json, initial.log, observations.log, final.log, mutant.log, eager-getter.patch, setup.log, setup-off.log, node-types.log, counts.log, counts-final.log and call-target-readers.log. No full package test or full gate was run. The unit advances step 18's accessor coverage toward Outcome 24 by pinning real parser-path shapes and their current stops. It does not implement fixes, port the complete factory, prove dynamic caller reachability, or establish full-program TypeScript compilation.

Counts refresh passed in 88.390s. Only three rows were added; every existing row is unchanged:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| getters_census_class_map.a | 13 | 13 | 4 | 15 | 6 | 0 |
| getters_census_throwing_object.a | 3 | 3 | 5 | 8 | 3 | 0 |
| getters_census_binary_inline.a | 37 | 37 | 29 | 57 | 17 | 0 |

TestCallTargetReaders passed in 38.465s. The new tests read no raw call-target fields.
