const o: { type: "string" | number[] } = { type: [0, 1] };
const t = o.type;
if (typeof t !== "string") { t.push(2); for (const x of t) { console.log(String(x)); } }
