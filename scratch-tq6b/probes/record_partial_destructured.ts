// #tq6yzqy's Record trace: record_partial_agreeing.ts's road, read by destructuring instead of a property
// access, so the read takes the binding pattern's road. Agrees or stops; a differs is the record read as a
// fixed object's slot. Node prints: true, then undefined.
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
