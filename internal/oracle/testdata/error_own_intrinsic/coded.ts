import { Buffer } from "node:buffer";
const own = Object.prototype.hasOwnProperty;
function test(value: { message: string }, key: string): boolean {
    return own.call(value, key);
}
try {
    Buffer.allocUnsafe(-1);
} catch (error) {
    if (error instanceof Error) {
        console.log(String(test(error, "name")));
        console.log(String(test(error, "message")));
    }
}
