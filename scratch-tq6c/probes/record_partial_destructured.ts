// #tq6yzqy's Record trace: partial_boolean.ts's road (internal/oracle/testdata/record_seen_as_intersection),
// read by destructuring instead of a property access, so the read takes the binding pattern's road. The
// argument stops at recordStorageView (#mft3zc6) at 14:7 unless the destructuring stops first; it's a probe,
// not a pin, since which stop comes first isn't proven by reading. It must stop on both backends; a differs
// is the record read as a fixed object's slot. Node prints: true, then undefined.
type Flags = Partial<Record<"valid", boolean>>;
type Entry = { valid?: boolean } & { valid?: boolean };
function probe(value: Entry): void {
	const { valid } = value;
	console.log(String(valid));
}
const set: Flags = { valid: true };
const unset: Flags = {};
probe(set);
probe(unset);
