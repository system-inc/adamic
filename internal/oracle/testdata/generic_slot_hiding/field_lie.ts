// #4ysqm2m, the writable-field road: a Holder, whose t is a Target, seen as a Loose, whose t is a Source. A
// write through the Loose stores a Source whose hidden label is a number, and the read through Holder takes it
// as a Target. tsc accepts the write, since a Source is assignable to a Target. The invariant-mutable rule now
// refuses the Holder seen as Loose, where Node prints 5.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Holder { t: Target; }
interface Loose { t: Source; }

function describe(target: Target): string {
    return target.label ?? "absent";
}
const holder: Holder = { t: { kind: 1, label: "one" } };
const loose: Loose = holder;
const lying = { kind: 2, label: 5 };
const hidden: Source = lying;
console.log(describe(holder.t));
loose.t = hidden;
console.log(describe(holder.t));
