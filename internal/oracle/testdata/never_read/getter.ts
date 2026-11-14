// A class serving the union's member through a literal-typed getter (#p6vawef, Language's second
// premise for the own-function trust): get kind() can answer only "circle", so two reads of the kind
// can't disagree, and the never read past both checks is trapped and never reached.
interface Circle { readonly kind: "circle"; readonly radius: number }
interface Square { readonly kind: "square"; readonly side: number }
type Shape = Circle | Square;
class Ring implements Circle {
	get kind(): "circle" { return "circle"; }
	readonly radius: number = 2;
}
function corners(shape: Shape): string {
	if (shape.kind === "circle") return "circle " + String(shape.radius);
	if (shape.kind === "square") return "square";
	{ const seen: unknown = shape; return "stray " + typeof seen; }
}
console.log(corners(new Ring()) + " " + corners({ kind: "square", side: 1 }));
