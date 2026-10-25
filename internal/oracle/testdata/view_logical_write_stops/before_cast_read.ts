// before_cast.ts's read as a plain read through the view, in a function lowered before the cast.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function bump(view: View): void {
	const count: number = view.count;
	console.log(`${count}`);
}
function initial(): string | number { return 'abc'; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
bump(base as View);
