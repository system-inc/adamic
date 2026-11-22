// A member holds valid as a union (#tq6yzqy): { valid?: boolean | string } is boxed, so the members
// don't agree, and the intersection keeps the optional-boolean stop though tsc reads valid as
// boolean | undefined. Node prints: false, then undefined.
type Entry = { valid?: boolean } & { valid?: boolean | string } & { name: string };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
probe({ name: "entry", valid: false });
probe({ name: "entry" });
