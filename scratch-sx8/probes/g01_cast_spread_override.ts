// The publish gap, copy path. `{ kind: 1 } as Token` is a construction cast (internal/lower/construction.go:12):
// every field of Token is a required scalar and the literal leaves pos out, so constructLiteral (:85) reserves pos
// as an absent slot and the object carries write_order. The spread copies that, pos still absent, and the explicit
// pos is then written into the slot (internal/native/emit_objects.go:150) with no adamic_object_publish.
// n is read again after the spread, so the spread copies rather than reusing n.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
const m = { ...n, pos: 5 };
console.log(String(m.pos)); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
console.log(String('pos' in m)); // Node: true
console.log(Object.keys(n).join()); // Node: kind
