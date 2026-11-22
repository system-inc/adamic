// #tq6yzqy's Record trace: an index-signature dictionary passed where the intersection goes, read
// in the callee, which has no record of its own in sight. Expected: the argument stops at
// recordStorageView (records.go, viewSite's argument), fixed objects and records having different
// storage. Node prints: true, then undefined.
type Entry = { valid?: boolean } & { other?: boolean };
type Table = { [key: string]: boolean };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
const table: Table = {};
table["valid"] = true;
probe(table);
probe({});
