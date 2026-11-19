// optional_chain_operand.ts's honest twin: the argument of f(source)?.kind holds a string label or none.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): Target | undefined {
    console.log(target.label ?? "absent");
    return target.kind > 0 ? target : undefined;
}
const plain = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const first: Source = plain;
const second: Source = labelled;
console.log(String(describe(first)?.kind ?? 0));
console.log(String(describe(second)?.kind ?? 0));
