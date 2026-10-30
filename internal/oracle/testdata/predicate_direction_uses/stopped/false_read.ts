// An inferred predicate whose false side a caller reads (#6fas8b6): if (!isAOrB(n)) reads n.c, a C
// only if isA and isB both told the truth on false, which nothing checks, so it stays NotYet, naming
// the call.
interface A {
    readonly kind: "A";
    readonly a: number;
}
interface B {
    readonly kind: "B";
    readonly b: string;
}
interface C {
    readonly kind: "C";
    readonly c: boolean;
}
type Node = A | B | C;
const aKinds: string[] = ["A"];
const bKinds: string[] = ["B"];
// Unproven: the prover can't read a list, so each call is checked.
function isA(n: Node): n is A {
    return aKinds.includes(n.kind);
}
function isB(n: Node): n is B {
    return bKinds.includes(n.kind);
}
function some<T>(array: readonly T[], predicate: (value: T) => boolean): boolean {
    for (const v of array) {
        if (predicate(v)) {
            return true;
        }
    }
    return false;
}
const nodes: Node[] = [{ kind: "A", a: 1 }, { kind: "B", b: "x" }, { kind: "C", c: true }];
function isAOrB(n: Node) {
    return isA(n) || isB(n);
}
function flag(n: Node): boolean {
    if (!isAOrB(n)) {
        return n.c;
    }
    return false;
}
const c: Node = { kind: "C", c: true };
console.log(String(flag(c)) + " " + String(nodes.length));
