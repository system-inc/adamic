import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const xs = [performance];
const vs: readonly View[] = xs;
const v = vs[0];
console.log(v === undefined ? "none" : String(v.eventLoopUtilization().idle >= 0));
