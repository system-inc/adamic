import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const p = performance satisfies View;
const seen: View = p;
console.log(String(seen.eventLoopUtilization().idle >= 0));
