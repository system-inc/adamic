// Language's road 3 on #0b9xqhh, read through a second type: the view's value comes out of a Map, a road the
// flow can't follow, and is read as Seen, whose label no view checks. checkViewWidenedReads can't tell it from
// a plain value, so the read is NotYet, never raw. Node prints one.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Seen { readonly kind: number; readonly label?: string; }

const labelled = { kind: 1, label: "one" };
const source: Source = labelled;
const viewed: Target = source;
const map = new Map<string, Seen>();
map.set("only", viewed);
function read(seen: Seen): string {
    return seen.label ?? "absent";
}
console.log(read(map.get("only") ?? { kind: 0 }));
