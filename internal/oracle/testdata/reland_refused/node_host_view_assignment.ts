import { performance } from "node:perf_hooks";
let p: { eventLoopUtilization(): { readonly idle: number } } | undefined = undefined;
p = performance;
console.log(String(p.eventLoopUtilization().idle >= 0));
