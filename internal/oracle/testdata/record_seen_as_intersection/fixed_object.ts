// The control for partial_number.ts (#mft3zc6): fixed objects, a const literal of a plain type and one
// of the intersection itself, seen as the same intersection, are judged field by field as the fixed
// object it's held as, and lower. Node prints: 1, undefined, 2.
type Plain = { count?: number; label: string };
type Entry = { count?: number } & { count?: number };
function probe(value: Entry): void {
	console.log(String(value.count));
}
const set: Plain = { count: 1, label: "set" };
const unset: Plain = { label: "unset" };
probe(set);
probe(unset);
const direct: Entry = { count: 2 };
probe(direct);
