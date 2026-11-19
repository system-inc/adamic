// Compiler's probe on #0b9xqhh: the relation is the argument of a call in a template literal's span,
// `${label(source)}`. The span's value is lowered through value, and so is the call's argument, so it's
// viewed: the honest span prints, and the lying one stops at the read of label (exit 70) where Node prints
// 5. template_span_argument_honest.ts is the twin that agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function label(target: Target): string {
    return target.label ?? "absent";
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
console.log(`first: ${label(first)}`);
console.log(`second: ${label(second)}`);
