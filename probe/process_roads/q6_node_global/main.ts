import { browserIsSet } from "./core.ts";
Object.assign(process, { browser: true });
console.log(String(browserIsSet()));
