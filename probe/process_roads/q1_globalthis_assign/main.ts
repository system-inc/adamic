import { browserIsSet } from "./core.ts";
Object.assign(globalThis.process, { browser: true });
console.log(String(browserIsSet()));
