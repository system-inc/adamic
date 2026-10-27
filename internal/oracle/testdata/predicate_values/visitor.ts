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
// The field, element and return types below are inferred: annotated, a predicate-typed one is
// refused outright.
function pick() {
    return isS;
}
// One call, read only on its true side: a loop around a false side would carry that side's
// narrowing to the next iteration's read, which the walk stops.
function length(node: Node, test: (node: Node) => node is Statement): number {
    if (test(node)) {
        return node.s.length;
    }
    return 0;
}
const nodes: Node[] = [{ kind: "E", e: 1 }, { kind: "S", s: "first" }, { kind: "S", s: "second" }];
const identity = (node: Node): Node => node;
for (const node of nodes) {
    const visited = visitNode(node, identity, isS);
    console.log(visited === undefined ? "none" : visited.s);
}
console.log(visitNodes(nodes, isS).map(statement => statement.s).join(","));
const visitor = { test: isS };
const tests = [isS, pick(), visitor.test];
let count = 0;
for (const node of nodes) {
    for (const test of tests) {
        count = count + length(node, test);
    }
}
console.log(String(count));
const alias = isS;
console.log(String(alias === isS));
console.log(String(pick() === visitor.test));
