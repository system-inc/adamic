interface Expression {
    readonly kind: 'E';
    readonly e: number;
}
interface Statement {
    readonly kind: 'S';
    readonly s: string;
}
type Node = Expression | Statement;
type Test = typeof isS;
let calls = 0;
// Unproven, and it lies on false: it says no to a Statement.
function isS(n: Node): n is Statement {
    calls = calls + 1;
    return calls > 1000;
}
const statement: Node = { kind: 'S', s: 'x' };
// Unproven, and it lies on false, as a function expression: a value by construction.
const isT = function (n: Node): n is Statement {
    calls = calls + 1;
    return calls > 1000;
};
function f(n: Node): number {
    if (!isT(n)) {
        return n.e;
    }
    return 0;
}
console.log(String(f(statement)));
