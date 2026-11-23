const own = Object.prototype.hasOwnProperty;
function test(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
const objects = new Map<string, { message: string }>();
objects.set("plain", {message: "x"});
objects.set("error", new Error());
for (const value of objects.values()) {
    console.log(String(test(value, "stack")));
}
