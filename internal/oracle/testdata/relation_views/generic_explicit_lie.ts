// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: append<Target> fixes T, so held is
// Target[] seen as itself and the value's slot at the call, the resolved signature's instantiated parameter,
// is Target | undefined. A Source crosses there as a view of Target, so the lie's label, a number, stops at its
// checked read (exit 70) where Node prints 5. Read from append's declaration instead, the slot would be
// T | undefined, which relates nothing to Source, and the lie would cross raw: mutant (i)'s witness.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

function append<T>(to: T[] | undefined, value: T | undefined): T[] | undefined {
    if (value === undefined) {
        return to;
    }
    if (to === undefined) {
        return [value];
    }
    to.push(value);
    return to;
}
function describe(target: Target): string {
    return target.label ?? "absent";
}
const bare: Source = { kind: 1 };
const lying = { kind: 2, label: 5 };
const hidden: Source = lying;
const held: Target[] = [];
append<Target>(held, bare);
append<Target>(held, hidden);
for (const target of held) {
    console.log(describe(target));
}
