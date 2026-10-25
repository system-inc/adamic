// The assignment's function is lowered before the cast that makes count checked, and its read and write
// are checked all the same.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function bump(view: View): number {
	return view.count &&= 7;
}
function initial(): string | number { return 5; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const result = bump(base as View);
const read: string | number = raw.count;
console.log(typeof read === 'number' ? `number ${read + 1} ${result}` : `string ${read}`);
