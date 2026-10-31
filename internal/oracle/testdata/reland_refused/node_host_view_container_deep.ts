import { performance } from "node:perf_hooks";
type View = { eventLoopUtilization(): { readonly idle: number } };
const box = { a: { a: { a: { a: { a: { a: { a: { a: { a: { a: { p: performance } } } } } } } } } } };
const seen: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly a: { readonly p: View } } } } } } } } } } } = box;
console.log(String(seen.a.a.a.a.a.a.a.a.a.a.p.eventLoopUtilization().idle >= 0));
