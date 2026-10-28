class Box { v = 1; }
let x: Box | Record<string, number> = new Box();
function swap(): void { const d: Record<string, number> = {}; d["v"] = 5; x = d; }
if (x instanceof Box) {
    swap();
    console.log(String(x.v));
}
