const o: { x: number | undefined } = { x: 1 };
function f(): void { console.log('f'); }
o.x = undefined;
f();
console.log(String(o.x));
