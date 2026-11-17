function bad<T, U>(xs: readonly T[], step: (m: U, v: T) => U, seed: U): U;
function bad<T>(xs: readonly T[], step: (m: T, v: T) => T, seed?: T): T | undefined {
    let memo = seed;
    if (arguments.length >= 3) {
        // The first callback argument is NOT the previous accumulator.
        for (const value of xs) memo = step(value, value);
    }
    return memo;
}
const actual = bad([0], (memo: string) => memo, "ok");
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
