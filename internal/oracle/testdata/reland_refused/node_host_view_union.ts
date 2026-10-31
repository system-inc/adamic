import { performance } from "node:perf_hooks";
const p: number | { eventLoopUtilization(): { readonly idle: number } } = performance;
if (typeof p !== "number") console.log(String(p.eventLoopUtilization().idle >= 0)); else console.log("none");
