type Value = string | number | null | undefined;
const entries = new Map<string, Value>();
entries.set("null", null);
entries.set("undefined", undefined);
entries.set("number", 7);

function print(value: Value): void {
    if (value === null) console.log("null");
    else if (value === undefined) console.log("undefined");
    else if (typeof value === "number") console.log(`${value}`);
    else console.log(value);
}

print(entries.get("null"));
console.log(entries.has("undefined") ? "true" : "false");
print(entries.get("undefined"));
console.log(entries.has("missing") ? "true" : "false");
print(entries.get("number"));
