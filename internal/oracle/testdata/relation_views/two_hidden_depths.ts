// Compiler's should on #0b9xqhh: the relation hides two optional fields at two depths, Target's own label and
// Inner's note under the inner Source declares, and both lie. The walk finds label first; recordRelationView
// holds note to the same conditions, and relationViewValue asserts each one registered checked, not only the
// first. The honest call prints all three reads, and the lying one stops at its first, the read of note
// through inner (exit 70), where Node prints 6. two_hidden_depths_honest.ts is the twin that agrees.
interface SourceInner { readonly text: string; }
interface Source { readonly kind: number; readonly inner: SourceInner; }
interface Inner { readonly text: string; readonly note?: string; }
interface Target { readonly kind: number; readonly label?: string; readonly inner: Inner; }

function show(target: Target): void {
    console.log(target.inner.note ?? "no note");
    console.log(target.label ?? "absent");
    console.log(target.inner.text);
}
const honest = { kind: 1, label: "one", inner: { text: "a", note: "first" } };
const lying = { kind: 2, label: 5, inner: { text: "b", note: 6 } };
const first: Source = honest;
const second: Source = lying;
show(first);
show(second);
