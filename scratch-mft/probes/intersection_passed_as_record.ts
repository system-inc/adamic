// #mft3zc6's source side through an argument: a fixed intersection value passed to a parameter typed as
// a finite Partial record, read there. It must stop at the same relation. Node prints: 3, undefined.
type Entry = { count?: number } & { label?: number };
function total(table: Partial<Record<"count" | "label", number>>): void {
	console.log(String(table.count));
	console.log(String(table.label));
}
const entry: Entry = { count: 3 };
total(entry);
