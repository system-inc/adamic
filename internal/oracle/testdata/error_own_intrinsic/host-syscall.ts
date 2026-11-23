import { readFileSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
try {
    readFileSync("/tmp/lane-round5-system-target/missing", "utf8");
} catch (error) {
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log("before");
        console.log(String(own.call(value, "syscall")));
    }
}
