import * as self from "./escape_computed_member.ts";
interface Expression {
    readonly kind: "E";
    readonly e: number;
}
interface Statement {
    readonly kind: "S";
    readonly s: string;
}
type Node = Expression | Statement;
let calls = 0;
// Unproven, and it lies on false: it says no to a Statement.
export function isS(n: Node): n is Statement {
    calls = calls + 1;
    return calls > 1000;
}
const statement: Node = { kind: "S", s: "x" };
function f(n: Node, test: (n: Node) => n is Statement): number {
    if (test(n)) {
        return 0;
    }
    return n.e;
}
console.log(String(f(statement, self["isS"])));
