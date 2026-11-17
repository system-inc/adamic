function pick(a: number): number;
function pick(a: number, f: (x: number) => string): string;
function pick(a: number, f?: (x: number) => string): number | string {
    if (f) return f(a);
    return a;
}
function f(x: number): string { return "bad"; }
const o: { g: (a: number, f: (x: number) => string) => number } = { g: pick };
const r = o.g(1, f);
console.log(r + 1);
console.log(r.toFixed(0));
