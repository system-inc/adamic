const kinds: number[] = [0, 1];
const t: (number | string)[] = Array.from(kinds);
t.push("c");
for (const x of t) { console.log(String(x)); }
