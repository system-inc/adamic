interface Read { a?: number }
function keep(x: object): object { return x; }
const held: { a: string } = { a: "text" };
const r: Read = keep(held);
console.log(r.a === undefined ? "none" : String(r.a + 1));
