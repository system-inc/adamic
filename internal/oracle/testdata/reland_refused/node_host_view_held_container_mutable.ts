import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const box = { p: performance };
const seen: { p: View } = box;
console.log(String(seen.p.eventLoopUtilization().idle >= 0));
