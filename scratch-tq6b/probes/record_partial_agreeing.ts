// #tq6yzqy's Record trace, the road the relation rules leave open: Partial<Record<"valid", boolean>> is held
// as a record (records_partial.go) and declares valid, so seen as an intersection whose members each declare
// valid?: boolean, no optional field is absent (the optional-widening rule has nothing to refuse), and
// recordStorageView answers same storage, since sameRecordStorage's present() is nil for an intersection
// (records.go:711, :719). The read must then take weakFieldRead's storage test (object.go:729-744), reading
// the record's key, or stop. A fixed object reaches the same read. Wrong would be a differs or a sanitizer
// report. Node prints: false, then undefined, then true.
type Flags = Partial<Record<"valid", boolean>>;
type Entry = { valid?: boolean } & { valid?: boolean };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
const set: Flags = { valid: false };
const unset: Flags = {};
probe(set);
probe(unset);
const direct: Entry = { valid: true };
probe(direct);
