import { readFileSync, mkdtempSync, rmSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
const directory = mkdtempSync("/tmp/adamic-own-key-");
try {
    readFileSync(directory + "/missing", "utf8");
} catch (error) {
    rmSync(directory, { recursive: true });
    if (error instanceof Error) {
        const value: { message: string } = error;
        console.log(String(own.call(value, "code")));
    }
}
