const list: (string | (number | string)[])[] = [[0, 1]];
const t = list[0];
if (t !== undefined && typeof t !== "string") { t.push("c"); for (const x of t) { console.log(String(x)); } }
