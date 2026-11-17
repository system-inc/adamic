Object.defineProperty(Array.prototype, "push", {
    value: function(this: (string | number)[]) { this[0] = 0; return 1; },
});
function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of ["ok"]) out.push(value);
    return out;
}
const actual = bad()[0];
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
