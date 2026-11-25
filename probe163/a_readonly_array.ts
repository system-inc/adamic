interface Literal { kind: "literal"; text: string }
interface Dynamic { kind: "dynamic"; name: string }
type Expr = Literal | Dynamic;
interface Call { kind: "call"; readonly arguments: readonly Expr[] }
function isStringLiteralLike(node: Expr): node is Literal {
    return node.kind === "literal";
}
function isRequireCall(node: Call, requireLiteral: boolean): boolean {
    return node.arguments.length === 1;
}
interface Other { kind: "other" }
function walk(node: Call, cb: (node: Call | Other, argument: Literal) => void): void {
    if (isRequireCall(node, false) && isStringLiteralLike(node.arguments[0]!)) {
        cb(node, node.arguments[0]);
    }
}
const literal: Literal = { kind: "literal", text: "literal" };
const dynamic: Dynamic = { kind: "dynamic", name: "name" };
const first: readonly Expr[] = [literal];
const second: readonly Expr[] = [dynamic];
walk({ kind: "call", arguments: first }, (_call, arg) => { console.log(arg.text); });
walk({ kind: "call", arguments: second }, (_call, arg) => { console.log(arg.text); });
console.log("done");
