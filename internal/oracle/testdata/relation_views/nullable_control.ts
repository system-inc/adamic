// rulings.md:72's second slice (#0b9xqhh), nullable on both sides, the control: a Source | undefined, held as
// one nullable pointer, goes into Target | null | undefined, which a union boxes. The two sides have two
// representations, so rule 3300 refuses it as before. The ruling's number | undefined against
// object | undefined is no relation tsc accepts, so it would stop as a type error and never reach the rule;
// this pair is the one that does.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function pick(present: boolean, source: Source): Source | undefined {
    return present ? source : undefined;
}
function show(target: Target | null | undefined): string {
    return target === undefined || target === null ? "none" : target.label ?? "absent";
}
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
console.log(show(pick(true, hidden)));
