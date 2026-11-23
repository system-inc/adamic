import { Buffer } from "node:buffer";
import { rmSync, readFileSync, mkdtempSync } from "node:fs";
try {
    rmSync(".");
} catch (error) {
    if (error instanceof Error) {
        console.log(String(error.hasOwnProperty("name")));
        console.log(String(error.hasOwnProperty("code")));
    }
}
const directory = mkdtempSync("/tmp/adamic-own-key-");
try {
    readFileSync(directory + "/missing", "utf8");
} catch (error) {
    rmSync(directory, { recursive: true });
    if (error instanceof Error) console.log(String(error.hasOwnProperty("code")));
}
try {
    Buffer.allocUnsafe(-1);
} catch (error) {
    if (error instanceof Error) console.log(String(error.hasOwnProperty("code")));
}
