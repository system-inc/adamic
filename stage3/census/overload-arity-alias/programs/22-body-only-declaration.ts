function pick(a: number): number[];
function pick(a: number, f?: (x: number) => string): number[] | string {
    if (f) return "bad";
    return [a];
}
function f(x: number): string { return "bad"; }
console.log("body-only set accepted");
