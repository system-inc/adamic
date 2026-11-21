class Merged { a = 1; }
interface Merged { f?: string; }
const m = new Merged();
console.log(String(m.f) + " " + Object.keys(m).join(","));
