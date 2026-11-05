const kinds = new Map<string, number>([["a", 0]]);
const view: string | ReadonlyMap<string, number | string> = kinds;
if (typeof view !== "string") { for (const [key, value] of view) { console.log(key + " = " + String(value)); } }
