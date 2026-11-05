const o: { x?: number | null } = { x: 1 };
function f(): void { delete o.x; }
o.x = null;
f();
console.log(String(o.x));
