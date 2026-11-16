// Compiler's probe: a method that reads this.x, saved and restored to an object of another literal whose x
// sits at the same index but holds a string. A call through it would read the string's slot as a number
// natively, where Node prints s1.
interface Host { m(): string }
const a = { x: 1, m() { return String(this.x + 1); } };
const b = { x: "s", m() { return "b"; } };
const h: Host = a;
const h2: Host = b;
const original = h.m;
h2.m = original;
console.log(h2.m());
