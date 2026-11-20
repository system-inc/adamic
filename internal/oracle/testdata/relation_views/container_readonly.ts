// rulings.md:72's second slice (#0b9xqhh), containers: a Source[] goes into a readonly Target[] parameter, so
// the element pair is viewed, registered at Target as any viewed pair is, and each element's fields are read
// through Target, checked, after a for...of hands it out. The absent label reads undefined and a present one
// its string, as on Node.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): string {
    return String(target.kind) + ":" + (target.label ?? "absent");
}
function each(targets: readonly Target[]): void {
    for (const target of targets) {
        console.log(describe(target));
    }
}
const bare: Source = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const sources: Source[] = [bare, hidden];
each(sources);
