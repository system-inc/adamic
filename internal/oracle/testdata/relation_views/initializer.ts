// rulings.md:72 (#0b9xqhh): an initializer whose type lacks Target's optional label crosses as a checked view.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const absent: Target = bare;
const present: Target = hidden;
console.log(absent.label ?? "absent");
console.log(present.label ?? "absent");
