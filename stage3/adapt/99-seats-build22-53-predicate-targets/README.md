# Family 53 · computed names and statement exclusion

Seats build22-53 for #bbssdcy, at Adamic main ef741d4f0860597ff8d4df5dd042973b2f20819d
and TypeScript v6.0.3
(050880ce59e30b356b686bd3144efe24f875ebc8).
Adamic ruled this adaptation on Nov 22 2026 (#bbssdcy, family 53) and
accepts the head (Nov 26 2026, 22:55:55Z) with the scope written below.
The statement conditional is reshaped under Compiler's 09:06:15Z answer to
keep the base operand type. Adamic's Nov 23 15:12:21Z ruling disallows the
delegate. The site at `transformers/destructuring.ts:185:9` waits for a
proven predicate under a separate ruling. The declaration-host partition
at `utilitiesPublic.ts:894:9` goes to Adamic before it is written.

The ruling rows at this pin are `docs/rulings.md:32,148 to 150,157,259,398`.
Row 32 governs observation of phantom brands. Rows 148 to 150 govern checked
predicates in .ts and when a false-side read requires a check. Row 157 governs
safe reads after a branded predicate claim. Row 259 limits when an exact kind
predicate counts as the comparison. Row 398 records the meaning of NotYet.
These edits change tsc's source expressions and leave those compiler rules in place.

The scope of the accept, in Adamic's words: the read sequence differs only
for AST objects whose properties aren't plain data, meaning a getter on
`kind`, `name` or `parent`, or a node missing a property its declaration
requires. The Nov 23 outside-contract rule already covers both: the contract
is the public `.d.ts`, which declares those as data properties, required
where `name` is required. A caller who hands tsc a getter where the `.d.ts`
declares data, or a node without its required `name`, is outside it. A
before-transform callback returning such nodes is that caller. On
compiler-owned nodes and factory nodes, the reviewer showed every returned
reference and key order identical. The witness of what the edits drop is
review-w read 10's probe: an observing getter counted 2 `kind` reads of the
accessor in the base against 1 in the head, and 8 `parent` reads of a block
against 4 (5 against 4 in its throwing fixture), and a throwing getter threw
at the base's second `kind` and fifth `parent` read where the head returned
the same node and `undefined`. Its reproduction through the public
`transpileModule`, with a before-transform callback installing the observer,
threw `accessor kind observer` in the base while the head emitted the same
259 bytes with no diagnostics.

## Computed accessor name

At the meter's `utilities.ts:6875:9` in `getAllAccessorDeclarations`,
`isDynamicName(accessor.name)` replaces `hasDynamicName(accessor)`.
The predicate obtains the name and calls the same helper:

```ts
export function hasDynamicName(declaration: Declaration): declaration is DynamicNamedDeclaration | DynamicNamedBinaryExpression {
    const name = getNameOfDeclaration(declaration);
    return !!name && isDynamicName(name);
}
```

`AccessorDeclaration` is `GetAccessorDeclaration | SetAccessorDeclaration`.
Each interface has its own literal kind (`GetAccessor` or `SetAccessor`)
and requires `readonly name: PropertyName`. Neither kind matches any
`getNonAssignedNameOfDeclaration` switch case, so both reach
`return (declaration as NamedDeclaration).name;`. That property holds a
node, so `!!name` is true. In `getNameOfDeclaration` the defined accessor
passes the undefined check, and the returned name satisfies the first
operand of the `||`. Thus the assigned-name fallback is skipped and both
paths supply the same `accessor.name` to the same helper.

```ts
export function isDynamicName(name: DeclarationName): boolean {
    if (!(name.kind === SyntaxKind.ComputedPropertyName || name.kind === SyntaxKind.ElementAccessExpression)) {
        return false;
    }
    const expr = isElementAccessExpression(name) ? skipParentheses(name.argumentExpression) : name.expression;
    return !isStringOrNumericLiteralLike(expr) &&
        !isSignedNumericLiteral(expr);
}
```


```ts
export function getNameOfDeclaration(declaration: Declaration | Expression | undefined): DeclarationName | undefined {
    if (declaration === undefined) return undefined;
    return getNonAssignedNameOfDeclaration(declaration) ||
        (isFunctionExpression(declaration) || isArrowFunction(declaration) || isClassExpression(declaration) ? getAssignedName(declaration) : undefined);
}

export function getNonAssignedNameOfDeclaration(declaration: Declaration | Expression): DeclarationName | undefined {
    switch (declaration.kind) {
        case SyntaxKind.Identifier:
            return declaration as Identifier;
        case SyntaxKind.JSDocPropertyTag:
        case SyntaxKind.JSDocParameterTag: {
            const { name } = declaration as JSDocPropertyLikeTag;
            if (name.kind === SyntaxKind.QualifiedName) {
                return name.right;
            }
            break;
        }
        case SyntaxKind.CallExpression:
        case SyntaxKind.BinaryExpression: {
            const expr = declaration as BinaryExpression | CallExpression;
            switch (getAssignmentDeclarationKind(expr)) {
                case AssignmentDeclarationKind.ExportsProperty:
                case AssignmentDeclarationKind.ThisProperty:
                case AssignmentDeclarationKind.Property:
                case AssignmentDeclarationKind.PrototypeProperty:
                    return getElementOrPropertyAccessArgumentExpressionOrName((expr as BinaryExpression).left as AccessExpression);
                case AssignmentDeclarationKind.ObjectDefinePropertyValue:
                case AssignmentDeclarationKind.ObjectDefinePropertyExports:
                case AssignmentDeclarationKind.ObjectDefinePrototypeProperty:
                    return (expr as BindableObjectDefinePropertyCall).arguments[1];
                default:
                    return undefined;
            }
        }
        case SyntaxKind.JSDocTypedefTag:
            return getNameOfJSDocTypedef(declaration as JSDocTypedefTag);
        case SyntaxKind.JSDocEnumTag:
            return nameForNamelessJSDocTypedef(declaration as JSDocEnumTag);
        case SyntaxKind.ExportAssignment: {
            const { expression } = declaration as ExportAssignment;
            return isIdentifier(expression) ? expression : undefined;
        }
        case SyntaxKind.ElementAccessExpression:
            const expr = declaration as ElementAccessExpression;
            if (isBindableStaticElementAccessExpression(expr)) {
                return expr.argumentExpression;
            }
    }
    return (declaration as NamedDeclaration).name;
}
```

A plain name still returns false. A computed name with a string or numeric
literal-like expression or signed numeric expression still returns false.

## Statement exclusion

At the meter's `transformers/declarations/diagnostics.ts:733:17`, in
`findNearestDeclaration`, the final conditional returns the same `result`
for the three declaration kinds and otherwise returns `undefined`. The
existing cast encloses the replacement conditional:

```ts
isVariableDeclaration(result) || isPropertyDeclaration(result) || isParameter(result) ? result : undefined
```

The locked checker reads `VariableDeclaration | ParameterDeclaration |
PropertyDeclaration | undefined` for the cast operand on both sides. The
existing declaration predicates supply the type narrowing as well as the kind
test. The round 1 kind comparisons admitted `Statement` in the operand type and
have been replaced.

```ts
const result = findAncestor(node, n => isExportAssignment(n) || isStatement(n) || isVariableDeclaration(n) || isPropertyDeclaration(n) || isParameter(n));
if (!result) return undefined;

if (isExportAssignment(result)) return result;

if (isReturnStatement(result)) {
    return findAncestor(result, (n): n is Exclude<FunctionLikeDeclaration, ConstructorDeclaration> => isFunctionLikeDeclaration(n) && !isConstructorDeclaration(n));
}
```

The walk tests the current node, returns it when the callback succeeds, and
otherwise advances through `parent`. The export and return branches above
leave before the final conditional. Each remaining result therefore passed
`isStatement(n) || isVariableDeclaration(n) || isPropertyDeclaration(n) || isParameter(n)`.

```ts
export function isStatement(node: Node): node is Statement {
    const kind = node.kind;
    return isStatementKindButNotDeclarationKind(kind)
        || isDeclarationStatementKind(kind)
        || isBlockStatement(node);
}

function isBlockStatement(node: Node): node is Block {
    if (node.kind !== SyntaxKind.Block) return false;
    if (node.parent !== undefined) {
        if (node.parent.kind === SyntaxKind.TryStatement || node.parent.kind === SyntaxKind.CatchClause) {
            return false;
        }
    }
    return !isFunctionBlock(node);
}
```

The first two helpers compare only the supplied kind. `isFunctionBlock`
tests `node && node.kind === SyntaxKind.Block && isFunctionLike(node.parent)`,
and `isFunctionLike` tests parent existence and kind. Thus both calls of
`isStatement` inspect the same kind and parent chain. The walk and the
intervening export and return tests write neither. A statement accepted by the
walk still tests true and produces undefined. `VariableDeclaration`,
`PropertyDeclaration` and `Parameter` are absent from both statement-kind
lists and each fails the initial Block test. The three declaration predicates
compare exactly those kinds, so every other result the walk can yield on this
path produces undefined in both forms. The original declaration union cast
stays in place.

## The census at main 0e7517ac

The unit census by pass (J-xphepg9 entry, J-bvccrq2 directory) reads 390 to
389 stopping and 0 units newly stopping in each pass.
`getAllAccessorDeclarations` lowers, off `utilities.ts:6875:9`. Unit
`createGetIsolatedDeclarationErrors` leaves `findNearestDeclaration`'s
`transformers/declarations/diagnostics.ts:733:17` and first-stops at
`:725:48`, NotYet, "a checked predicate result returned from a function the
checker infers a type predicate for, an ArrowFunction, whose uses the walk
doesn't follow, which its direction walk doesn't read". That site is the
`findAncestor` callback, upstream's bytes, which the adapter doesn't touch.
The reason is on the meter's list (one site, Lower's #1761dah). The base never
listed the site, because the unit stopped first at `:733:17`, so it is an exposed
stop, not a new one. The edit at `:733` is what carries the walk to the
callback's inferred predicate: the conditional now narrows `result` through
`isVariableDeclaration`, `isPropertyDeclaration` and `isParameter`.

## Verification

The adapter selects two AST slots and checks the exact premise text with CRLF
normalized to LF. It checks compiler value bindings before planning either
edit, including on reapply. A same-file binding must be the reviewed
declaration. An external binding must be an unaliased named import from the
compiler namespace with a direct re-export to the reviewed declaration.
A second binding of a reviewed name is premise drift. Before reading the
route, the guard matches all 78 barrel statements in order and by exact text
against `barrel.json`. The list includes 74 export stars, two namespace
imports and their two named exports. The route check is independent of
stock TypeScript's TS2308 ambiguity diagnostic. A write to the barrel's
exports from a loaded module is outside this guard. The stock source graph
scan found zero literal relative require calls and zero CommonJS export writes.
Through the compiler namespace it found one write, `src/tsc/tsc.ts:6`
setting `ts.Debug.loggingHost`, and through named imports from it two more,
`src/compiler/sys.ts:1970` setting `Debug.isDebugging` and
`src/typescript/typescript.ts:9` setting `Debug.loggingHost`. Each sets a
field of the exported `Debug` object and replaces no own binding of either
namespace or of a starred module's exports.

A change to tsc's generated barrel refuses the apply until the list is read
again. `make-barrel.cjs`, beside this README, writes the list from the barrel
this adapter receives, `src/compiler/_namespaces/ts.ts` at its `-before` ref:
run `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-53-predicate-targets/make-barrel.cjs <before> <after> > stage3/adapt/99-seats-build22-53-predicate-targets/barrel.json` from the repository,
with the barrel from the `-before` and `-after` refs (they must be equal; the
same file twice serves when only the input is at hand). The generator holds
today's shape: 78 statements, 74 stars at lines 3 to 76, and the four tail
statements by exact text. A deliberate upstream change to that shape updates
the generator as well as the list, and either change is read for what each
new or moved module loads and exposes before it lands.

Run `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-53-predicate-targets/test.cjs <applied-tree>` from the repository.
The test reads the adapter's before snapshot, checks clean reapply bytes,
and checks nine controls on both input states. The controls cover the
ancestor filter, export return, accessor lookup switch, statement and
accessor import aliases, barrel value, added star with side effects, added
star with a checking directive, and reordered stars.
The locked TypeScript 6.0.3 and its namespace helpers exercise declaration and enclosing-function
identity, statement exclusion and eight accessor cases in a VM. Finally,
the test compares both edited files with the adapter's own after snapshot.
