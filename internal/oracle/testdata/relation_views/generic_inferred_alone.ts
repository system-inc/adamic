// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: keep's T is inferred from the argument
// alone, so T is Source and the slot at the call is Source | undefined, the argument's own type. Nothing is
// widened: rule 3300 finds no hidden field, no view is made, and the program agrees with Node.
interface Source { readonly kind: number; }

function keep<T>(value: T | undefined): T | undefined {
    return value;
}
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const kept = keep(hidden);
console.log(kept === undefined ? "none" : String(kept.kind));
