interface Options { x?: number | null; [key: string]: number | null | undefined; }
const o: Options = { x: 1 };
function f(): void { delete o.x; }
o.x = null;
f();
console.log(String(o.x));
