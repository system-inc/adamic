const v: Set<number | string> = new Set([0, 1]);
v.add("c");
for (const x of v) { console.log(String(x)); }
