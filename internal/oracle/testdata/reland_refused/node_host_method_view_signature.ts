import { statSync } from "node:fs";
const k: { isFile(): boolean } = statSync(".");
console.log(String(k.isFile()));
