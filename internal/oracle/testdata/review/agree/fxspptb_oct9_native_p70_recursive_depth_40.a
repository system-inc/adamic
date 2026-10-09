interface Root { readonly kind: 'Receiver'; }
interface Link { readonly label: string; readonly next?: Target; }
interface Count { readonly count: number; readonly next?: Target; }
type Target = Link | Count;
interface Receiver extends Root { readonly value: Target; }
let chain: Target = { label: 'leaf' };
for (let index = 0; index < 40; index++) {
	chain = index % 2 === 0 ? { count: index, next: chain } : { label: `n${index}`, next: chain };
}
const raw = { kind: 'Receiver' as const, value: chain };
const base: Root = raw;
const view = base as Receiver;
const value = view.value;
console.log(`${value !== undefined}`);
