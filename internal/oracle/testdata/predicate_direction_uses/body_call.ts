// A checked call in the body of a function the checker infers a predicate for, outside its returned
// expression (#6fas8b6): the checker infers n is A from the kind test alone, so isLeaf's call is
// classified as any call is, and the program agrees with Node.
interface A {
    readonly kind: "A";
    readonly a: number;
}
interface B {
    readonly kind: "B";
    readonly b: string;
}
type Node = A | B;
const leafKinds: string[] = ["B"];
let leaves = 0;
// Unproven: the prover can't read a list, so its call is checked.
function isLeaf(n: Node): n is B {
    return leafKinds.includes(n.kind);
}
function isA(n: Node) {
    if (isLeaf(n)) {
        leaves = leaves + 1;
    }
    return n.kind === "A";
}
const nodes: Node[] = [{ kind: "A", a: 1 }, { kind: "B", b: "x" }, { kind: "A", a: 2 }];
for (const n of nodes) {
    if (isA(n)) {
        console.log(String(n.a));
    } else {
        console.log(n.b);
    }
}
console.log(String(leaves));
