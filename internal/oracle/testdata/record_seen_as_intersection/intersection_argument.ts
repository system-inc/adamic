// The source side (#mft3zc6) through an argument: the fixed intersection value passed to a function
// whose parameter is a finite Partial record, read there through the record. It stops at the same
// relation as the declaration does. Node prints: 3.
type Entry = { count?: number } & { label?: number };
function total(table: Partial<Record<"count" | "label", number>>): void {
	console.log(String(table.count));
}
const entry: Entry = { count: 3 };
total(entry);
