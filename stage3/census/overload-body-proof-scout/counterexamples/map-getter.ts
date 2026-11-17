function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    let reads = 0;
    const source = { get value(): string | number { return ++reads === 1 ? "ok" : 0; } };
    for (const unused of [0]) {
        if (typeof source.value === "string") out.set("key", source.value);
    }
    return out;
}
const actual = bad().get("key");
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
