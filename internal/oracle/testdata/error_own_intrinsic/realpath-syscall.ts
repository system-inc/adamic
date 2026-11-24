import { realpathSync, mkdtempSync, rmSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
const directory = mkdtempSync("/tmp/adamic-own-key-");
try {
    realpathSync(directory + "/missing");
} catch (error) {
    rmSync(directory, { recursive: true });
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log("before");
        console.log(String(own.call(value, "syscall")));
    }
}
