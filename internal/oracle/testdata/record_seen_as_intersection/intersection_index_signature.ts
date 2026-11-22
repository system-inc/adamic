// The source side (#mft3zc6, J-rmbk4bb's probe): a fixed intersection value seen as an index-signature
// dictionary, read by a literal and a computed key. On main it compiled and natively was an
// AddressSanitizer SEGV, where Node prints: 3, undefined. It stops at the relation.
type Entry = { count?: number } & { label?: number };
const entry: Entry = { count: 3 };
const table: { [key: string]: number | undefined } = entry;
const key = "label";
console.log(String(table["count"]));
console.log(String(table[key]));
