// The publish gap, reused path. As g01, but nothing reads n after the spread, so the reuse plan may write pos
// into n itself (internal/native/reuse.go:630), again with no adamic_object_publish.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
const m = { ...n, pos: 5 };
console.log(m.pos); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(JSON.stringify(m)); // Node: {"kind":1,"pos":5}
console.log('pos' in m); // Node: true
