interface Options { readonly overrideImportMode?: number; readonly other?: string; }
type Shape = { readonly kind: "circle"; readonly radius: number } | { readonly kind: "square"; readonly side: number };
const shapes: Shape[] = [{ kind: "circle", radius: 1 }, { kind: "square", side: 2 }];
const circles = shapes.filter((shape): shape is { readonly kind: "circle"; readonly radius: number } => shape.kind === "circle");
console.log(String(circles.length));
