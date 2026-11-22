// { valid?: boolean } & { valid?: true } (#tq6yzqy): the intersection reads valid as true | undefined,
// which is held as an optional boolean as boolean | undefined is, so the members agree. No value of the
// intersection holds false, since { valid: false } isn't a { valid?: true }. One value is built with
// valid true and one without, and each is read. Node prints: true, then undefined.
type Entry = { valid?: boolean } & { valid?: true } & { name: string };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
const present: Entry = { name: "present", valid: true };
const absent: Entry = { name: "absent" };
probe(present);
probe(absent);
