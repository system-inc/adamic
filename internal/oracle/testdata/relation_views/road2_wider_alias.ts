// Language's road 2 on #0b9xqhh: between two checked reads of label through the view, a second, wider alias
// writes a number there. Each read is checked as it runs, so the first prints and the second stops at the
// read (exit 70) where Node prints 2.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Loose { kind: number; label?: string | number; }

const loose: Loose = { kind: 1, label: "first" };
const source: Source = loose;
function twice(target: Target, rewrite: () => void): void {
    console.log(target.label ?? "absent");
    rewrite();
    console.log(target.label ?? "absent");
}
twice(source, () => { loose.label = 2; });
