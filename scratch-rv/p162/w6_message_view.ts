const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
function say(value: { message: string }): void {
    value.message = "said";
}
const error = new Error();
console.log(String(probe(error, "message")));
say(error);
console.log(String(probe(error, "message")));
