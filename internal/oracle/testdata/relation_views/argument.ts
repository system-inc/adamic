// rulings.md:72 (#0b9xqhh): an argument whose type lacks Target's optional label crosses as a checked view.
// The absent label reads undefined and a present one reads its string, as on Node. A raw read would print
// the same here; the view's checks are what the lying twins (road2, road5, nested_union) stop on.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): string {
    return target.label === undefined ? "absent" : target.label;
}
const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
console.log(describe(bare));
console.log(describe(hidden));
