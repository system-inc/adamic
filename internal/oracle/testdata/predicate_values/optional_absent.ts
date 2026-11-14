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
// Unproven, and it lies on false: it says no to a Statement. It's never called here.
function isS(n: Node): n is Statement {
    calls = calls + 1;
    return calls > 1000;
}
const statement: Node = { kind: 'S', s: 'x' };
// test is absent, so test?.(n) is undefined, the if goes false, and n is read at its declared Node.
function f(n: Node, test?: Test): string {
    if (test?.(n)) {
        return '';
    }
    return n.kind;
}
const held: Test = isS;
console.log(String(f(statement)));
console.log(String(f(statement, undefined)));
console.log(String(held === isS));
