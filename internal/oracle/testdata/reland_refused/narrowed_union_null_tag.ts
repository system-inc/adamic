class Box { v = 1; }
let x: Box | number | null = new Box();
function clear(): void { x = null; }
if (x instanceof Box) {
    clear();
    console.log(String(x.v));
}
