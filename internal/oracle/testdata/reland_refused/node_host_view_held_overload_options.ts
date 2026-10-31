import * as fs from "node:fs";
const real: (path: string, options?: fs.EncodingOption) => string = fs.realpathSync;
console.log(String(real(".", "utf8").length > 0));
