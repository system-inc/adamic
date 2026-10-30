// An inferred predicate passed to a parameter typed with it (#6fas8b6): total's test is a typeof
// isAOrB, which the contract pass doesn't read, since it reads only a predicate written in a function
// type. The call through it narrows, so the function stays NotYet, naming the use.
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
// The parameter's type is the inferred predicate's own, so a call through it narrows by its claim.
function total(array: readonly Node[], test: typeof isAOrB): number {
    let sum = 0;
    for (const v of array) {
        if (test(v)) {
            sum = sum + (v.kind === "A" ? v.a : v.b.length);
        }
    }
    return sum;
}
console.log(String(total(nodes, isAOrB)));
