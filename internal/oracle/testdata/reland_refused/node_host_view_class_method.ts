import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
class Maker { make(): typeof performance { return performance; } }
const seen: { make(): View } = new Maker();
console.log(String(seen.make().eventLoopUtilization().idle >= 0));
