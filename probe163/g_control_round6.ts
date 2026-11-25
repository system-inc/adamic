interface NodeArray<T> extends ReadonlyArray<T> { readonly pos: number; readonly end: number; readonly hasTrailingComma: boolean; transformFlags: number }
interface Literal { kind: "literal"; text: string }
interface Dynamic { kind: "dynamic"; name: string }
type Expr = Literal | Dynamic;
interface Call { kind: "call"; readonly arguments: NodeArray<Expr> }
function isStringLiteralLike(node: Expr): node is Literal {
    return node.kind === "literal";
}
type Require = Call & { arguments: [Expr & Literal] };
function isRequireCall(node: Call, requireLiteral: true): node is Require;
function isRequireCall(node: Call, requireLiteral: boolean): boolean {
    return node.arguments.length === 1;
}
interface Other { kind: "other" }
function walk(node: Call, cb: (node: Call | Other, argument: Literal) => void): void {
    if (isRequireCall(node, true)) {
        cb(node, node.arguments[0]);
    }
}
const literal: Literal = { kind: "literal", text: "literal" };
const dynamic: Dynamic = { kind: "dynamic", name: "name" };
function run(node: Call): void { walk(node, (_call, arg) => { console.log(arg.text); }); }
console.log(literal.text, dynamic.name, typeof run);
console.log("done");
