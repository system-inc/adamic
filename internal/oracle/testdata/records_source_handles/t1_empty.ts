// Language's t1 on #ny0c3gb, its held value a record literal: Held to `{}`, then `{}` to an all-optional
// Read, each hop legal where Held to Read is not. Node prints text, a string in a number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
const m: Held = { a: "text" };
m["b"] = "more";
const wide: {} = m;
const r: Read = wide;
console.log(String(r.a));
