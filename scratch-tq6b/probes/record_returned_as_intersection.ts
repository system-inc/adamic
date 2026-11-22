// #tq6yzqy's Record trace: a record returned from a function whose return type is the intersection.
// The return stops at the optional-widening rule (optional_widening.go:505). Node prints: true.
type Entry = { valid?: boolean } & { other?: boolean };
function made(): Entry {
	const table: Record<string, boolean> = {};
	table["valid"] = true;
	return table;
}
console.log(String(made().valid));
