class Base { a = 1; }
interface Base { f?: string; }
class Derived extends Base { b = 2; }
const m = new Derived();
console.log(String(m.f) + " " + Object.keys(m).join(","));
m.f = "m";
console.log(String(m.f) + " " + Object.keys(m).join(","));
