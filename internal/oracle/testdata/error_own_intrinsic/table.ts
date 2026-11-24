const own = Object.prototype.hasOwnProperty;
function test(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const empty = new Error();
const full = new Error("x");
console.log(String(test(empty, "message")));
console.log(String(test(empty, "stack")));
console.log(String(test(full, "message")));
console.log(String(test(full, "name")));
full.name = "Custom";
console.log(String(test(full, "name")));
console.log(String(test({message: "plain"}, "stack")));
console.log(String(test({message: "plain"}, "message")));
