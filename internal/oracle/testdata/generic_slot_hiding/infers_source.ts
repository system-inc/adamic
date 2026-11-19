// #4ysqm2m: tsc's own word on append's T. append(held, hidden) infers T from both arguments at one priority,
// held's Target and hidden's Source, and takes their common supertype, Source, so r is Source[] | undefined
// and reading label off its element is TS2339. It doesn't type-check, and its row holds that diagnostic: the
// cause the fix stands on is the checker's, not a reading of it.
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
const hidden: Source = { kind: 1 };
const held: Target[] = [];
const r = append(held, hidden);
console.log(r?.[0]?.label ?? "none");
