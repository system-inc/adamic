enum Kind { A, B }
const kinds: Kind[] = [Kind.A];
const option: { type: string | Map<string, (number | string)[]> } = { type: new Map([["a", kinds]]) };
const type = option.type;
if (type instanceof Map) {
    const list = type.get("a");
    if (list !== undefined) {
        list.push("lie");
    }
}
console.log(kinds.map((kind) => String(kind)).join(","));
