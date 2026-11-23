const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const error = new Error("m");
const copy: { message: string; name: string } = { ...error, name: "copied" };
console.log(String(probe(copy, "message")));
console.log(String(probe(error, "name")));
