function pick(a: number): number;
function pick(a: number, f: (x: number) => string): string;
function pick(a: number, f?: (x: number) => string): number | string {
    if (f) return f(a);
    return a;
}
function f(x: number): string { return "bad"; }
function invoke(g: (a: number, f: (x: number) => string) => number): number { return g(1, f); }
const r = invoke(pick);
console.log(r + 1);
console.log(r.toFixed(0));
