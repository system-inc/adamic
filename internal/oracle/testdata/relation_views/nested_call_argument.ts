// Compiler's probe on #0b9xqhh: the relation is a plain argument of a call nested in another's,
// outer(inner(source)), where inner takes the Target. Both calls' arguments are lowered through value, so the
// relation is viewed: the honest call prints, and the lying one stops at the read of label (exit 70) where
// Node prints [5]. nested_call_argument_honest.ts is the twin that agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function inner(target: Target): string {
    return target.label ?? "absent";
}
function outer(text: string): string {
    return "[" + text + "]";
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
console.log(outer(inner(first)));
console.log(outer(inner(second)));
