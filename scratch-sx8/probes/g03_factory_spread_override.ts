// The publish gap through a factory. make's one return is a construction cast, so its literal reserves pos and
// text as absent slots (internal/lower/construction.go:218 reserveFactoryFields, :245). The spread overrides both,
// and kind too, which was present: Node keeps kind first and adds pos and text after it in the literal's order.
interface Token { kind: number; pos: number; text: string }
function make(): Token {
  return { kind: 1 } as Token;
}
const n = make();
const m = { ...n, pos: 5, kind: 2, text: 'x' };
console.log(m.pos); // Node: 5
console.log(m.text); // Node: x
console.log(Object.keys(m).join()); // Node: kind,pos,text
console.log(JSON.stringify(m)); // Node: {"kind":2,"pos":5,"text":"x"}
console.log('text' in m); // Node: true
