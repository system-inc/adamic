interface X { a: number; f?: string }
function put(x: X, value: string): void { x.f = value; }
class Merged { a = 1; }
interface Merged { f?: string; }
const m = new Merged();
const before = { ...m };
console.log(String("f" in m) + " " + JSON.stringify(m) + " " + Object.keys(before).join(",") + " " + String(before.f));
put(m, "m");
const after = { ...m };
console.log(String("f" in m) + " " + JSON.stringify(m) + " " + Object.keys(after).join(",") + " " + String(after.f));
