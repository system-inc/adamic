import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const p = performance satisfies View;
console.log(String(p.eventLoopUtilization().idle >= 0));
