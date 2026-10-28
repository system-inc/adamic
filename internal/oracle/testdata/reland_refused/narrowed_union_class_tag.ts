class Box { v = 1; }
class Other { w = "a"; u = 2; }
let x: Box | Other | number = new Box();
function swap(): void { x = new Other(); }
if (x instanceof Box) {
    swap();
    console.log(String(x.v));
}
