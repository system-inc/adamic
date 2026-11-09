interface Read { a?: string }
function keep(x: object): object { return x; }
const held: { a: number } = { a: 5 };
const r: Read = keep(held);
console.log(r.a === undefined ? "none" : r.a.toUpperCase());
