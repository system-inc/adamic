Road 3 for family 3 (#bbssdcy, Adamic's Nov 24 16:4xZ ruling 6): `declaration` takes the base `Declaration` its callers pass, with an `isVariableLike` guard before `declarationBelongsToPrivateAmbientMember`. Compiler accepts this head (her Nov 24 20:47:38Z pick, under Adamic's widened split of Nov 22 20:50:45Z, on its value identity) after its review and Loom's unit census.

Ruling 6 called for "the variable-like union its callers pass" and a type-only adaptation. Round 3 measured TS2379 at adapted `checker.ts:12513:59` for both union candidates. This head uses `Declaration` and changes one line of JavaScript. Compiler chose that deviation and tells Adamic so she can object. Owner: @system_adamic_seats builds and keeps this adaptation, as every `99-seats-*` directory, and @system_adamic_compiler_correctness reads it at each tsc re-pin, as she does every adapter.

The adapter changes `declaration: any` to `declaration: Declaration` and adds the `isVariableLike` guard before the private ambient helper call. The guard continues to report implicit any for excluded declarations. This serves rulings 39, 290 and 406.

One difference no read of tsc observes, kept for a re-pin that makes it public: for the two JSDoc tags the guard skips, stock's `declarationBelongsToPrivateAmbientMember` fills the tag's `modifierFlagsCache` during checking, and the guard leaves it unfilled until the first public read. Read raw from a host after compilation, the `@callback` `@param` tag holds 805306368 in stock and 536870912 with the guard; the `@property` tag converges by the end of checking. The field is `@internal` (`types.ts:945`), and `getCombinedModifierFlags` and `getEffectiveModifierFlags` fill it the same way and return 0 on both sides, so Compiler ruled it outside the split's identity (Nov 24, #bbssdcy).

| Call | Enclosing function | Static type | Kind | reportErrors |
| --- | --- | --- | --- | --- |
| `checker.ts:12462:16` | `getWidenedTypeForVariableLikeDeclaration` | `VariableDeclaration \| ParameterDeclaration \| BindingElement \| PropertySignature \| PropertyDeclaration \| JSDocPropertyLikeTag` | `SyntaxKind` | `reportErrors` |
| `checker.ts:12611:20` | `getTypeOfVariableOrParameterOrPropertyWorker` | `Declaration` | `SyntaxKind` | `Omitted` |
| `checker.ts:45211:51` | `checkVariableLikeDeclaration` | `VariableDeclaration \| ParameterDeclaration \| BindingElement \| PropertyDeclaration` | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.BindingElement \| SyntaxKind.VariableDeclaration` | `Omitted` |
| `checker.ts:45218:51` | `checkVariableLikeDeclaration` | `VariableDeclaration \| ParameterDeclaration \| BindingElement \| PropertyDeclaration` | `SyntaxKind.Parameter \| SyntaxKind.PropertyDeclaration \| SyntaxKind.BindingElement \| SyntaxKind.VariableDeclaration` | `Omitted` |

## Gate map v5

- `adapter-structural-pins`: the adapter finds the two functions inside `createTypeChecker`, matches the exact helper and export then branch, and finds the parameter and guard by node. The export branch pin also fixes the omitted `reportErrors` argument. The test uses the adapter's own before and after refs.
- `adapter-output-test`: the test replays the adapter, type-checks the after snapshot, and compares the two emitted compilers on five input programs.
- `adapter-table-mentions`: this edit changes a parameter type and a guard expression. It touches no declared data table.
