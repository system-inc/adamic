// tsc's optional scalars: ObjectLiteralExpression's multiLine?: boolean, a node's optional flags, and
// ?. of a required number through a receiver that may be absent.
interface Node { readonly kind: number; readonly pos: number; }
interface ObjectLiteralExpression extends Node { readonly kind: 210; readonly multiLine?: boolean; readonly flags?: number; }

function literal(node: Node): ObjectLiteralExpression | undefined {
    return node.kind === 210 ? node as ObjectLiteralExpression : undefined;
}
function describe(node: Node): string {
    const object = literal(node);
    const multiLine = object?.multiLine;
    const flags = object?.flags;
    return String(object?.pos) + ":" + (multiLine === undefined ? "unset" : String(multiLine)) + ":" + (flags === undefined ? "none" : String(flags + 1));
}
function lines(node: ObjectLiteralExpression): string {
    return node.multiLine ? "multi" : "single";
}
const multi: ObjectLiteralExpression = { kind: 210, pos: 4, multiLine: true, flags: 1 };
const bare: ObjectLiteralExpression = { kind: 210, pos: 9 };
console.log(describe(multi));
console.log(describe(bare));
console.log(describe({ kind: 1, pos: 2 }));
console.log(lines(literal(multi) ?? bare));
console.log(lines(literal(bare) ?? multi));
