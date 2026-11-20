class Merged { constructor(public a: number) {} }
interface Merged { f?: string; }
const m = new Merged(1);
console.log(String(m.f) + " " + Object.keys(m).join(","));
m.f = "m";
console.log(String(m.f) + " " + Object.keys(m).join(","));
