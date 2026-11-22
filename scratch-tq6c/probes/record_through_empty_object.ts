// #tq6yzqy's Record trace: a record seen as {} first, then {} seen as the intersection, so no single
// relation names both the record and the intersection. The first hop stops at recordStorageView
// (records.go:551), a Record<string, boolean> seen as {}, since {} is an object type with no record
// element. Node prints: false.
type Entry = { valid?: boolean } & { other?: boolean };
const table: Record<string, boolean> = { valid: false };
const empty: {} = table;
const entry: Entry = empty;
console.log(String(entry.valid));
