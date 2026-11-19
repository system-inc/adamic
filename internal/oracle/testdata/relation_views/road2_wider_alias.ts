// Language's road 2 on #0b9xqhh, its optional form: the wider alias's label is optional, and a write of a
// possibly absent optional own field is NotYet until :73's certificate covers it (#x0nd1mh), so this waits on
// :73. road2_required_alias.ts is the same road with the alias's label required, which reaches the reads.
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
