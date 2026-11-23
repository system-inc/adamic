// binder.ts:1416's shape (#axqgdck): a CallExpression out of a union of nodes whose kind is open on both members. The
// target's literal kind is tested at the cast, the ledger's tagged interface downcast, and its reads stay checked.
enum SyntaxKind { Identifier = 80, CallExpression = 213, Block = 241 }
interface Expression { readonly kind: SyntaxKind; readonly end: number }
interface Statement { readonly kind: SyntaxKind; readonly pos: number }
interface CallExpression extends Expression { readonly kind: SyntaxKind.CallExpression; readonly argumentCount: number }
function argumentsOf(node: Expression | Statement): number {
    return (node as CallExpression).argumentCount;
}
const call: CallExpression = { kind: SyntaxKind.CallExpression, end: 9, argumentCount: 2 };
console.log(String(argumentsOf(call)));
