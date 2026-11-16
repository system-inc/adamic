// h4's twin: the literal the method is restored to holds x as a number at the same index, so the layout
// rule admits the restore and the call reads its x as Node does, printing 3.
interface Host { m(): string }
const a = { x: 1, m() { return String(this.x + 1); } };
const b = { x: 2, m() { return "b"; } };
const h: Host = a;
const h2: Host = b;
const original = h.m;
h2.m = original;
console.log(h2.m());
