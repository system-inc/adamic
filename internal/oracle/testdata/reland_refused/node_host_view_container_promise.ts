import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
async function main(): Promise<void> { const pr: Promise<View> = Promise.resolve(performance); const p = await pr; console.log(String(p.eventLoopUtilization().idle >= 0)); }
await main();
