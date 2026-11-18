// #6c4swvt (c): an object literal written by a computed key into a record whose values are a union holding one
// dictionary the program declares, read back and narrowed to it. Built as that dictionary, the read agrees with Node.
interface MapLike<T> {
    [index: string]: T;
}
type Value = string | number | string[] | MapLike<string[]> | undefined;
interface Options {
    [option: string]: Value;
}
function key(): string {
    return "paths";
}
const options: Options = {};
options[key()] = { k: ["a", "b"] };
const value = options["paths"];
if (typeof value === "object" && !Array.isArray(value)) {
    const list = value["k"];
    console.log(list ? String(list.length) : "missing");
}
