import { performance } from "node:perf_hooks";
const p: { eventLoopUtilization(): { readonly idle: number } } = performance;
console.log(String(p.eventLoopUtilization().idle >= 0));
