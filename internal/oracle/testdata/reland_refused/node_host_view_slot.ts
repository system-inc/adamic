import { performance } from "node:perf_hooks";
const box: { readonly p: { eventLoopUtilization(): { readonly idle: number } } } = { p: performance };
console.log(String(box.p.eventLoopUtilization().idle >= 0));
