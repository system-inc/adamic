// The switch is live: isStatement is a checked predicate (calls > 0 is no proof) and it escapes as a
// value, held below and never called. Array.isArray is asked by its own name, so its false side is
// the engine's answer, true both ways, and each read of x narrowed by it lowers (Lower's ruling).
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
// Checked, and honest.
function isStatement(n: Node): n is Statement {
    calls = calls + 1;
    return n.kind === 'S' && calls > 0;
}
const held = isStatement;
function not(x: number[] | string): string {
    if (!Array.isArray(x)) {
        return x.toUpperCase();
    }
    return String(x.length);
}
function otherwise(x: number[] | string): string {
    if (Array.isArray(x)) {
        return String(x.length);
    } else {
        return x.toUpperCase();
    }
}
function equalsFalse(x: number[] | string): string {
    if (Array.isArray(x) === false) {
        return x.toUpperCase();
    }
    return String(x.length);
}
function early(x: number[] | string): string {
    if (Array.isArray(x)) {
        return String(x.length);
    }
    return x.toUpperCase();
}
console.log(not('ab') + ' ' + not([1, 2]) + ' ' + otherwise('cd') + ' ' + otherwise([3]));
console.log(equalsFalse('ef') + ' ' + equalsFalse([4, 5, 6]) + ' ' + early('gh') + ' ' + early([]));
console.log(typeof held);
