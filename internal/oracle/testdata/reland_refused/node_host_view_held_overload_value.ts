import * as fs from "node:fs";
import { tmpdir } from "node:os";
const mk: (prefix: string) => string = fs.mkdtempSync;
const made = mk(tmpdir() + "/adamic-");
console.log(String(made.length > 0));
fs.rmSync(made, { recursive: true });
