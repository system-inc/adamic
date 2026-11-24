import { Buffer } from "node:buffer";
const own = Object.prototype.hasOwnProperty;
try {
    Buffer.allocUnsafe(-1);
} catch (error) {
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log(String(own.call(value, "code")));
    }
}
