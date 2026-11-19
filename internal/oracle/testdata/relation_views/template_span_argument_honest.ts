// template_span_argument.ts's honest twin: the argument of the span's call holds a string label or none.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function label(target: Target): string {
    return target.label ?? "absent";
}
const plain = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const first: Source = plain;
const second: Source = labelled;
console.log(`first: ${label(first)}`);
console.log(`second: ${label(second)}`);
