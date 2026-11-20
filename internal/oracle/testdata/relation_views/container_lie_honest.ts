// rulings.md:72's second slice (#0b9xqhh), containers: container_lie.ts's twin, whose second element's hidden
// label is a string. Each read of label through Target passes its check and prints what Node prints.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): string {
    return target.label ?? "absent";
}
const honest = { kind: 1, label: "one" };
const labelled = { kind: 2, label: "two" };
const first: Source = honest;
const second: Source = labelled;
const sources: Source[] = [first, second];
const viewed: readonly Target[] = sources;
for (const target of viewed) {
    console.log(describe(target));
}
