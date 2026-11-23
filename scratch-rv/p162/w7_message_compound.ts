const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const error = new Error();
console.log(String(probe(error, "message")));
error.message += "more";
console.log(String(probe(error, "message")));
console.log(error.message);
