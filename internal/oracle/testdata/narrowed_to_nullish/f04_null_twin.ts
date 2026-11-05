class B { x: number | null = null; }
const b = new B();
function f(): void { b.x = 4; }
b.x = null;
f();
console.log(String(b.x));
