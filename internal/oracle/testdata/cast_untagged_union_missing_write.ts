interface A { note: string; extra: string }
interface B { readonly name: string; note: string }
// A write through an untagged union's view to a field the value lacks (#axqgdck, Language's y13): it stops at the
// write. Node writes the new property and prints the note.
function make(which: number): A | B {
    const a: A = { note: 'na', extra: 'ea' };
    const b: B = { name: 'nb', note: 'nb' };
    return which === 0 ? a : b;
}
const u = make(1);
(u as A).extra = 'written';
console.log((u as B).note);
