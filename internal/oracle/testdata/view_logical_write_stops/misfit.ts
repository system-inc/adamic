// ||= of a string through a view whose count is number | string, over an object that only ever held a
// number in count: the write stops, as misfit_equals.ts's plain = does, where Node prints "after seven".
// A ruled negative witness (#995z2fa): misfit_repaired.ts declares count string | number, as the view does,
// and agrees with Node.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number | string; }
function initial(): number { return 0; }
const raw = { kind: 'Holder' as const, count: initial() };
const base: Root = raw;
const viewed = base as View;
console.log('before');
viewed.count ||= 'seven';
console.log(`after ${raw.count}`);
