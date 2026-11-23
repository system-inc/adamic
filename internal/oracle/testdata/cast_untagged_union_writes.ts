// Writes through an untagged union's view (#axqgdck), from Language's y05 and y18 on #03e5ke4, held to Node. A write
// in a callee taking a plain A lands in the B's own note slot, as Node writes it. A link stored through the view into
// a B makes a cycle the program never reads back; the counted build frees it whole, held in graph regions.
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

interface P { owner: Holder | undefined }
interface Holder { readonly id: number; link: P }
interface Named { readonly name: string; link: P }
function made(which: number): Holder | Named {
    const holder: Holder = { id: 1, link: { owner: undefined } };
    const named: Named = { name: 'b', link: { owner: undefined } };
    return which === 0 ? holder : named;
}
function tie(holder: Holder): void {
    holder.link = { owner: holder };
}
tie(made(0) as Holder);
const tied = made(1);
tie(tied as Holder);
console.log((tied as Named).name);
