const own = Object.prototype.hasOwnProperty;
function test(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
console.log(String(test(new TypeError("x"), "stack")));
