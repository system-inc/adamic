enum Kind { A, B }
type Slot = string | Map<string, readonly Kind[]> | Map<string, (number | string)[]>;
const kinds: Kind[] = [Kind.A];
const option: { type: Slot } = { type: new Map<string, Kind[]>([["a", kinds]]) };
const type = option.type;
if (type instanceof Map) {
    const list = type.get("a");
    if (list !== undefined) {
        console.log(String(list.length));
    }
}
console.log(kinds.map((kind) => String(kind)).join(","));
