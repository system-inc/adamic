// Representative member-count fixture. Not the full compiler interfaces.
interface Left { readonly text: string; readonly payload: {readonly label:string}; readonly note?: string; }
interface Right { readonly child: {readonly label: string}; readonly payload: {readonly label:string}; readonly note?: string; }
type Target = Left | Right;
interface Holder { readonly kind: 'Holder'; readonly value: Target; }
const raw = {kind:'Holder' as const, value:{text:'name',payload:{label:true}}};
const base: {readonly kind:'Holder'} = raw;
const view = base as Holder;
const value = view.value;
console.log(`${value.payload.label}`);
