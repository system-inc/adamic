// An inferred predicate passed to filter (#6fas8b6): filter's parameter is typed with a type
// predicate, value is S, so the call through it narrows. The contract pass refuses it first, naming
// the argument, as it refuses any predicate argument whose body it can't prove.
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
console.log(String(nodes.filter(isAOrB).length));
