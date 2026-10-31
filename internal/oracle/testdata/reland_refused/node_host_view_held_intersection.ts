import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const p: typeof performance & object = performance;
const seen: View = p;
console.log(String(seen.eventLoopUtilization().idle >= 0));
