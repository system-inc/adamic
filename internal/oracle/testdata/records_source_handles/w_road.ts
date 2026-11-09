// Language's W road on #ny0c3gb: a record held as Held, seen as W { b }, then as R { b; a? }. Every hop is legal
// in TypeScript and Held isn't assignable to R, and `{}` doesn't reach R (b is required), so the `{}` source
// doesn't keep this read checked: rulings.md:71 refuses it at `const r: R = w`, and this pins that stop.
interface Held { a: string; b: number; [key: string]: string | number }
interface W { b: number }
interface R { b: number; a?: number }
const m: Held = { a: "text", b: 1 };
m["c"] = "more";
const w: W = m;
const r: R = w;
console.log(String(r.a));
