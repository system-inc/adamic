import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const m = new Map<string, typeof performance>([["k", performance]]);
const seen: ReadonlyMap<string, View> = m;
const v = seen.get("k");
console.log(v === undefined ? "none" : String(v.eventLoopUtilization().idle >= 0));
