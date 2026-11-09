import { statSync } from "node:fs";
const k: { readonly isFile: () => boolean } = statSync(".");
console.log(String(k.isFile()));
