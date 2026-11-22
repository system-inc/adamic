// The same write through an intersection, which stops today: the object's own boolean, written undefined
// through the intersection's optional field, then read through its own type.
// Node prints: true, undefined, undefined.
type Own = { valid: boolean; name: string };
type Entry = { valid?: boolean | undefined } & { name: string };
const own: Own = { valid: true, name: "own" };
const entry: Entry = own;
console.log(String(own.valid));
entry.valid = undefined;
console.log(String(entry.valid));
console.log(String(own.valid));
