function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    const later = () => out.push(0);
    for (const value of ["ok"]) out.push(value);
    later();
    return out;
}
const actual = bad()[1];
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
