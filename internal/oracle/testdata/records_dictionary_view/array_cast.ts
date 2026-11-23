// Compiler's pin (#2v3xtgm): an array handed on as object, cast to MapLike<unknown> as tsc casts a package.json's
// exports (moduleSpecifiers.ts:1109). Node reads its "0" by key. The view isn't built for a value that isn't a record,
// so both backends stop at the cast, a run-time NotYet, after the line before it.
interface MapLike<T> {
    [index: string]: T;
}
function first(exports: object): unknown {
    return (exports as MapLike<unknown>)["0"];
}

const files: string[] = ["a.js", "b.js"];
console.log("before");
console.log(typeof first(files));
console.log("after");
