// rulings.md:72's second slice (#0b9xqhh), a generic call's argument: checker.ts:19948's shape,
// append(result, getDeclaredTypeOfSymbol(symbol)). append's T is inferred from the array, since a naked
// T | undefined infers at a lower priority, so the value's slot at the call, read from the call's resolved
// signature, is Target | undefined, with no type parameter left in it, and a Source crosses as a view of
// Target. Each read of label through Target is checked and agrees with Node. generic_append_lie.ts is the twin
// that lies.
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
