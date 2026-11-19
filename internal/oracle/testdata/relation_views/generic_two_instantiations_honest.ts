// generic_two_instantiations.ts's honest twin: both instantiations' values hold a string label, and each
// lowering's view reads it as Node does.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function admit<Tag>(tag: Tag, source: Source): Target {
    const target: Target = source;
    return target;
}
const plain = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const first: Source = plain;
const second: Source = labelled;
const fromNumber = admit(1, first);
const fromString = admit("two", second);
console.log(fromNumber.label ?? "absent");
console.log(fromString.label ?? "absent");
