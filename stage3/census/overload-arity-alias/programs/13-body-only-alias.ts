function pick(a: number): number[];
function pick(a: number, f?: (x: number) => string): number[] | string {
    if (f) return "bad";
    return [a];
}
function f(x: number): string { return "bad"; }
const g: (a: number, f: (x: number) => string) => number[] = pick;
const r = g(1, f)[0] ?? 0;
console.log(String(r + 1));
console.log(r.toFixed(0));
