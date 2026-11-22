// The publish gap, as g01 but with nothing reading n after the spread. n is a global, which the reuse plan never
// takes over (internal/native/reuse.go, owned), so this is the copy path too; g06 is the reused one.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
const m = { ...n, pos: 5 };
console.log(String(m.pos)); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
console.log(String('pos' in m)); // Node: true
