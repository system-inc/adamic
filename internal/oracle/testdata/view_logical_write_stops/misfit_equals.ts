// misfit.ts's write as a plain =, which stops at the checked write.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number | string; }
function initial(): number { return 0; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
console.log('before');
viewed.count = 'seven';
console.log(`after ${raw.count}`);
