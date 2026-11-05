class B { x: number | undefined = undefined; }
const b = new B();
function f(): void { b.x = 4; }
f();
console.log(String(b.x));
