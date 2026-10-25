// and.ts's read as a plain read through the view, which stops at the checked read.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 'abc'; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
const count: number = viewed.count;
console.log(`${count}`);
