// Compiler's probe through .call: a method that reads this.x, saved off a and called with b as its this,
// where b's x sits at the same index but holds a string. Natively the call would read the string's slot as
// a number, where Node prints s1, so it stops at the .call: b isn't the literal the method was read off.
interface Host { m(): string }
const a = { x: 1, m() { return String(this.x + 1); } };
const b = { x: "s", m() { return "b"; } };
const h: Host = a;
const original = h.m;
console.log(original.call(b));
