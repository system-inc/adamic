// Compiler's probe on #0b9xqhh: the relation sits in a generic body, lowered once for each instantiation, and
// both instantiations' values are read through Target, so each lowering's view is the one its read meets. The
// honest first prints, and the lying second stops at its checked read (exit 70) where Node prints 5.
// generic_two_instantiations_honest.ts is the twin that agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function admit<Tag>(tag: Tag, source: Source): Target {
    const target: Target = source;
    return target;
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
const fromNumber = admit(1, first);
const fromString = admit("two", second);
console.log(fromNumber.label ?? "absent");
console.log(fromString.label ?? "absent");
