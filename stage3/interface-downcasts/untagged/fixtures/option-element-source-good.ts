// Representative member-count fixture. Not the full compiler interfaces.
interface StringOption { readonly kind: 1; readonly text: string; readonly note?: string; }
interface NumberOption { readonly kind: 2; readonly child: {readonly label: string}; readonly note?: string; }
interface BooleanOption { readonly kind: 3; readonly child: {readonly label: string}; readonly note?: string; }
interface CustomOption { readonly kind: 4; readonly child: {readonly label: string}; readonly note?: string; }
interface TsConfigOnlyOption { readonly flavor: 5; readonly child: {readonly label: string}; readonly note?: string; }
type Target = StringOption | NumberOption | BooleanOption | CustomOption | TsConfigOnlyOption;
interface Holder { readonly kind: 'Holder'; readonly value: Target; }
const raw = {kind:'Holder' as const, value:{kind:1, text:'name'}};
const base: {readonly kind:'Holder'} = raw;
const view = base as Holder;
const value = view.value;
console.log(`${value === raw.value}`);
const raw2 = {kind:'Holder' as const,value:{kind:2,child:{label:'name'}}};
const base2: {readonly kind:'Holder'}=raw2;
const view2=base2 as Holder;
const value2=view2.value;
console.log(`${value2 === raw2.value}`);
const raw3 = {kind:'Holder' as const,value:{kind:3,child:{label:'name'}}};
const base3: {readonly kind:'Holder'}=raw3;
const view3=base3 as Holder;
const value3=view3.value;
console.log(`${value3 === raw3.value}`);
const raw4 = {kind:'Holder' as const,value:{kind:4,child:{label:'name'}}};
const base4: {readonly kind:'Holder'}=raw4;
const view4=base4 as Holder;
const value4=view4.value;
console.log(`${value4 === raw4.value}`);
const raw5 = {kind:'Holder' as const,value:{flavor:5,child:{label:'name'}}};
const base5: {readonly kind:'Holder'}=raw5;
const view5=base5 as Holder;
const value5=view5.value;
console.log(`${value5 === raw5.value}`);
