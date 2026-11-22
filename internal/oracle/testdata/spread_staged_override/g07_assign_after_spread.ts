// The publish gap's Object.assign road. The target is a spread's copy of a construction cast, pos still reserved
// absent, and Object.assign writes pos from an exact literal source. The runtime's assign writes through
// adamic_object_write_field, which publishes the slot and makes it ready (class_static.c), so this road needs no
// publish of the spread's own; it's pinned so it stays so.
interface Token { kind: number; pos: number }
const n = { kind: 1 } as Token;
const m = Object.assign({ ...n }, { pos: 5 });
console.log(String(m.pos)); // Node: 5
console.log(Object.keys(m).join()); // Node: kind,pos
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
console.log(String('pos' in m)); // Node: true
