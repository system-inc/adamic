const entries: Iterable<[string, number | string]> = new Map<string, number>([["a", 1]]);
for (const [key, value] of entries) { console.log(key + " = " + String(value)); }
