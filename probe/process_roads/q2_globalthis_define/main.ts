import { browserIsSet } from "./core.ts";
Object.defineProperty(globalThis.process, "browser", { value: true });
console.log(String(browserIsSet()));
