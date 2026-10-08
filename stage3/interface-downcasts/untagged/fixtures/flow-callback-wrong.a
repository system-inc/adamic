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
function read(holder:Holder):Target {return holder.value;}
function apply(holder:Holder, callback:(holder:Holder)=>Target):Target {return callback(holder);}
const raw={kind:'Holder' as const,value:{kind:99,text:'name'}};
const base:{readonly kind:'Holder'}=raw;
const view=base as Holder;
const result=apply(view,read);
console.log(`${result!==undefined}`);
