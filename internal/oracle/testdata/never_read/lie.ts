// trusted.ts's corners handed a value its type lies about (#p6vawef): a computed key writes "triangle"
// where Drawing's shape is declared a Shape, as records_union_lying.ts writes a number into a string.
// Node reaches fail and throws, and Adamic stops at the never read itself, loudly, exit 70, naming the
// read and the switch that made it never (rulings.md:74).
type Shape = "circle" | "square";
interface Drawing { shape?: Shape; [name: string]: string | undefined; }
function fail(shape: unknown): never { throw new Error("unexpected shape " + String(shape)); }
function corners(shape: Shape): number {
	switch (shape) {
		case "circle": return 0;
		case "square": return 4;
	}
	return fail(shape);
}
console.log(String(corners("circle")));
const key: string = "shape";
const drawing: Drawing = { [key]: "triangle" };
if (drawing.shape !== undefined) console.log(String(corners(drawing.shape)));
