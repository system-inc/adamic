import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const box = { get: () => performance };
const seen: { readonly get: () => View } = box;
console.log(String(seen.get().eventLoopUtilization().idle >= 0));
