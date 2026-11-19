// Lower's second row on #0b9xqhh: the absent optional field's type is a union of objects, whose contract
// view builds as a union's (a top-level union target would be legacyView's, which this slice refuses). The
// absent and honest reads print, and the lying side stops at its read (exit 70) where Node prints wide.
interface Source { readonly kind: number; }
interface Circle { readonly tag: "circle"; readonly radius: number; }
interface Square { readonly tag: "square"; readonly side: number; }
interface Target { readonly kind: number; readonly shape?: Circle | Square; }

function size(target: Target): string {
    const shape = target.shape;
    if (shape === undefined) {
        return "absent";
    }
    return shape.tag === "circle" ? String(shape.radius) : String(shape.side);
}
const bare: Source = { kind: 1 };
const honest = { kind: 2, shape: { tag: "circle" as const, radius: 2 } };
const lying = { kind: 3, shape: { tag: "square" as const, side: "wide" } };
const truthful: Source = honest;
const liar: Source = lying;
console.log(size(bare));
console.log(size(truthful));
console.log(size(liar));
