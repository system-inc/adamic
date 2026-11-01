| Reason | Before | After | Change |
|---|---:|---:|---:|
| a () => never seen as a (node: ElementAccessExpression &#124; EnumMember &#124; PropertyAccessExpression) => string &#124; number &#124; undefined (one keeps something weakly that the other keeps strongly) | 0 | 1 | +1 |
| a getter in a checked field contract | 1 | 0 | -1 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 2 | +1 |
| a union or optional field read in a program with record storage | 74 | 77 | +3 |
| a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 1 | 2 | +1 |
| a value of type HasJSDoc | 0 | 1 | +1 |
| a value of type RequireOrImportCall | 0 | 1 | +1 |
| checked predicate membership over an open tag domain | 47 | 50 | +3 |
| checked predicate overload target AssignmentExpression<AmpersandAmpersandEqualsToken &#124; BarBarEqualsToken &#124; EqualsToken &#124; QuestionQuestionEqualsToken> & ... | 0 | 1 | +1 |
| checked predicate overload target BindableStaticElementAccessExpression | 0 | 2 | +2 |
| checked predicate overload target JSDocTag | 0 | 1 | +1 |
| checked predicate overload target PrefixUnaryExpression & { operand: NumericLiteral; } | 0 | 1 | +1 |
| checked predicate overload target PrefixUnaryExpression & { operator: SyntaxKind.PlusToken; operand: NumericLiteral; } | 0 | 1 | +1 |
| checked view union arm awaits views-v3: array element kind | 0 | 1 | +1 |
| writing a checked object field without its source-slot type certificate | 13 | 17 | +4 |
