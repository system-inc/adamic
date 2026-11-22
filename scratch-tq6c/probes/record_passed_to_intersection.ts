// #tq6yzqy's Record trace: an index-signature dictionary passed where the intersection goes, read
// in the callee, which has no record of its own in sight. The argument stops at the optional-widening
// rule (optional_widening.go:505), valid and other being absent from Table. Node prints: true, then
// undefined.
type Entry = { valid?: boolean } & { other?: boolean };
type Table = { [key: string]: boolean };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
const table: Table = {};
table["valid"] = true;
probe(table);
probe({});
