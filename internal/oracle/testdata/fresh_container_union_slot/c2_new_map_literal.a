interface Option {
    type: "string" | "number" | Map<string, number | string>;
}
const option: Option = { type: new Map([["a", 1]]) };
const type = option.type;
if (type instanceof Map) {
    type.set("b", "two");
    for (const [key, value] of type) {
        console.log(key + " = " + String(value));
    }
}
