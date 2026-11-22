// The publish gap through a nested spread. The inner spread copies the construction cast, pos still reserved
// absent, and the outer spread copies that copy and writes pos into it, in objectLiteral's spread arm, the
// site g01 takes.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
const m = { ...{ ...n }, pos: 5 };
console.log(String(m.pos)); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
console.log(String('pos' in m)); // Node: true
