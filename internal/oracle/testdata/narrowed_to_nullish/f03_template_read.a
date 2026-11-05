class B { x: number | undefined = undefined; }
const b = new B();
function f(): void { b.x = 4; }
b.x = undefined;
f();
console.log(`${b.x}`);
console.log(String(b.x === undefined));
