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
class C21 {
    constructor(readonly test: Test) {}
    run(n: Node): number {
        if (!this.test(n)) {
            return n.e;
        }
        return 0;
    }
}
console.log(String(new C21(isS).run(statement)));
