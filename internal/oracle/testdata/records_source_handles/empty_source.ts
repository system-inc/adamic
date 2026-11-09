// The implicit `{}` source (#ny0c3gb, Language's t1): Read declares only an optional a, so `{}` is
// assignable to it, and a record seen as `{}` can be seen as Read, whatever its handle. Held, the record's
// handle, relates to Read in neither direction, so only that `{}` question keeps Read's read checked.
// Dropping it (record-source-handles.mutants.md) leaves the read unchecked, and this goes red.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
const m: Held = { a: "text" };
m["b"] = "more";
for (const key of Object.keys(m)) console.log(key);
function first(r: Read): string {
	return String(r.a);
}
console.log(first({ a: 1 }));
console.log(first({}));
