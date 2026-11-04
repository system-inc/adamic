type Circle = { readonly kind: "circle"; readonly radius?: number };
type Square = { readonly kind: "square"; readonly side: number };
type Shape = Circle | Square;
const shapes: Shape[] = [{ kind: "circle", radius: 3 }, { kind: "square", side: 2 }, { kind: "circle" }];
const circles = shapes.filter((shape): shape is { readonly kind: "circle" } => shape.kind === "circle");
console.log(String(circles.length));
for (const circle of circles) console.log(circle.kind);
for (const shape of shapes) if (shape.kind === "circle") console.log(String(shape.radius));
