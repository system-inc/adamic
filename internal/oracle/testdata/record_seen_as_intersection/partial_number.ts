// A finite Partial<Record<K, V>> is held as a record (records_partial.go), and it declares every field of
// an intersection of plain shapes, which is held as a fixed object (#mft3zc6). So no optional field is
// absent for the optional-widening rule to refuse, and the argument stops at recordStorageView instead:
// a record and a fixed object have different storage. It compiled on main and panicked natively,
// "compiler bug: a record reached a fixed-object read of count", where Node prints: 1, undefined, 2.
type Entry = { count?: number } & { count?: number };
function probe(value: Entry): void {
	console.log(String(value.count));
}
const set: Partial<Record<"count", number>> = { count: 1 };
const unset: Partial<Record<"count", number>> = {};
probe(set);
probe(unset);
const direct: Entry = { count: 2 };
probe(direct);
