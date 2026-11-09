// A record literal as a destructuring default (#ny0c3gb, Language's forms): the binding element holds it
// as `{}`, and that binding is seen as an all-optional Read. Node prints text, a string in a number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
interface Box { held?: {} }
const m: Held = { a: "text" };
m["b"] = "more";
const box: Box = {};
const { held = m } = box;
const r: Read = held;
console.log(String(r.a));
