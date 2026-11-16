// A union whose other member's slot is the whole Kind takes the open value as that member, whole, with
// no decomposition (#egmxf20): a slot of one literal in a member the literal isn't built as is no
// promise to it, and it lowers.
enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius: number }
interface Labeled { readonly kind: Kind; readonly label: string }
type Shape = Circle | Labeled;
function show(shape: Shape): string {
	return String(shape.kind);
}
function other(): number { return 7; }
const kind: Kind = other();
console.log(show({ kind, label: "x" }) + " " + show({ kind: kind, label: "y" }) + " " + show({ kind: Kind.Circle, radius: 1 }));
