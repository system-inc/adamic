// #tq6yzqy's Record trace: a Record<string, boolean> seen as an intersection whose fields are all
// optional booleans. tsc allows it (the target is weak, but the record has no properties to share, so
// the weak check never runs), and the read takes the intersection's field. It stops at the relation, by
// the optional-widening rule (optional_widening.go:505, words at :531): valid and other are optional in
// Entry and absent from the record. Node prints: false, then undefined.
type Entry = { valid?: boolean } & { other?: boolean };
const table: Record<string, boolean> = { valid: false };
const entry: Entry = table;
console.log(String(entry.valid));
console.log(String(entry.other));
