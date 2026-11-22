// A fixed intersection value seen as an index-signature dictionary, read by a literal and a computed key.
// Node prints: 3, undefined.
type Entry = { count?: number } & { label?: number };
const entry: Entry = { count: 3 };
const table: { [key: string]: number | undefined } = entry;
const key = "label";
console.log(String(table["count"]));
console.log(String(table[key]));
