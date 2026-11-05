const o: { type: "string" | (number | string)[] } = { type: [0, "b"] };
const t = o.type;
if (typeof t !== "string") { t.push("c"); for (const x of t) { console.log(String(x)); } }
