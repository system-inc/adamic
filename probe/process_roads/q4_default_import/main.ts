import nodeProcess from "node:process";
import { browserIsSet } from "./core.ts";
Object.assign(nodeProcess, { browser: true });
console.log(String(browserIsSet()));
