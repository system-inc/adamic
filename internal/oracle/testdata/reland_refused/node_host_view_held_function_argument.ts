import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
function run(make: () => View): number { return make().eventLoopUtilization().idle; }
console.log(String(run(() => performance) >= 0));
