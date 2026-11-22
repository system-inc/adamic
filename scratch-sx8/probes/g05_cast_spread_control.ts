// Control for g01: pos is written before the spread, so the source already holds it present and the override
// keeps its place. This one should agree with or without a publish after the spread's write.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
n.pos = 3;
const m = { ...n, pos: 5 };
console.log(String(m.pos)); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
console.log(Object.keys(n).join()); // Node: kind,pos
