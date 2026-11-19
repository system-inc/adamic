// optional_method_argument.ts's honest twin: the argument of a?.read(source) holds a string label or none,
// and a missing reader short-circuits the call.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

class Reader {
    read(target: Target): string {
        return target.label ?? "absent";
    }
}
function find(present: boolean): Reader | undefined {
    return present ? new Reader() : undefined;
}
const plain = { kind: 1 };
const labelled = { kind: 2, label: "two" };
const first: Source = plain;
const second: Source = labelled;
console.log(find(true)?.read(first) ?? "none");
console.log(find(true)?.read(second) ?? "none");
console.log(find(false)?.read(second) ?? "none");
