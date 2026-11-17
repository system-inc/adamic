function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    const alias = out;
    for (const value of ["ok"]) out.push(value);
    alias.push(0);
    return out;
}
const actual = bad()[1];
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
