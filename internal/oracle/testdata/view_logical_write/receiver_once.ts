// The receiver is evaluated once, for the checked read and the checked write both, and the right side
// only when the write is taken.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 0; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
let calls = 0;
function holder(): View {
	calls++;
	return base as View;
}
function seven(): number {
	console.log(`right after ${calls} call`);
	return 7;
}
holder().count ||= seven();
holder().count ||= seven();
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1} after ${calls} calls` : `string ${read}`);
