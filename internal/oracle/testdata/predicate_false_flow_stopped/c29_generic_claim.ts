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
// Unproven, and it lies on false: it says no to an Expression.
function isE(n: unknown): n is Expression {
    calls = calls + 1;
    return calls > 1000;
}
const expression: Expression = { kind: 'E', e: 1 };
function f<T extends Expression>(n: T | Statement, test: (n: unknown) => n is T): string {
    if (!test(n)) {
        return n.s;
    }
    return '';
}
console.log(String(f(expression, isE)));
