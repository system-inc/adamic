function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    for (const value of ["ok"]) out.set("key", value);
    out.set("key", 0);
    return out;
}
const actual = bad().get("key");
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
