const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const error = new Error("m");
console.log(String(probe(error, "name")));
({ name: error.name } = { name: "Renamed" });
console.log(String(probe(error, "name")));
