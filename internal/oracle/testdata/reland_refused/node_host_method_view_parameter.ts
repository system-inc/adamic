import { statSync } from "node:fs";
function check(k: { readonly isFile: () => boolean }): string { return String(k.isFile()); }
console.log(check(statSync(".")));
