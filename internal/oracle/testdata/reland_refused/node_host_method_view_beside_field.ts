import { statSync } from "node:fs";
const e: { readonly size: number; isDirectory(): boolean } = statSync(".");
console.log(String(e.isDirectory() && e.size >= 0));
