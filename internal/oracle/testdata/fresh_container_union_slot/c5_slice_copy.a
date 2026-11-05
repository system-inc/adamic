enum Kind { A, B }
interface Option {
    type: "string" | (number | string)[];
}
const kinds: Kind[] = [Kind.A, Kind.B];
const option: Option = { type: kinds.slice() };
const type = option.type;
if (typeof type !== "string") {
    type.push("c");
    console.log(type.map((item) => String(item)).join(","));
}
console.log(String(kinds.length));
