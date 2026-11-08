import { browserIsSet } from "./core.ts";
Reflect.set(globalThis.process, "browser", true);
console.log(String(browserIsSet()));
