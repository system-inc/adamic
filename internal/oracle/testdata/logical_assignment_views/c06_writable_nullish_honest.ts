const r: { x: number | string } = { x: 1 };
let w: { x: number | string } | undefined = undefined;
w ??= r;
w.x = "s";
console.log(String(r.x));
