function bad(values: number[], makeValue: ((v: number) => string) | undefined): Map<string, string>;
function bad(values: number[], makeValue: (v: number) => string | number = v => v): Map<string, string | number> {
    const out = new Map<string, string | number>();
    for (const value of values) out.set("key", makeValue(value));
    return out;
}
// The second argument was supplied: arity does NOT suppress the default.
const actual = bad([0], undefined).get("key");
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
