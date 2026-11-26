type OptionRecord = { [name: string]: string[] };
type CompilerOptionsValue = string | number | boolean | (string | number)[] | OptionRecord | null | undefined;

function describe(value: CompilerOptionsValue): string {
    if (value === null) return "null";
    if (value === undefined) return "undefined";
    if (typeof value === "object") return "object";
    if (typeof value === "string") return "string:" + value;
    if (typeof value === "number") return `number:${value}`;
    return value ? "boolean:true" : "boolean:false";
}

const entries = new Map<string, CompilerOptionsValue>();
const list: (string | number)[] = ["adamic", 42];
const record: OptionRecord = { tags: ["example"] };

entries.set("null", null);
entries.set("undefined", undefined);
entries.set("string", "text");
entries.set("number", 7);
entries.set("boolean", true);
entries.set("array", list);
entries.set("record", record);

for (const key of ["null", "undefined", "string", "number", "boolean", "array", "record", "missing"]) {
    console.log(key + ":" + (entries.has(key) ? "true" : "false") + ":" + describe(entries.get(key)));
}
console.log("array identity:" + (entries.get("array") === list ? "true" : "false"));
console.log("record identity:" + (entries.get("record") === record ? "true" : "false"));
console.log(`size:${entries.size}`);
entries.set("null", undefined);
console.log("overwritten:" + (entries.has("null") ? "true" : "false") + ":" + describe(entries.get("null")));
entries.set("null", null);
console.log("restored:" + (entries.has("null") ? "true" : "false") + ":" + describe(entries.get("null")));
entries.delete("number");
entries.set("number", 7);

entries.forEach((value, key) => { console.log("forEach:" + key + ":" + describe(value)); });
for (const key of entries.keys()) console.log("keys:" + key);
for (const value of entries.values()) console.log("values:" + describe(value));
for (const [key, value] of entries.entries()) console.log("entries:" + key + ":" + describe(value));
for (const [key, value] of entries) console.log("for-of:" + key + ":" + describe(value));
for (const pair of [...entries]) console.log("spread:" + pair[0] + ":" + describe(pair[1]));
for (const value of Array.from(entries.values())) console.log("Array.from:" + describe(value));
