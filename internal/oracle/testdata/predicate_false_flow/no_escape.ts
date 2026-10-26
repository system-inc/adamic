// No checked predicate escapes here: isS is proven, so the false-direction walk is off and every shape
// below, each narrowing n on the false side of a call through a predicate value, lowers as on main.
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
// Proven: a tag comparison on the parameter.
function isS(n: Node): n is Statement {
    return n.kind === 'S';
}
function tick(): number {
    return 1;
}
const expression: Node = { kind: 'E', e: 7 };
function c01(n: Node, test: Test): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c01 ' + String(c01(expression, isS)));

function c02(n: Node, test: Test): number {
    if (test(n)) {
        return 0;
    } else {
        return n.e;
    }
}
console.log('c02 ' + String(c02(expression, isS)));

function c03(n: Node, test: Test): number {
    return test(n) ? 0 : n.e;
}
console.log('c03 ' + String(c03(expression, isS)));

function c04(n: Node, test: Test): boolean | number {
    return test(n) || n.e;
}
console.log('c04 ' + String(c04(expression, isS)));

function c05(n: Node, test: Test): boolean | number {
    return !test(n) && n.e;
}
console.log('c05 ' + String(c05(expression, isS)));

function c06(n: Node, test: Test): number {
    if (test(n)) {
        return 0;
    }
    return n.e;
}
console.log('c06 ' + String(c06(expression, isS)));

function c07(n: Node, test: Test): number {
    if (test(n)) {
        throw new Error('a statement');
    }
    return n.e;
}
console.log('c07 ' + String(c07(expression, isS)));

function c08(ns: Node[], test: Test): number {
    let total = 0;
    for (const n of ns) {
        if (test(n)) {
            continue;
        }
        total = total + n.e;
    }
    return total;
}
console.log('c08 ' + String(c08([expression], isS)));

function c09(n: Node, test: Test): number {
    const t = test;
    if (!t(n)) {
        return n.e;
    }
    return 0;
}
console.log('c09 ' + String(c09(expression, isS)));

function c10(n: Node, test: Test): number {
    const ok = test(n);
    if (!ok) {
        return n.e;
    }
    return 0;
}
console.log('c10 ' + String(c10(expression, isS)));

function c11(n: Node, test: Test): number {
    switch (true) {
        case test(n):
            return 0;
        default:
            return n.e;
    }
}
console.log('c11 ' + String(c11(expression, isS)));

function c12(n: Node, test: Test): number {
    if (test(n) === false) {
        return n.e;
    }
    return 0;
}
console.log('c12 ' + String(c12(expression, isS)));

function c14(n: Node, test: Test): number {
    if (test(n) !== true) {
        return n.e;
    }
    return 0;
}
console.log('c14 ' + String(c14(expression, isS)));

function c15(n: Node, test: Test): number {
    if (false === test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c15 ' + String(c15(expression, isS)));

function c16(n: Node, test: Test): number {
    if (!!test(n)) {
        return 0;
    } else {
        return n.e;
    }
}
console.log('c16 ' + String(c16(expression, isS)));

function c17(n: Node, test: Test): number {
    if ((tick(), test(n))) {
        return 0;
    }
    return n.e;
}
console.log('c17 ' + String(c17(expression, isS)));

function c18(n: Node, test: Test): number {
    while (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c18 ' + String(c18(expression, isS)));

function c19(n: Node, opts: { test: Test }): number {
    if (!opts.test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c19 ' + String(c19(expression, { test: isS })));

function c20(n: Node, { test }: { test: Test }): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c20 ' + String(c20(expression, { test: isS })));

class C21 {
    constructor(readonly test: Test) {}
    run(n: Node): number {
        if (!this.test(n)) {
            return n.e;
        }
        return 0;
    }
}
console.log('c21 ' + String(new C21(isS).run(expression)));

function c22(n: Node, test: Test): () => number {
    return () => {
        if (!test(n)) {
            return n.e;
        }
        return 0;
    };
}
console.log('c22 ' + String(c22(expression, isS)()));

function c25(n: Node, tests: Test[]): number {
    if (!tests[0]!(n)) {
        return n.e;
    }
    return 0;
}
console.log('c25 ' + String(c25(expression, [isS])));

function c26(n: Node, ...tests: Test[]): number {
    const t = tests[0];
    if (t && !t(n)) {
        return n.e;
    }
    return 0;
}
console.log('c26 ' + String(c26(expression, isS)));

function c27(n: Node, get: () => Test): number {
    if (!get()(n)) {
        return n.e;
    }
    return 0;
}
console.log('c27 ' + String(c27(expression, () => isS)));

function c28<P extends Test>(n: Node, test: P): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c28 ' + String(c28(expression, isS)));

function c31(box: { n: Node }, test: Test): number {
    if (!test(box.n)) {
        return box.n.e;
    }
    return 0;
}
console.log('c31 ' + String(c31({ n: expression }, isS)));

function c32(box: { n: Node }, test: Test): number {
    if (!test(box.n)) {
        return box['n'].e;
    }
    return 0;
}
console.log('c32 ' + String(c32({ n: expression }, isS)));

function c35(n: Node, test: Test): number {
    if (!test(n)!) {
        return n.e;
    }
    return 0;
}
console.log('c35 ' + String(c35(expression, isS)));

function c36(n: Node, test: Test = isS): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c36 ' + String(c36(expression)));

function c37(n: Node, test: Test): number {
    const t: typeof test = test;
    if (!t(n)) {
        return n.e;
    }
    return 0;
}
console.log('c37 ' + String(c37(expression, isS)));

function c38(n: Node, test: Test): number {
    let seen = 0;
    do {
        seen = seen + 1;
    } while (!test(n) && seen < n.e);
    return seen;
}
console.log('c38 ' + String(c38(expression, isS)));

function c40<T extends Node>(n: T, test: Test): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c40 ' + String(c40(expression, isS)));

function isStatement(n: Node): n is Statement {
    return n.kind === 'S';
}
function c41(n: Node, test: Test | typeof isStatement): number {
    if (!test(n)) {
        return n.e;
    }
    return 0;
}
console.log('c41 ' + String(c41(expression, isS)));

function c42(n: Node, test: Test): number {
    const ok = !test(n) && true;
    if (ok) {
        return n.e;
    }
    return 0;
}
console.log('c42 ' + String(c42(expression, isS)));

function c43(n: Node, test: Test): number {
    let t = test;
    if (!t(n)) {
        return n.e;
    }
    return 0;
}
console.log('c43 ' + String(c43(expression, isS)));

function c48(n: Node, pair: [Test]): number {
    if (!pair[0](n)) {
        return n.e;
    }
    return 0;
}
console.log('c48 ' + String(c48(expression, [isS])));

function c49(n: Node, test: Test): number {
    if (!((t: Test) => t)(test)(n)) {
        return n.e;
    }
    return 0;
}
console.log('c49 ' + String(c49(expression, isS)));

function c50(n: Node, test: Test): number {
    try {
        if (test(n)) {
            throw new Error('a statement');
        }
    } catch {
        return 0;
    }
    return n.e;
}
console.log('c50 ' + String(c50(expression, isS)));

function c51(n: Node, test: Test): () => number {
    if (!test(n)) {
        return () => n.e;
    }
    return () => 0;
}
console.log('c51 ' + String(c51(expression, isS)()));

function c52(n: Node, test: Test): () => number {
    const m = n;
    if (!test(m)) {
        return () => m.e;
    }
    return () => 0;
}
console.log('c52 ' + String(c52(expression, isS)()));

function c54(b: { c: { n: Node } }, test: Test): number {
    if (!test(b.c.n)) {
        return b.c.n.e;
    }
    return 0;
}
console.log('c54 ' + String(c54({ c: { n: expression } }, isS)));

function c56(n: Node, test: Test): number {
    const ok = test(n);
    return ok ? 0 : n.e;
}
console.log('c56 ' + String(c56(expression, isS)));

function c57(n: Node, test: Test): number {
    const bad = !test(n);
    if (bad) {
        return n.e;
    }
    return 0;
}
console.log('c57 ' + String(c57(expression, isS)));
