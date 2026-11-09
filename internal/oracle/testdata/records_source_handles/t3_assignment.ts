// Language's t3 on #ny0c3gb, its held value a record literal: Held assigned into a binding declared `{}`,
// then that binding seen as an all-optional Read. Node prints text, a string in a number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
const m: Held = { a: "text" };
m["b"] = "more";
let seen: {} = {};
seen = m;
const r: Read = seen;
console.log(String(r.a));
