// This slice's known stop on #0b9xqhh: a shorthand property's value is read by shorthand as a local, never
// through value, so no lowering views it, and checkRelationViewsTaken refuses it in rule 3300's words. Without
// that post-pass it would cross with no view and read label raw, a number where Target says string (Node
// prints 5).
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Box { readonly target: Target; }

function show(box: Box): string {
    return String(box.target.label ?? "absent");
}
const lying = { kind: 1, label: 5 };
const target: Source = lying;
console.log(show({ target }));
