// A Record<string, boolean> handed on as object keeps its booleans unboxed, as a record of numbers does. Node reads the
// boolean; both backends stop at the cast to MapLike<unknown> on the record's value kind, a run-time NotYet, after the
// line before it.
interface MapLike<T> {
    [index: string]: T;
}
function flag(exports: object): unknown {
    return (exports as MapLike<unknown>)["strict"];
}

const flags: Record<string, boolean> = {};
flags["strict"] = true;
console.log("before");
console.log(typeof flag(flags));
console.log("after");
