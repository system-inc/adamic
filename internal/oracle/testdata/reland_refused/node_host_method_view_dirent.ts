import { readdirSync } from "node:fs";
for (const d of readdirSync(".", { withFileTypes: true })) { const k: { isFile(): boolean } = d; console.log(typeof k.isFile()); break; }
