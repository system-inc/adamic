import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
function same<T extends View>(o: T): T { return o; }
console.log(String(same(performance).eventLoopUtilization().idle >= 0));
