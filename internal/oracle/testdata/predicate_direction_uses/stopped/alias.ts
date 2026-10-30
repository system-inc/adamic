// A checked call that initializes a const its function's returned expression reads (#6fas8b6): the
// checker infers n is A through the alias, so the call stays NotYet though it's outside the returned
// expression.
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
function isAlsoA(n: Node) {
    const ok = isA(n);
    return ok;
}
console.log(String(some(nodes, isAlsoA)));
