import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
function first<T extends View>(os: readonly T[]): string { const v = os[0]; console.log(v === undefined ? "none" : String(v.eventLoopUtilization().idle >= 0)); return ""; }
first([performance]);
