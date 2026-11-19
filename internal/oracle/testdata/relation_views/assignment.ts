// rulings.md:72 (#0b9xqhh): an assignment whose type lacks Target's optional label crosses as a checked view.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
let target: Target = { kind: 0, label: "zero" };
console.log(target.label ?? "absent");
target = bare;
console.log(target.label ?? "absent");
target = hidden;
console.log(target.label ?? "absent");
