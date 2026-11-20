// A relation inside a generic body is lowered once for each instantiation, and each lowering views it. The
// second instantiation's value is widened to Seen, whose label no view checks, and its flow reaches only that
// instantiation's view, so the read through Seen is checked where it's made (#zpzm9ta) and stops at the lie,
// exit 70, where Node prints 5. Were the first lowering to consume the entry, the second would cross with no
// view, the read wouldn't be weighed, and it would be raw, a number where Seen says string.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Seen { readonly kind: number; readonly label?: string; }

function admit<Tag>(tag: Tag, source: Source): Target {
    const target: Target = source;
    return target;
}
const plain: Source = { kind: 1 };
const lying = { kind: 2, label: 5 };
const hidden: Source = lying;
const first = admit(1, plain);
const second = admit("two", hidden);
const seen: Seen = second;
console.log(first.label ?? "absent");
console.log(String(seen.label ?? "absent"));
