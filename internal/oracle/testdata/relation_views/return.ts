// rulings.md:72 (#0b9xqhh): a return whose type lacks Target's optional label crosses as a checked view.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function admit(source: Source): Target {
    return source;
}
const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
console.log(admit(bare).label ?? "absent");
console.log(admit(hidden).label ?? "absent");
