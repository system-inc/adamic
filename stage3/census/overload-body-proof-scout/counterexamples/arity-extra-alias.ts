function fold<T, U>(xs: readonly T[], step: (m: U, v: T) => U, seed: U): U;
function fold<T>(xs: readonly T[], step: (m: T, v: T) => T): T | undefined;
function fold<T>(xs: readonly T[], step: (m: T, v: T) => T, seed?: T): T | undefined {
    let memo = arguments.length <= 2 ? xs[0] : seed;
    let i = arguments.length <= 2 ? 1 : 0;
    for (; i < xs.length; i++) memo = step(memo!, xs[i]);
    return memo;
}
// Assignability may use the two-argument overload and allow an extra argument.
const wider: (xs: number[], step: (m: number, v: number) => number, extra: string) => number | undefined = fold;
const actual = wider([1], memo => memo, "bad");
console.log(JSON.stringify({ expectedType: "number", actualType: typeof actual, actual }));
