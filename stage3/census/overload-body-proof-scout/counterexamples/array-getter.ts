function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    let reads = 0;
    const source = { get value(): string | number { return ++reads === 1 ? "ok" : 0; } };
    for (const unused of [0]) {
        if (typeof source.value === "string") out.push(source.value);
    }
    return out;
}
const actual = bad()[0];
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
