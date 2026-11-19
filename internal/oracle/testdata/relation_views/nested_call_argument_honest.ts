// nested_call_argument.ts's honest twin: inner's argument holds a string label or none.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function inner(target: Target): string {
    return target.label ?? "absent";
}
function outer(text: string): string {
    return "[" + text + "]";
}
const plain = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const first: Source = plain;
const second: Source = labelled;
console.log(outer(inner(first)));
console.log(outer(inner(second)));
