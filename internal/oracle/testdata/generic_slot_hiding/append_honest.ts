// #4ysqm2m: an append whose values are Targets, so T is Target, held goes in as itself and nothing a write
// stores can hide label. It lowers and agrees with Node.
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
const bare: Target = { kind: 1 };
const labelled: Target = { kind: 2, label: "two" };
const held: Target[] = [];
append(held, bare);
append(held, labelled);
for (const target of held) {
    console.log(describe(target));
}
