// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: generic_append.ts's twin, whose second
// value hides a label that's a number. The value crosses as a view of Target at append's resolved slot, and
// the read of its label through Target stops (exit 70) where Node prints 5.
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
append(held, bare);
append(held, hidden);
for (const target of held) {
    console.log(describe(target));
}
