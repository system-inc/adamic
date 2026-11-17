function bad(values: number[], map?: (v: number) => string): string[];
function bad(values: number[], map: (v: number) => string | number = v => v): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of values) out.push(map(value));
    return out;
}
const actual = bad([0])[0];
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
