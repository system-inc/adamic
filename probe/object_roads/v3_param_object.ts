interface Read { a?: number }
function readA(o: object): string { const r: Read = o; return r.a === undefined ? "none" : String(r.a + 1); }
const held: { a: string } = { a: "text" };
console.log(readA(held));
