// two_hidden_depths.ts's honest twin: label and note, the two fields hidden at two depths, each hold a
// string or nothing.
interface SourceInner { readonly text: string; }
interface Source { readonly kind: number; readonly inner: SourceInner; }
interface Inner { readonly text: string; readonly note?: string; }
interface Target { readonly kind: number; readonly label?: string; readonly inner: Inner; }

function show(target: Target): void {
    console.log(target.inner.note ?? "no note");
    console.log(target.label ?? "absent");
    console.log(target.inner.text);
}
const plain = { kind: 1, inner: { text: "a" } };
const labelled = { kind: 2, label: "two", inner: { text: "b", note: "second" } };
const first: Source = plain;
const second: Source = labelled;
show(first);
show(second);
