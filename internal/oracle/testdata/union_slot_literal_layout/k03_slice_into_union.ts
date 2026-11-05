const kinds: number[] = [0, 1];
const t: string | (number | string)[] = kinds.slice();
if (typeof t !== "string") { t.push("c"); for (const x of t) { console.log(String(x)); } }
