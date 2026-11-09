interface Held { a: string }
interface Read { a?: number }
const held: Held = { a: "text" };
const o: object = held;
const r: Read = o;
console.log(r.a === undefined ? "none" : String(r.a + 1));
