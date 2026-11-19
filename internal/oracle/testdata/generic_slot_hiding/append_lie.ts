// #4ysqm2m: append(held, hidden). tsc infers append's T from both arguments at one priority, so T is Source,
// the common supertype (inference.ts pins it), and held, a Target[], goes in as a Source[]. append's body then
// pushes a Source whose hidden label is a number, and held's own reads see it as a Target. On main it lowered,
// and native read the number as a string and segfaulted where Node prints 5. The invariant-mutable rule now
// refuses held, whose element a write can hide label from, at the first call.
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
