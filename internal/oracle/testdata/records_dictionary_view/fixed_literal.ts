// An object literal handed on as object is a fixed object, its fields in slots, never a record's table. Node reads its
// name by key; both backends stop at the cast to MapLike<unknown>, a run-time NotYet, after the line before it.
interface MapLike<T> {
    [index: string]: T;
}
function packageName(exports: object): unknown {
    return (exports as MapLike<unknown>)["name"];
}

console.log("before");
console.log(typeof packageName({ name: "literal" }));
console.log("after");
