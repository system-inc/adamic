interface Node { readonly kind: number; }
interface Ident extends Node { readonly kind: 1; readonly a: number; readonly b: number; readonly c: number; readonly d: number; readonly e: number; readonly f: number; readonly g: number; readonly h: number; }
interface Other extends Node { readonly kind: 2; }
const size = 1024;
const count = 40000;

const nodes: Node[] = [];
for (let i = 0; i < size; i++) {
    const made: Ident = { kind: 1, a: i, b: 1, c: 2, d: 3, e: 4, f: 5, g: 6, h: 7 };
    nodes.push(made);
}
let total = 0;
for (let round = 0; round < count; round++) {
    for (let i = 0; i < size; i++) {
        const n = nodes[i];
        if (n === undefined) { continue; }
        const v = n as Ident;
        total += v.a + v.b + v.c + v.d + v.e + v.f + v.g + v.h;
    }
}
console.log(String(total));
