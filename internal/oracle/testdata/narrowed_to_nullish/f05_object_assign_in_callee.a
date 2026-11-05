const o: { x: number | undefined } = { x: 1 };
function f(): void { Object.assign(o, { x: 4 }); }
o.x = undefined;
f();
console.log(String(o.x));
