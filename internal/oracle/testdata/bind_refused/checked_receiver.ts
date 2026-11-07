// Node binds the method and prints 7. Main stops this cast before reaching bind.
interface H { m: () => number }
const raw: unknown = { m: (): number => 7 };
const host = raw as H;
const bound = host.m.bind(host);
console.log(String(bound()));
