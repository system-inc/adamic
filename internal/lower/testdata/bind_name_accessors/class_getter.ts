type Unary = (a: number) => number;
function invoke(callback: (a: number, b: number) => number): number { return callback(2, 7); }
const opaque = (): number => 1;
console.log(String(opaque()));
class Host {
  get f(): Unary { return (a: number, b?: number): number => a + (b ?? 0); }
}
const host: { f: Unary } = new Host();
const bound = host.f.bind(host);
console.log(String(invoke(bound)));
