// rulings.md:72's second slice (#0b9xqhh), nullable on both sides: holder && holder.source, of type
// undefined | Source, returned as Target | undefined, the shape of moduleNameResolver.ts:140 with a source that
// can lie. The object arms are viewed, so the second Holder's hidden label, a number, stops at its read through
// Target (exit 70) where Node prints 5. nullable_both.ts holds the honest twin.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Holder { readonly source: Source; }

function inner(holder: Holder | undefined): Target | undefined {
    return holder && holder.source;
}
function show(target: Target | undefined): string {
    return target === undefined ? "none" : target.label ?? "absent";
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Holder = { source: honest };
const second: Holder = { source: lying };
console.log(show(inner(undefined)));
console.log(show(inner(first)));
console.log(show(inner(second)));
