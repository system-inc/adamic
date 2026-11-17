function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    const later = () => out.set("key", 0);
    for (const value of ["ok"]) out.set("key", value);
    later();
    return out;
}
const actual = bad().get("key");
console.log(JSON.stringify({ expectedType: "string", actualType: typeof actual, actual }));
