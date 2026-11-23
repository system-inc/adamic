const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const error = new Error("m");
console.log(String(probe(error, "name")));
error["name"] = "Renamed";
console.log(String(probe(error, "name")));
