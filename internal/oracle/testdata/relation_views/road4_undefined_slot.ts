// Language's road 4 on #0b9xqhh, the widening: an optional parameter's slot is Target | undefined, and a source
// that holds neither undefined nor null is viewed as Target there. A present value's absent label, a present
// one, and no argument at all agree with Node.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function show(target?: Target): string {
    return target === undefined ? "none" : target.label ?? "absent";
}
const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
console.log(show(bare));
console.log(show(hidden));
console.log(show());
