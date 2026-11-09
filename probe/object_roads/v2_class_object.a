class Held { a = "text"; }
interface Read { a?: number }
function keep(x: object): object { return x; }
const r: Read = keep(new Held());
console.log(r.a === undefined ? "none" : String(r.a + 1));
