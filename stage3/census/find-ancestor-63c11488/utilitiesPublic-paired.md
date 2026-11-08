# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 4 | 4 |
| Moved by reason | 1 | 0 |
| New findings | 2 | 1 |
| New sites | 1 | 1 |
| Units whose first stop moved | 3 | 3 |
| Regressions | 0 | 0 |
| before Boundary | 117 | 117 |
| before NotYet | 72 | 72 |
| before Refused | 37 | 36 |
| after Boundary | 115 | 115 |
| after NotYet | 73 | 73 |
| after Refused | 35 | 34 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| JSON.parse: its result's type can't be proven from the text | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 28 | 28 | 28 | 28 | 0 |
| a call through ?. (an optional call) | 1 | 1 | 1 | 1 | 0 |
| a cast the runtime can't check | 17 | 17 | 17 | 17 | 0 |
| a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 | 1 | 1 | 0 |
| a function returning T \| undefined | 1 | 1 | 1 | 1 | 0 |
| a read narrowed to undefined of a property a call may write | 1 | 1 | 1 | 1 | 0 |
| a readonly ["cs", "de", "es", "fr", "it", "ja", "ko", "pl", "pt-br", "ru", "tr", "zh-cn", "zh-tw"] seen as a readonly string[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) | 1 | 1 | 1 | 1 | 0 |
| a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 | 1 | 1 | 1 | 0 |
| a value of type JSDocArray \| undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 1 | 1 | 1 | 1 | 0 |
| an object refinement using an open numeric enum as a literal tag | 4 | 4 | 3 | 3 | 0 |
| an unproven predicate argument for parameter predicate (argument "predicate") | 2 | 2 | 2 | 2 | 0 |
| an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression | 4 | 4 | 4 | 4 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.VariableDeclaration | 1 | 1 | 1 | 1 | 0 |
| checked predicate membership over an open tag domain | 9 | 9 | 9 | 9 | 0 |
| checked predicate overload target BindableStaticElementAccessExpression | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } | 0 | 1 | 0 | 1 | 0 |
| checked predicate overload target ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target JSDocTag | 1 | 1 | 1 | 1 | 0 |
| checked predicate overload target NamedDeclaration & { name: DeclarationName; } | 1 | 1 | 1 | 1 | 0 |
| failed statement; state rolled back; continued at next statement or declaration | 107 | 107 | 107 | 107 | 0 |
| function signature or prologue failed; continued at next declaration | 10 | 8 | 10 | 8 | 2 |
| optional property id in Node absent from structural source never, which can hide fields | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) | 1 | 0 | 1 | 0 | 1 |
| overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | 1 | 0 | 1 | 0 | 1 |
| overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node | 1 | 1 | 1 | 1 | 0 |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 1 | 1 | 1 | 1 | 0 |
| reading commentRanges | 1 | 1 | 1 | 1 | 0 |
| reading expr | 1 | 1 | 1 | 1 | 0 |
| reading expression | 1 | 1 | 1 | 1 | 0 |
| reading flags | 1 | 1 | 1 | 1 | 0 |
| reading jsDoc | 1 | 1 | 1 | 1 | 0 |
| reading last | 1 | 1 | 1 | 1 | 0 |
| reading name | 1 | 1 | 1 | 1 | 0 |
| reading paramIdx | 1 | 1 | 1 | 1 | 0 |
| reading previousSibling | 1 | 1 | 1 | 1 | 0 |
| reading sig | 1 | 1 | 1 | 1 | 0 |
| reading tag | 3 | 3 | 3 | 3 | 0 |
| reading tags | 2 | 2 | 2 | 2 | 0 |
| reading type | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field without its source-slot type certificate | 8 | 8 | 8 | 8 | 0 |

## Raw Findings

Moved by reason:

- 1: failed statement; state rolled back; continued at next statement or declaration -> failed statement; state rolled back; continued at next statement or declaration

New findings:

- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:1562:1`: NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }
- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:1567:5`: Boundary: failed statement; state rolled back; continued at next statement or declaration

Units whose first stop moved:

- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:786:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked)
- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:786:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked)
- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:1562:1`: NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:1562:1`: NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }

Units whose first stop moved:

- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:786:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked)
- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:786:1`: Refused: overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked)
- `/workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:1562:1`: NotYet: checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
