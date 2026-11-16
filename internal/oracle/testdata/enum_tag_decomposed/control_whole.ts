// A slot typed the whole Kind takes the open value (#egmxf20): it promises no one member, and lowers.
enum Kind { Circle, Square }
interface Tagged { readonly tag: "tagged"; readonly kind: Kind; readonly size: number }
interface Named { readonly tag: "named"; readonly name: string }
type Item = Tagged | Named;
function describe(item: Item): string {
	if (item.tag === "tagged") return "tagged " + String(item.kind) + " " + String(item.size);
	return "named " + item.name;
}
function other(): number { return 7; }
const kind: Kind = other();
const held: Tagged = { tag: "tagged", kind, size: 3 };
console.log(describe({ tag: "tagged", kind, size: 2 }) + " " + describe(held) + " " + describe({ tag: "named", name: "x" }));
