interface X { a: number; f?: string }
function put(x: X, value: string): void { x.f = value; }
function read(x: X): string { return String(x.f) + "/" + typeof x.f; }
class Merged { a = 1; }
interface Merged { f?: string; }
const m = new Merged();
put(m, "m");
console.log(read(m) + " " + Object.keys(m).join(","));
