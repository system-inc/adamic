// Language's road 3 on #0b9xqhh: the view is stored in an array, a Map and a closure, then read as Target in a
// function the flow can't follow (a callback, a Map value). A view registers its checks by receiver type and
// member for the whole program, so every read of label through Target is checked whatever road the value took:
// these agree with Node. road3_other_type.ts reads the same values through a second type, which is NotYet.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const labelled = { kind: 1, label: "one" };
const source: Source = labelled;
const viewed: Target = source;
const array: Target[] = [viewed];
const map = new Map<string, Target>();
map.set("only", viewed);
const later = (): Target => viewed;
function read(target: Target): string {
    return target.label ?? "absent";
}
console.log(array.map(read).join(","));
console.log(read(map.get("only") ?? { kind: 0 }));
console.log(read(later()));
