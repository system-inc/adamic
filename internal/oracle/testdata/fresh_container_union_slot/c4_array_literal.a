enum Kind { A, B }
interface Option {
    type: "string" | (number | string)[];
}
const option: Option = { type: [Kind.A, Kind.B] };
const type = option.type;
if (typeof type !== "string") {
    type.push("c");
    console.log(type.map((item) => String(item)).join(","));
}
