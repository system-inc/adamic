import { statSync } from "node:fs";
const xs = [statSync(".")];
const all: ReadonlyArray<{ readonly isFile: () => boolean }> = xs;
const k = all[0];
console.log(k === undefined ? "none" : String(k.isFile()));
