// Compiler's probe on #0b9xqhh: the relation is the argument of a call that is an optional chain's operand,
// f(source)?.kind. The argument is lowered through value as any argument is, so it's viewed, and the read of
// label inside the call is checked: the honest call prints, and the lying one stops at the read (exit 70)
// where Node prints 5. optional_chain_operand_honest.ts is the twin that agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function describe(target: Target): Target | undefined {
    console.log(target.label ?? "absent");
    return target.kind > 0 ? target : undefined;
}
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
console.log(describe(first)?.kind ?? 0);
console.log(describe(second)?.kind ?? 0);
