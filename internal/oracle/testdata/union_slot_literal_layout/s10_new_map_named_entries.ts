const entries: [string, number][] = [["a", 1]];
const v: Map<string, number | string> = new Map(entries);
v.set("b", "two");
for (const [key, value] of v) { console.log(key + " = " + String(value)); }
