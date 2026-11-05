const v: Map<string, number | string> = new Map([["a", 1]]);
v.set("b", "two");
for (const [key, value] of v) { console.log(key + " = " + String(value)); }
