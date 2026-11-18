type Unary = (a: number) => number;
function invoke(callback: (a: number, b: number) => number): number { return callback(2, 7); }
const opaque = (): number => 1;
console.log(String(opaque()));
function two(a: number, b?: number): number { return a + (b ?? 0); }
const host: { f: Unary } = { f: (a: number): number => a };
const writer = { set f(value: Unary) { host.f = value; } };
writer.f = two;
const bound = host.f.bind(host);
console.log(String(invoke(bound)));
