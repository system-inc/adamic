interface Read { a?: number }
let seen: {} = {};
const m: { a: string } = (seen = { a: "text" });
const r: Read = seen;
console.log(r.a === undefined ? "none" : String(r.a + 1));
console.log(m.a);
