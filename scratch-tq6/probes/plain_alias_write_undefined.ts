// A boolean the object's own type says is always present, written undefined through a wider alias whose
// field is optional and may be undefined (exactOptionalPropertyTypes allows the write there), then read
// through the object's own type. Node prints: true, undefined, undefined.
type Own = { valid: boolean; name: string };
type Wide = { valid?: boolean | undefined; name: string };
const own: Own = { valid: true, name: "own" };
const wide: Wide = own;
console.log(String(own.valid));
wide.valid = undefined;
console.log(String(wide.valid));
console.log(String(own.valid));
