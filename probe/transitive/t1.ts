interface Held { a: string }
interface Read { a?: number }
const m: Held = { a: "text" };
const wide: {} = m;
const r: Read = wide;
console.log(r.a === undefined ? "none" : String(r.a + 1));
