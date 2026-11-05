enum Kind { A, B }
const kinds = new Map<string, Kind>();
kinds.set("a", Kind.A);
const option: { type: string | Map<string, number | string> } = { type: kinds };
const type = option.type;
if (type instanceof Map) {
    type.set("b", "lie");
}
for (const [key, value] of kinds) {
    console.log(key + " = " + String(value));
}
