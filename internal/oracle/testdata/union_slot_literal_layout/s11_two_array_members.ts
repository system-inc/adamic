const o: { type: string | number[] | (number | string)[] } = { type: [0, 1] };
const t = o.type;
if (typeof t !== "string") { for (const x of t) { console.log(String(x)); } }
