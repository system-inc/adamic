import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const get = () => performance;
const seen: () => View = get;
console.log(String(seen().eventLoopUtilization().idle >= 0));
