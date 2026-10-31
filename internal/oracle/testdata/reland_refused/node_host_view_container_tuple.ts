import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const pair: [string, typeof performance] = ["a", performance];
const seen: readonly [string, View] = pair;
console.log(String(seen[1].eventLoopUtilization().idle >= 0));
