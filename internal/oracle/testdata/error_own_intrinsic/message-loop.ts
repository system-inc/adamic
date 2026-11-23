const own = Object.prototype.hasOwnProperty;
function test(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const error = new Error();
for (let i = 0; i < 3; i++) {
    console.log(String(test(error, "message")));
    error.message = String(i);
}
