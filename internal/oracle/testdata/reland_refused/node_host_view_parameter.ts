import { performance } from "node:perf_hooks";
function origin(p: { toJSON(): { readonly timeOrigin: string } }): number { return p.toJSON().timeOrigin.length; }
console.log(String(origin(performance)));
