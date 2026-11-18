// #6c4swvt (c): an object literal handed to a parameter whose union holds one dictionary the program declares, and no
// fixed object it fits, is built as that dictionary, so the body's narrow to it reads a record. Built fixed, the narrow
// stopped on an honest value natively (Language's g5 p06).
interface MapLike<T> {
    [index: string]: T;
}
type Value = string | number | MapLike<string[]> | undefined;
function first(value: Value, key: string): string {
    if (typeof value === "object") {
        const list = value[key];
        return list ? String(list.length) + ":" + (list[0] ?? "none") : "missing";
    }
    return "scalar";
}
console.log(first({ k: ["a", "b"] }, "k"));
console.log(first({ k: ["a"] }, "z"));
console.log(first("x", "k"));
