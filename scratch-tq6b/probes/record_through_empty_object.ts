// #tq6yzqy's Record trace: a record seen as {} first, then {} seen as the intersection, so no single
// relation names both the record and the intersection. Expected: the first hop stops at
// recordStorageView, a Record<string, boolean> seen as {} (records.go; entries_records.go's
// recordEmptyObject comment says a record seen as {} stops there today). Node prints: false.
type Entry = { valid?: boolean } & { other?: boolean };
const table: Record<string, boolean> = { valid: false };
const empty: {} = table;
const entry: Entry = empty;
console.log(String(entry.valid));
