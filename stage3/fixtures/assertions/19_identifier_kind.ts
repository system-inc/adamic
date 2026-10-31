// From TypeScript 6.0.3, src/compiler/checker.ts:44336
// Census reason: type assertion sites
// Support types and driver are reduced scaffolding; the upstream function is unchanged.

const SyntaxKind = { Identifier: 80, PropertyAccessExpression: 212 } as const;
interface Node { readonly kind: number; }
type Expression = Node;
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface PrivateIdentifier extends Node { readonly kind: 81; readonly escapedText: string; }
interface PropertyAccessExpression extends Node { readonly kind: 212; readonly name: Identifier; }

function getIdentifierFromEntityNameExpression(node: Expression): Identifier | PrivateIdentifier | undefined {
        switch (node.kind) {
            case SyntaxKind.Identifier:
                return node as Identifier;
            case SyntaxKind.PropertyAccessExpression:
                return (node as PropertyAccessExpression).name;
            default:
                return undefined;
        }
    }

const id: Identifier = { kind: 80, escapedText: "name" };
const access: PropertyAccessExpression = { kind: 212, name: id };
console.log(getIdentifierFromEntityNameExpression(id)?.escapedText ?? "missing");
console.log(getIdentifierFromEntityNameExpression(access)?.escapedText ?? "missing");
console.log(getIdentifierFromEntityNameExpression({ kind: 110 })?.escapedText ?? "missing");
