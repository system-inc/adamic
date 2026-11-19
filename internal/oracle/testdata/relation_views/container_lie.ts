// rulings.md:72's second slice (#0b9xqhh), containers: a Source[] viewed as a readonly Target[] holds an
// element whose hidden label is a number. Its element read passes, since the element's contract is told by its
// kind, and the read of label through Target stops (exit 70) where Node prints 5. container_lie_honest.ts is
// the twin that agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): string {
    return target.label ?? "absent";
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
const sources: Source[] = [first, second];
const viewed: readonly Target[] = sources;
for (const target of viewed) {
    console.log(describe(target));
}
