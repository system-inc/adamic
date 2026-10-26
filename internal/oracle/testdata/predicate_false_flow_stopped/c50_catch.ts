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
function f(n: Node, test: Test): number {
    try {
        if (test(n)) {
            throw new Error('a statement');
        }
    } catch {
        return 0;
    }
    return n.e;
}
console.log(String(f(statement, isS)));
