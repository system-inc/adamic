// A Record<string, number> handed on as object keeps its numbers unboxed, and a read as a Union would take a number's
// bits for a pointer. Node reads the number; both backends stop at the cast to MapLike<unknown> on the record's value
// kind, a run-time NotYet, after the line before it.
interface MapLike<T> {
    [index: string]: T;
}
function count(exports: object): unknown {
    return (exports as MapLike<unknown>)["a"];
}

const counts: Record<string, number> = {};
counts["a"] = 1;
console.log("before");
console.log(typeof count(counts));
console.log("after");
