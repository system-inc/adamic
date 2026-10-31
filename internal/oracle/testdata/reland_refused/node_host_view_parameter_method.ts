import { performance } from "node:perf_hooks";
function idle(p: { eventLoopUtilization(): { readonly idle: number } }): boolean { return p.eventLoopUtilization().idle >= 0; }
console.log(String(idle(performance)));
