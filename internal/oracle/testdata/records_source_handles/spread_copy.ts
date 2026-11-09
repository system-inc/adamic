// A spread copy of a record literal (#ny0c3gb, Language's forms): the copy is a literal of its own, seen
// as `{}` and then as an all-optional Read. Node prints text, a string in a number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
const m: Held = { a: "text" };
m["b"] = "more";
const copy = { ...m };
const wide: {} = copy;
const r: Read = wide;
console.log(String(r.a));
