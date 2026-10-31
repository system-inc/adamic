import { performance } from "node:perf_hooks";
const p: { toJSON(): { readonly timeOrigin: string } } = performance;
console.log(String(p.toJSON().timeOrigin.length));
