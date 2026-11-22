// The publish gap on the reused path for certain. g02's n is a global, which the reuse plan never takes over
// (internal/native/reuse.go, owned). Here n is a local of run, read by the spread alone and dead after it, so the
// spread takes n's object over when it's unique and writes pos into it, a slot n's construction cast reserved absent.
interface Token { kind: number; pos: number }
function run(): void {
  const n = { kind: 1 } as Token;
  const m = { ...n, pos: 5 };
  console.log(String(m.pos)); // Node: 5
  console.log(Object.keys(m).join()); // Node: kind,pos
  console.log(String(JSON.stringify(m))); // Node: {"kind":1,"pos":5}
  console.log(String('pos' in m)); // Node: true
}
run();
