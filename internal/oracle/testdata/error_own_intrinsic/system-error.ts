import { rmSync, readFileSync } from "node:fs";
const own = Object.prototype.hasOwnProperty;
function probe(value: { message: string }, key: string): void {
    console.log(String(own.call(value, key)));
}
try {
    rmSync(".");
} catch (error) {
    if (error instanceof Error) {
        probe(error, "name");
        probe(error, "code");
        probe(error, "message");
        probe(error, "stack");
    }
}
try {
    readFileSync("a\0b", "utf8");
} catch (error) {
    if (error instanceof Error) {
        probe(error, "name");
        probe(error, "code");
    }
}
