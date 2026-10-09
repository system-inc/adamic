// Representative member-count fixture. Not the full compiler interfaces.
interface Identifier { readonly kind: 1; readonly text: string; readonly note?: string; }
interface ObjectBindingPattern { readonly kind: 2; readonly child: {readonly label: string}; readonly note?: string; }
interface ArrayBindingPattern { readonly kind: 3; readonly child: {readonly label: string}; readonly note?: string; }
interface PrivateIdentifier { readonly kind: 4; readonly child: {readonly label: string}; readonly note?: string; }
interface StringLiteral { readonly kind: 5; readonly child: {readonly label: string}; readonly note?: string; }
interface NumericLiteral { readonly kind: 6; readonly child: {readonly label: string}; readonly note?: string; }
interface BigIntLiteral { readonly kind: 7; readonly child: {readonly label: string}; readonly note?: string; }
interface ElementAccessExpression { readonly kind: 8; readonly child: {readonly label: string}; readonly note?: string; }
interface ComputedPropertyName { readonly kind: 9; readonly child: {readonly label: string}; readonly note?: string; }
interface JsxNamespacedName { readonly kind: 10; readonly child: {readonly label: string}; readonly note?: string; }
interface NoSubstitutionTemplateLiteral { readonly kind: 11; readonly child: {readonly label: string}; readonly note?: string; }
interface PropertyAccessEntityNameExpression { readonly flavor: 12; readonly child: {readonly label: string}; readonly note?: string; }
type Target = Identifier | ObjectBindingPattern | ArrayBindingPattern | PrivateIdentifier | StringLiteral | NumericLiteral | BigIntLiteral | ElementAccessExpression | ComputedPropertyName | JsxNamespacedName | NoSubstitutionTemplateLiteral | PropertyAccessEntityNameExpression;
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
const raw5 = {kind:'Holder' as const,value:{kind:5,child:{label:'name'}}};
const base5: {readonly kind:'Holder'}=raw5;
const view5=base5 as Holder;
const value5=view5.value;
console.log(`${value5 === raw5.value}`);
const raw6 = {kind:'Holder' as const,value:{kind:6,child:{label:'name'}}};
const base6: {readonly kind:'Holder'}=raw6;
const view6=base6 as Holder;
const value6=view6.value;
console.log(`${value6 === raw6.value}`);
const raw7 = {kind:'Holder' as const,value:{kind:7,child:{label:'name'}}};
const base7: {readonly kind:'Holder'}=raw7;
const view7=base7 as Holder;
const value7=view7.value;
console.log(`${value7 === raw7.value}`);
const raw8 = {kind:'Holder' as const,value:{kind:8,child:{label:'name'}}};
const base8: {readonly kind:'Holder'}=raw8;
const view8=base8 as Holder;
const value8=view8.value;
console.log(`${value8 === raw8.value}`);
const raw9 = {kind:'Holder' as const,value:{kind:9,child:{label:'name'}}};
const base9: {readonly kind:'Holder'}=raw9;
const view9=base9 as Holder;
const value9=view9.value;
console.log(`${value9 === raw9.value}`);
const raw10 = {kind:'Holder' as const,value:{kind:10,child:{label:'name'}}};
const base10: {readonly kind:'Holder'}=raw10;
const view10=base10 as Holder;
const value10=view10.value;
console.log(`${value10 === raw10.value}`);
const raw11 = {kind:'Holder' as const,value:{kind:11,child:{label:'name'}}};
const base11: {readonly kind:'Holder'}=raw11;
const view11=base11 as Holder;
const value11=view11.value;
console.log(`${value11 === raw11.value}`);
const raw12 = {kind:'Holder' as const,value:{flavor:12,child:{label:'name'}}};
const base12: {readonly kind:'Holder'}=raw12;
const view12=base12 as Holder;
const value12=view12.value;
console.log(`${value12 === raw12.value}`);
