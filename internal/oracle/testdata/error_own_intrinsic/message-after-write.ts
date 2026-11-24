const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }): boolean {
    return own.call(value, "message");
}
const error = new Error();
console.log(String(probe(error)));
error.message = "later";
console.log(String(probe(error)));
