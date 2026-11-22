// { valid?: boolean } & { valid?: true } (#tq6yzqy): the intersection reads valid as true | undefined,
// which is held as an optional boolean as boolean | undefined is, so the members agree. No value of the
// intersection holds false, since { valid: false } isn't a { valid?: true }. Written true, read, then
// absent, read, then written true. Node prints: true, then undefined, then true.
type Entry = { valid?: boolean } & { valid?: true } & { name: string };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
probe({ name: "entry", valid: true });
const entry: Entry = { name: "entry" };
probe(entry);
entry.valid = true;
probe(entry);
