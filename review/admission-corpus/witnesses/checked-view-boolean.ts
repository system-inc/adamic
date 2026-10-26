interface Left { readonly flag: boolean; }
interface Right { readonly text: string; }
type Target = Left | Right;
interface Holder { readonly kind: 'Holder'; readonly value: Target; }
const raw = {kind:'Holder' as const, value:{flag:true}};
const base: {readonly kind:'Holder'} = raw;
const view = base as Holder;
const value = view.value;
console.log(`${value === raw.value}`);
