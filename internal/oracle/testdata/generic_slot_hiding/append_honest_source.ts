// #4ysqm2m: append_lie.ts with honest values, refused all the same. Its second Source holds a string label,
// but held is still a Target[] seen as Source[], which the type lets take any Source. The rule judges what the
// type lets through, never which value flows (Lower, as the declared-keys rule does, #npqr3sx), so this
// refusal is the price of the fix, not a false one. append_honest.ts, whose slot really is Target, is the
// program that keeps lowering.
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
const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const held: Target[] = [];
append(held, bare);
append(held, hidden);
for (const target of held) {
    console.log(describe(target));
}
