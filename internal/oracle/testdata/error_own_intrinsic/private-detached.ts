import { rmSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
try {
    rmSync(".");
} catch (error) {
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log(String(own.call(value, "#secret@1")));
    }
}
