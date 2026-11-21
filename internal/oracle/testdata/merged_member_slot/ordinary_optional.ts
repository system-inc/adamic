class Ordinary { a = 1; f?: string; }
const m = new Ordinary();
console.log(Object.keys(m).join(",") + " " + String("f" in m) + " " + JSON.stringify(m));
m.f = "m";
console.log(Object.keys(m).join(",") + " " + String("f" in m) + " " + JSON.stringify(m));
