class Box { v = 1; }
let x: Box | Uint8Array = new Box();
function swap(): void { x = new Uint8Array(4); }
if (x instanceof Box) {
    swap();
    console.log(String(x.v));
}
