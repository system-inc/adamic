// A record held inside a generic as Record<string, T | undefined>, the shape of tsc's core.ts extend (#ny0c3gb).
// The walk relates declarations, so it meets T uninstantiated and relates it as unknown (Adamic's ruling,
// 16:12:33Z): the record is held as Record<string, unknown>. Shape, a type literal, has an implicit index
// signature, so it relates to that, and its read stays checked; Read is all-optional, so `{}` reaches it,
// and its read stays checked too. Related as a rigid T (record-source-handles.mutants.md), Shape relates to
// the record in neither direction, its read goes unchecked, and this goes red.
interface Read { a?: number }
type Shape = { label: string; size?: number };
function extend<T>(keys: string[], value: T): number {
	const result: Record<string, T | undefined> = {};
	for (const key of keys) {
		result[key] = value;
	}
	for (const key of Object.keys(result)) console.log(key);
	return Object.keys(result).length;
}
console.log(String(extend(["alpha", "beta"], 1)));
function first(r: Read): string {
	return String(r.a);
}
function size(shape: Shape): string {
	return String(shape.size);
}
console.log(first({ a: 1 }));
console.log(first({}));
console.log(size({ label: "a", size: 2 }));
console.log(size({ label: "b" }));
