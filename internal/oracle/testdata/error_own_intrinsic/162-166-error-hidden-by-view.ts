const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }): boolean {
    return own.call(value, "message");
}
console.log(String(probe(new Error("x"))));
