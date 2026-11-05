enum Kind { A, B }
const o: { type: "string" | (number | string)[] } = { type: [Kind.A, Kind.B] };
const t = o.type;
if (typeof t !== "string") { t.push("c"); for (const x of t) { console.log(String(x)); } }
