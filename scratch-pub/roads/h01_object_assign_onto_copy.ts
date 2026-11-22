// Compiler's road (a) for #sx8pbw1's publish gap: Object.assign writing a reserved field onto a copy of a staged
// object. Node prints: 5, kind,pos, {"kind":1,"pos":5}, true.
type Token = { kind: number; pos: number };
const n = { kind: 1 } as Token;
const copy = { ...n };
Object.assign(copy, { pos: 5 });
console.log(String(copy.pos));
console.log(Object.keys(copy).join());
console.log(String(JSON.stringify(copy)));
console.log(String('pos' in copy));
