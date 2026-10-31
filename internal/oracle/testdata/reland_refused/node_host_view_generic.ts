import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
function use<T extends View>(o: T): number { return o.eventLoopUtilization().idle; }
console.log(String(use(performance) >= 0));
