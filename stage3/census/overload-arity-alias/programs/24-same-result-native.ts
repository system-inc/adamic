function pick(a: number): number;
function pick(a: number, f: (x: number) => number): number;
function pick(a: number, f?: (x: number) => number): number {
    if (f) return f(a);
    return a;
}
const r = pick(1);
console.log(String(r + 1));
console.log(r.toFixed(0));
