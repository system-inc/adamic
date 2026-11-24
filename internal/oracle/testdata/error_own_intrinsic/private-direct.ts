import { rmSync } from "node:fs";
try {
    rmSync(".");
} catch (error) {
    if (error instanceof Error) console.log(String(error.hasOwnProperty("#secret@1")));
}
