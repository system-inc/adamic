const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
function rename(value: { name: string }): void {
    value.name = "Renamed";
}
const error = new Error("m");
console.log(String(probe(error, "name")));
rename(error);
console.log(String(probe(error, "name")));
