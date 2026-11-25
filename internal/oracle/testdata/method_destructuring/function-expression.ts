interface H { f(x: number): number }
const h: H = { f: function (x: number): number { return x; } };
const { f } = h;
console.log(String(f(2)));
