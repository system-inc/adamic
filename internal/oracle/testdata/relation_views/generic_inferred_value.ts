// rulings.md:72's second slice (#0b9xqhh), a generic call: wrap's T is inferred from the hidden value alone,
// no array fixing it (an argument's inference outranks the contextual return type's), so T is Source and
// wrap(second) is a Source[]. Going into a readonly Target[] it's a container relation, viewed by its element
// pair: the honest call prints, and the lie's label, a number, stops at its checked read (exit 70) where Node
// prints 5.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function wrap<T>(value: T): T[] {
    return [value];
}
function describe(target: Target): string {
    return target.label ?? "absent";
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
const shown: readonly Target[] = wrap(first);
const lied: readonly Target[] = wrap(second);
for (const target of shown) {
    console.log(describe(target));
}
for (const target of lied) {
    console.log(describe(target));
}
