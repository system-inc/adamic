class Box { v = 1; }
let x: Box | Promise<number> = new Box();
function swap(): void { x = Promise.resolve(5); }
if (x instanceof Box) {
    swap();
    console.log(String(x.v));
}
