interface Expression {
    readonly kind: 'E';
    readonly e: number;
}
interface Statement {
    readonly kind: 'S';
    readonly s: string;
}
type Node = Expression | Statement;
let calls = 0;
// Unproven, and it lies on false: it says no to a Statement.
function isS(n: Node): n is Statement {
    calls = calls + 1;
    return calls > 1000;
}
function tick(): number {
    return 1;
}
// It escapes as a value, so the false-direction walk is live, and the direct call below stays R-T80's.
const held = isS;
const statement: Node = { kind: 'S', s: 'x' };
function f(n: Node): number {
    switch (true) {
        case isS(n):
            return 0;
        default:
            return n.e;
    }
}
console.log(typeof held + ' ' + String(f(statement)));
