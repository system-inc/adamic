import { performance } from "node:perf_hooks";
function view(): { eventLoopUtilization(): { readonly idle: number } } { return performance; }
console.log(String(view().eventLoopUtilization().idle >= 0));
