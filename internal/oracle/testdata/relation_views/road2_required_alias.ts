// Language's road 2 on #0b9xqhh: between two checked reads of label through the view, a second, wider alias
// that isn't the view writes a number there. Its label is required, so the write is a plain write of its own
// field. Each read through the view is checked as it runs, never once at the crossing: the first prints, and
// the second stops at the read (exit 70) where Node prints 2.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Loose { kind: number; label: string | number; }

const loose: Loose = { kind: 1, label: "first" };
const source: Source = loose;
function twice(target: Target, rewrite: () => void): void {
    console.log(target.label ?? "absent");
    rewrite();
    console.log(target.label ?? "absent");
}
twice(source, () => { loose.label = 2; });
