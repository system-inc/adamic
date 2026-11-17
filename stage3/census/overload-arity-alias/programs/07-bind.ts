function pick(a: number): number;
function pick(a: number, f: (x: number) => string): string;
function pick(a: number, f?: (x: number) => string): number | string {
    if (f) return f(a);
    return a;
}
function f(x: number): string { return "bad"; }
const bound = pick.bind(undefined, 1);
const r: number = bound(f);
console.log(r + 1);
console.log(r.toFixed(0));
