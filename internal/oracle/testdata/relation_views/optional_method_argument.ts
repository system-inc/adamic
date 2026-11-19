// Compiler's probe on #0b9xqhh: the relation is the argument of an optional method call, a?.read(source),
// lowered through the object's method and its arguments through value, so it's viewed. The read of label in
// the method is checked: the honest call prints, and the lying one stops at the read (exit 70) where Node
// prints 5. optional_method_argument_honest.ts is the twin that agrees.
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
const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
console.log(find(true)?.read(first) ?? "none");
console.log(find(true)?.read(second) ?? "none");
