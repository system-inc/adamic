// Representative member-count fixture. Not the full compiler interfaces.
interface Left { readonly text: string; readonly note?: string; }
interface Right { readonly child: {readonly label: string}; readonly note?: string; }
type Target = Left | Right;
interface Holder { readonly kind: 'Holder'; readonly value: Target; }
const raw = {kind:'Holder' as const, value:{text:'name'}};
const base: {readonly kind:'Holder'} = raw;
const view = base as Holder;
const value = view.value;
console.log(`${value === raw.value}`);
const raw2 = {kind:'Holder' as const,value:{child:{label:'name'}}};
const base2: {readonly kind:'Holder'}=raw2;
const view2=base2 as Holder;
const value2=view2.value;
console.log(`${value2 === raw2.value}`);
