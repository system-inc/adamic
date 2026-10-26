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
function both(n: Node): n is Statement;
function both(n: number): boolean;
function both(n: Node | number): boolean {
    calls = calls + 1;
    return typeof n === 'number' && calls > 1000;
}
function f(n: Node, test: typeof both): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log(String(f(statement, both)));
