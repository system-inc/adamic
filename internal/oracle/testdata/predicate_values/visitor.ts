// A checked .ts predicate used as a value (#vdt9f1q): isS is unproven but honest, so every call
// through its value passes its check, and the program agrees with Node, in tsc's visitor shape.
interface Expression {
    readonly kind: "E";
    readonly e: number;
}
interface Statement {
    readonly kind: "S";
    readonly s: string;
}
type Node = Expression | Statement;
const statementKinds: string[] = ["S"];
// Unproven: the prover can't read a list, so every call through its value is checked.
function isS(n: Node): n is Statement {
    return statementKinds.includes(n.kind);
}
function visitNode<T extends Node>(node: Node, visitor: (node: Node) => Node, test: (node: Node) => node is T): T | undefined {
    const visited = visitor(node);
    if (test(visited)) {
        return visited;
    }
    return undefined;
}
function visitNodes<T extends Node>(nodes: Node[], test: (node: Node) => node is T): T[] {
    const kept: T[] = [];
    for (const node of nodes) {
        if (test(node)) {
            kept.push(node);
        }
    }
    return kept;
}
interface Visitor {
    readonly test: (node: Node) => node is Statement;
}
function pick(): (node: Node) => node is Statement {
    return isS;
}
const nodes: Node[] = [{ kind: "E", e: 1 }, { kind: "S", s: "first" }, { kind: "S", s: "second" }];
const identity = (node: Node): Node => node;
for (const node of nodes) {
    const visited = visitNode(node, identity, isS);
    console.log(visited === undefined ? "none" : visited.s);
}
console.log(visitNodes(nodes, isS).map(statement => statement.s).join(","));
const visitor: Visitor = { test: isS };
const tests: ((node: Node) => node is Statement)[] = [isS, pick()];
let count = 0;
for (const node of nodes) {
    if (visitor.test(node) && tests[0](node) && tests[1](node)) {
        count = count + node.s.length;
    }
}
console.log(String(count));
const alias = isS;
console.log(String(alias === isS), String(pick() === visitor.test));
