// vacuous_downcast_refused/language.a as .ts: the vacuous cast stays a checked view, so the kind test
// passes and the read of count, which this Root never had, stops at the view's check where Node
// prints count NaN.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { count: number; }
const base: Root = { kind: 'Holder' };
console.log('viewing');
const viewed = base as View;
console.log(`count ${viewed.count + 1}`);
