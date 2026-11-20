// rulings.md:72's second slice (#0b9xqhh), containers: a Source[] going into a mutable Target[] stays refused in
// rule 3300's words (Lower's condition 3): a Target written through the wider array lands in storage typed
// Source[], the invariant-mutable hazard, whatever tsc accepts. nodeFactory.ts:7490's
// (SourceMapRange | undefined)[] is this shape. Its values are honest, so only the refusal stops it.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const labelled = { kind: 1, label: "one" };
const hidden: Source = labelled;
const sources: Source[] = [hidden];
const viewed: Target[] = sources;
for (const target of viewed) {
    console.log(target.label ?? "absent");
}
