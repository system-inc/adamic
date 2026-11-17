function bad<T, U>(xs: readonly T[], step: (m: U, v: T) => U, seed: U | undefined): U;
function bad<T>(xs: readonly T[], step: (m: T, v: T) => T, seed: T = xs[0]): T {
    let memo = seed;
    if (arguments.length >= 3) {
        for (const value of xs) memo = step(memo, value);
    }
    return memo;
}
const actual = bad([0], (memo: string) => memo, undefined);
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
