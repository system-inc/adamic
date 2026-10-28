// Spec: a downcast whose tag test is vacuous
// Expect: lowers
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
function initial(): string | number { return 'abc'; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
const count: number = viewed.count;
console.log(`${count}`);
