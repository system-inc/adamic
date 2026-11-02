// A real tuple read through a checked view that says a tuple agrees with Node on both backends: tuples keep their
// mark when array literals lose it (#nazvcep).
interface Pairish { readonly kind: 'p' | 'q'; }
interface WithPair extends Pairish { readonly kind: 'p'; readonly pair: readonly [number, number]; }
const raw = { kind: 'p' as const, pair: [1, 2] as const };
const p: Pairish = raw;
console.log('before');
console.log(String((p as WithPair).pair[0]));
