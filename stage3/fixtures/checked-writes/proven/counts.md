# Proven writable-view counts

Measured on main 45487a809f89885a3fc651cd590e7dabf31362dc.

98 original sites: 91 no-write, 7 compatible-write; 38 source-type families; 10 receiving proof families.

| Fixture | Node exit | Stage0 | Writable-view checks required after #63x2441 |
| --- | ---: | --- | ---: |
| 01_truthiness.a | 0 | NotYet | 0 |
| 02_control.a | 0 | NotYet | 0 |
| 03_kind_reader.a | 0 | Refused | 0 |
| 04_range_reader.a | 0 | Refused | 0 |
| 05_diagnostic_related.a | 0 | Refused | 0 |
| 06_modifiers_initialize.a | 0 | Refused | 0 |
| 07_assert_clause.a | 0 | Refused | 0 |
| 08_property_initializer.a | 0 | Refused | 0 |
| 09_static_modifiers.a | 0 | Refused | 0 |
| 10_module_parent.a | 0 | Refused | 0 |

No native binaries exist for these current refusals/NotYet outcomes, so allocation/free/retain/release counts are unmeasured. No central oracle counts row was added.

| Source-type family | Sites | Receiving witnesses |
| --- | ---: | --- |
| AssignmentPattern \| undefined | 1 | 01_truthiness.a |
| AutoAccessorPropertyDeclaration \| undefined | 1 | 01_truthiness.a |
| Block \| undefined | 1 | 01_truthiness.a |
| CallExpression \| undefined | 1 | 01_truthiness.a |
| CallLikeExpression \| undefined | 1 | 01_truthiness.a |
| Children | 1 | 04_range_reader.a |
| ClassStaticBlockDeclaration | 1 | 09_static_modifiers.a |
| Declaration | 1 | 01_truthiness.a |
| Declaration \| undefined | 7 | 01_truthiness.a |
| DiagnosticWithLocation | 2 | 05_diagnostic_related.a |
| EntityNameExpression \| undefined | 1 | 01_truthiness.a |
| ExportDeclaration \| undefined | 1 | 01_truthiness.a |
| Expression \| undefined | 6 | 01_truthiness.a |
| FlowNode \| undefined | 4 | 02_control.a |
| GeneratedIdentifier \| GeneratedPrivateIdentifier | 3 | 03_kind_reader.a |
| Identifier \| undefined | 1 | 01_truthiness.a |
| ImportTypeAssertionContainer | 1 | 07_assert_clause.a |
| IndexInfo \| undefined | 1 | 01_truthiness.a |
| IndexSignatureDeclaration | 1 | 06_modifiers_initialize.a |
| JSDocTypeExpression \| undefined | 1 | 01_truthiness.a |
| JsxAttributeLike \| undefined | 1 | 01_truthiness.a |
| MemberName \| undefined | 1 | 01_truthiness.a |
| ModuleName | 1 | 10_module_parent.a |
| Node \| undefined | 5 | 01_truthiness.a |
| NodeArray<Statement> | 1 | 04_range_reader.a |
| ParameterDeclaration \| undefined | 2 | 01_truthiness.a |
| PropertyAccessExpression \| undefined | 1 | 01_truthiness.a |
| PropertyName \| undefined | 1 | 01_truthiness.a |
| PropertySignature | 1 | 08_property_initializer.a |
| Signature \| undefined | 3 | 01_truthiness.a |
| SourceFile \| undefined | 2 | 01_truthiness.a |
| Symbol | 4 | 01_truthiness.a |
| Symbol \| undefined | 18 | 01_truthiness.a |
| Type | 1 | 01_truthiness.a |
| Type \| undefined | 10 | 01_truthiness.a |
| TypeNode \| undefined | 7 | 01_truthiness.a |
| TypeParameterDeclaration \| undefined | 1 | 01_truthiness.a |
| readonly TypeParameter[] \| undefined | 1 | 01_truthiness.a |
