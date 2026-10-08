# Remaining root binary rows

Measured with the original table compiler dc6b1529 and the same checker project and source hashes, at combine's actual failing node. Counts join exact `(where, reason)` to the raw table CSV and reject missing or ambiguous failing nodes. They do not count attempted contexts twice.

| Row | Operator | Unique sites |
| --- | --- | ---: |
| a BinaryExpression with a boolean and a number | `&&` | 79 |
| a BinaryExpression with a boolean and a number | `&#124;&#124;` | 37 |
| a BinaryExpression with a value and a number | `&&` | 54 |
| a BinaryExpression with a value and a number | `&#124;&#124;` | 3 |
| a BinaryExpression with a string and a string | `&#124;&#124;` | 22 |
| a BinaryExpression with a string and a string | `=` | 6 |
| a BinaryExpression with a string and a string | `&&` | 3 |

The first five previously measured rows have only their 190 assignment values remaining after the original 1,262 logical/equality sites. The three additional rows contain 198 logical sites and six string assignment values, with no arithmetic or comparison remainder. Their logical and assignment rules already exist on this branch; follow-up groups add exact shape fixtures and mutants rather than claim an additional operator implementation. The six string assignments are separate table sites from the original 190.

| Row | Operator | Left checker type | Right checker type | Result checker type | Sites |
| --- | --- | --- | --- | --- | ---: |
| a BinaryExpression with a boolean and a number | && | boolean | number | number &#124; false | 77 |
| a BinaryExpression with a boolean and a number | &#124;&#124; | boolean | number | number &#124; true | 37 |
| a BinaryExpression with a boolean and a number | && | boolean | NodeCheckFlags | false &#124; NodeCheckFlags | 2 |
| a BinaryExpression with a value and a number | && | Type &#124; undefined | number | number &#124; undefined | 12 |
| a BinaryExpression with a value and a number | && | Symbol &#124; undefined | number | number &#124; undefined | 10 |
| a BinaryExpression with a value and a number | && | Symbol | number | number | 6 |
| a BinaryExpression with a value and a number | && | Node &#124; undefined | number | number &#124; undefined | 4 |
| a BinaryExpression with a value and a number | && | Declaration[] &#124; undefined | number | number &#124; undefined | 2 |
| a BinaryExpression with a value and a number | && | Type | number | number | 2 |
| a BinaryExpression with a value and a number | &#124;&#124; | QuestionDotToken &#124; undefined | number | number &#124; QuestionDotToken | 2 |
| a BinaryExpression with a value and a number | && | ClassLikeDeclaration &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | CodeBlock &#124; undefined | CodeBlockKind | CodeBlockKind &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | Declaration &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | EmitNode &#124; undefined | EmitFlags | EmitFlags &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | EmitNode &#124; undefined | InternalEmitFlags | InternalEmitFlags &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | IndexInfo &#124; undefined | Ternary | Ternary &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | JSDoc &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | JSDocSignature &#124; SignatureDeclaration &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | Node | number | number | 1 |
| a BinaryExpression with a value and a number | && | ObjectType | number | number | 1 |
| a BinaryExpression with a value and a number | && | TsConfigSourceFile &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | Type &#124; undefined | Ternary | Ternary &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | VariableDeclarationList &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | readonly CommentRange[] &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | readonly JsxChild[] | number | number | 1 |
| a BinaryExpression with a value and a number | && | readonly JsxChild[] &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | readonly T[] &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | && | string[] &#124; undefined | number | number &#124; undefined | 1 |
| a BinaryExpression with a value and a number | &#124;&#124; | Node &#124; undefined | number | number &#124; Node | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | "" | string | 5 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | string &#124; undefined | string &#124; undefined | 5 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | string | string | 3 |
| a BinaryExpression with a string and a string | && | string &#124; undefined | string &#124; undefined | string &#124; undefined | 2 |
| a BinaryExpression with a string and a string | &#124;&#124; | __String &#124; undefined | __String | __String | 2 |
| a BinaryExpression with a string and a string | &#124;&#124; | string | "." | string | 2 |
| a BinaryExpression with a string and a string | && | string &#124; undefined | string | string &#124; undefined | 1 |
| a BinaryExpression with a string and a string | = | RedirectsCacheKey &#124; undefined | RedirectsCacheKey | RedirectsCacheKey | 1 |
| a BinaryExpression with a string and a string | = | __String | InternalSymbolName &#124; (string & { __escapedIdentifier: void; }) | InternalSymbolName &#124; (string & { __escapedIdentifier: void; }) | 1 |
| a BinaryExpression with a string and a string | = | __String | __String | __String | 1 |
| a BinaryExpression with a string and a string | = | string | string | string | 1 |
| a BinaryExpression with a string and a string | = | string &#124; undefined | string | string | 1 |
| a BinaryExpression with a string and a string | = | string &#124; undefined | string &#124; undefined | string &#124; undefined | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string | "" | string | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string | "*" | string | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | "6.0" | string | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | "Unexpected node." | string | 1 |
| a BinaryExpression with a string and a string | &#124;&#124; | string &#124; undefined | "x" | string | 1 |
