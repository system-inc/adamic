// Compiler's road (b): a nested spread of a staged object, then the reserved field written explicitly.
// Node prints: 5, kind,pos, {"kind":1,"pos":5}, true.
type Token = { kind: number; pos: number };
const n = { kind: 1 } as Token;
const m = { ...{ ...n }, pos: 5 };
console.log(String(m.pos));
console.log(Object.keys(m).join());
console.log(String(JSON.stringify(m)));
console.log(String('pos' in m));
