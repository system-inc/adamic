const kinds = new Map<string, number>([["a", 0]]);
const view: ReadonlyMap<string, number | string> = kinds;
for (const [key, value] of view) { console.log(key + " = " + String(value)); }
