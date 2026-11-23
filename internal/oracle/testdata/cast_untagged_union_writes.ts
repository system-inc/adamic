// A write through an untagged union's view (#axqgdck), from Language's y05 on #03e5ke4, held to Node: a write in a
// callee taking a plain A lands in the B's own note slot, as Node writes it. It has no cycle: in a graph program the
// view of A | B as A stops NotYet, since B's child (Y) and A's (X) don't relate (Memory's item 4, J-1ycny47).
interface X { readonly a: number; readonly b: string }
interface Y { readonly b: number; readonly a: string }
interface A { note: string; readonly child: X; readonly items: readonly X[]; readonly f: number }
interface B { readonly f: number; readonly g: string; readonly items: readonly Y[]; readonly child: Y; note: string }
function make(which: number): A | B {
    const a: A = { note: 'na', child: { a: 1, b: 'x' }, items: [{ a: 2, b: 'y' }], f: 3 };
    const b: B = { f: 6, g: 'g', items: [{ b: 5, a: 'w' }], child: { b: 4, a: 'z' }, note: 'nb' };
    return which === 0 ? a : b;
}
function mark(a: A, text: string): void {
    a.note = text;
}
const u = make(1);
mark(u as A, 'written');
mark(make(0) as A, 'honest');
const back = u as B;
console.log(back.note + ' ' + back.g + ' ' + String(back.f));
