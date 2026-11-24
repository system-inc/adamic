import { rmSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
try {
    rmSync(".");
} catch (error) {
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log("before");
        console.log(String(own.call(value, "info")));
    }
}
