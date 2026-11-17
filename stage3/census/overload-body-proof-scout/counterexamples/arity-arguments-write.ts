function bad<T, U>(xs: readonly T[], step: (m: U, v: T) => U, seed: U): U;
function bad<T>(xs: readonly T[], step: (m: T, v: T) => T, seed?: T): T | undefined {
    if (arguments.length >= 3) {
        arguments[2] = "bad";
        let memo: T = arguments[2];
        for (const value of xs) memo = step(memo, value);
        return memo;
    }
    return seed;
}
const actual = bad([1], memo => memo, 0);
console.log(JSON.stringify({ expectedType: "number", actualType: typeof actual, actual }));
