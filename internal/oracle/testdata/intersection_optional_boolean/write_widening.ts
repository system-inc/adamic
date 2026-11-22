// The scout's write through the intersection (#tq6yzqy): an object whose own type holds valid as a plain
// boolean, seen as the intersection, whose optional field can be written undefined, stops at the widening rule
// before any intersection is lowered. Node prints: true, undefined, undefined.
type Own = { valid: boolean; name: string };
type Entry = { valid?: boolean | undefined } & { name: string };
const own: Own = { valid: true, name: "own" };
const entry: Entry = own;
console.log(String(own.valid));
entry.valid = undefined;
console.log(String(entry.valid));
console.log(String(own.valid));
