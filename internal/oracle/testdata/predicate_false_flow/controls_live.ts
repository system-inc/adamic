// The switch is live: isStatement is a checked predicate (calls > 0 is no proof) and it escapes as a
// value, held below and never called. Nothing here narrows a read on the false side of a call
// through a predicate value, so the walk stops nothing: Language's controls the checker narrows
// nothing by, true-direction calls through a value, and R-T80's direct calls, checked at the call.
interface Expression {
    readonly kind: 'E';
    readonly e: number;
}
interface Statement {
    readonly kind: 'S';
    readonly s: string;
}
type Node = Expression | Statement;
type Test = typeof isP;
let calls = 0;
// Checked, and honest.
function isStatement(n: Node): n is Statement {
    calls = calls + 1;
    return n.kind === 'S' && calls > 0;
}
// Proven: a tag comparison on the parameter.
function isP(n: Node): n is Statement {
    return n.kind === 'S';
}
const held = isStatement;
const expression: Node = { kind: 'E', e: 7 };
const statement: Node = { kind: 'S', s: 'x' };
function c34(n: Node, test: Test): string {
    if (!(test(n) satisfies boolean)) {
        return n.kind;
    }
    return '';
}
function c44(n: Node, test: Test): string {
    switch (test(n)) {
        case false:
            return n.kind;
    }
    return '';
}
function c46(n: Node, test: Test): string {
    let ok = true;
    ok &&= test(n);
    if (!ok) {
        return n.kind;
    }
    return '';
}
function c53(n: Node, test: Test): string {
    if (test(n)) {
        return '';
    }
    function g(): string {
        return n.kind;
    }
    return g();
}
function trueSide(n: Node, opts: { test: Test }): string {
    if (opts.test(n)) {
        return n.s;
    }
    return 'not a statement';
}
function directNot(n: Node): number {
    if (!isStatement(n)) {
        return n.e;
    }
    return 0;
}
function directSwitch(n: Node): number {
    switch (true) {
        case isStatement(n):
            return 0;
        default:
            return n.e;
    }
}
function directEqualsFalse(n: Node): number {
    if (isStatement(n) === false) {
        return n.e;
    }
    return 0;
}
console.log(c34(expression, isP) + ' ' + c44(expression, isP) + ' ' + c46(expression, isP) + ' ' + c53(expression, isP));
console.log(trueSide(statement, { test: isP }) + ' ' + trueSide(expression, { test: isP }));
console.log(String(directNot(expression)) + ' ' + String(directSwitch(expression)) + ' ' + String(directEqualsFalse(expression)));
console.log(typeof held);
