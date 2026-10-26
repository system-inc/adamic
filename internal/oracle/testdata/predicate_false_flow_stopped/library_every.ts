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
const statements: Statement[] = [{ kind: 'S', s: 'x' }];
// every is the library's own member, but its `this is S[]` answer is isS's: on its false side the
// checker reads nodes as never.
function f(nodes: Statement[]): string {
    if (!nodes.every(isS)) {
        return typeof nodes;
    }
    return '';
}
console.log(f(statements));
