function pick(a: number): number;
function pick(a: number, f: (x: number) => string): string;
function pick(a: number, f?: (x: number) => string): number | string {
    if (f) return f(a);
    return a;
}
function f(x: number): string { return "bad"; }
const g: (a: number, f: (x: number) => string) => number = pick;
const args: [number, (x: number) => string] = [1, f];
const r = g(...args);
console.log(r + 1);
console.log(r.toFixed(0));
