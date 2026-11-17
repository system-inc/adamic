function pick(a: number): number[];
function pick(a: number, f?: (x: number) => string): number[] | string {
    if (f) return "bad";
    return [a];
}
function f(x: number): string { return "bad"; }
const r = pick(1)[0];
console.log(r + 1);
console.log(r.toFixed(0));
