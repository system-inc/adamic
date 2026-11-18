type Unary = (a: number) => number;
function invoke(callback: (a: number, b: number) => number): number { return callback(2, 7); }
const opaque = (): number => 1;
console.log(String(opaque()));
const host: { f: Unary } = {
  get f(): Unary { return (a: number, b?: number): number => a + (b ?? 0); }
};
const bound = host.f.bind(host);
console.log(String(invoke(bound)));
