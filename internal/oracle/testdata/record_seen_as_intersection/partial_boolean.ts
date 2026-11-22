// partial_number.ts's road with an optional boolean field (#tq6yzqy): every member of the intersection
// holds valid?: boolean, so it's held as a fixed object (intersectionMembersHoldMaybeBoolean), and a
// finite Partial<Record<"valid", boolean>> declaring valid gives the optional-widening rule nothing
// absent. The argument stops at recordStorageView: a record and a fixed object have different storage
// (#mft3zc6). Node prints: false, undefined, true.
type Entry = { valid?: boolean } & { valid?: boolean };
function probe(value: Entry): void {
	console.log(String(value.valid));
}
const set: Partial<Record<"valid", boolean>> = { valid: false };
const unset: Partial<Record<"valid", boolean>> = {};
probe(set);
probe(unset);
const direct: Entry = { valid: true };
probe(direct);
