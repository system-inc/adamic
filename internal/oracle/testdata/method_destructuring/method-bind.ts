interface G { get(): number }
class C { v = 3; get(): number { return this.v; } }
const c = new C();
const g: G = { get: c.get.bind(c) };
const { get } = g;
console.log(String(get()));
