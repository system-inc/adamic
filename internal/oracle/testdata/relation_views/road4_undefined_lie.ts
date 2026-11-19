// Language's road 4 on #0b9xqhh, the lie: a Source typed slot made to hold undefined through a wider, writable
// alias, then handed to Target | undefined, where the widening views a source holding no undefined as Target.
// The lie never reaches the view: the alias that could write undefined where Source is read is refused
// (adamic/invariant-mutable). Node prints none.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Holder { source: Source; }
interface LooseHolder { source: Source | undefined; }

function show(target?: Target): string {
    return target === undefined ? "none" : target.label ?? "absent";
}
const holder: Holder = { source: { kind: 1 } };
const loose: LooseHolder = holder;
loose.source = undefined;
console.log(show(holder.source));
