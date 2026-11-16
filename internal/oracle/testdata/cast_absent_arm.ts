// A member cast whose source holds undefined beside a represented union (#b5w3ycg slice 2): tsc's shapes,
// checker.ts's `entityName as EntityName` from an EntityName | undefined and `declaration as
// FunctionLikeDeclaration` from a sub-union with undefined, each reached through a parameter, a field, a
// call's result and an optional chain. Every cast here holds a present member, so it passes its absent check
// and its tag check, and the program prints what Node prints. The call's operand is evaluated once (Language's
// v01), and a holder whose other slot holds an explicit undefined, beside one whose slot is missing, casts only
// its present member (v17's honest twin; cast_absent_arm_missing_fails.ts is the lie).
enum Kind { Identifier = 1, Qualified = 2, Function = 3, Method = 4, Signature = 5 }
interface Identifier { readonly kind: Kind.Identifier; readonly text: string }
interface Qualified { readonly kind: Kind.Qualified; readonly left: string; readonly right: string }
interface FunctionDeclaration { readonly kind: Kind.Function; readonly name: string }
interface MethodDeclaration { readonly kind: Kind.Method; readonly name: string }
interface CallSignature { readonly kind: Kind.Signature; readonly parameters: number }
type EntityName = Identifier | Qualified;
type SignatureDeclaration = FunctionDeclaration | MethodDeclaration | CallSignature;
type FunctionLike = FunctionDeclaration | MethodDeclaration;
interface Holder { readonly name: EntityName | undefined; readonly declaration?: SignatureDeclaration }

function nameOf(entity: EntityName): string {
	return entity.kind === Kind.Identifier ? entity.text : entity.left + '.' + entity.right;
}
function fromParameter(entity: EntityName | undefined): string {
	return nameOf(entity as EntityName);
}
function identifierOf(entity: EntityName | undefined): string {
	return (entity as Identifier).text;
}
function fromField(holder: Holder): string {
	return nameOf(holder.name as EntityName);
}
function find(holder: Holder): SignatureDeclaration | undefined {
	return holder.declaration;
}
let finds = 0;
function counted(holder: Holder): SignatureDeclaration | undefined {
	finds++;
	return holder.declaration;
}
function fromCountedCall(holder: Holder): string {
	return (counted(holder) as FunctionLike).name;
}
function fromCall(holder: Holder): string {
	return (find(holder) as FunctionLike).name;
}
function fromOptionalChain(holder: Holder | undefined): string {
	return (holder?.declaration as FunctionLike).name;
}

const named: Holder = { name: { kind: Kind.Qualified, left: 'ts', right: 'Node' }, declaration: { kind: Kind.Method, name: 'visit' } };
console.log(fromParameter({ kind: Kind.Identifier, text: 'checker' }));
console.log(identifierOf({ kind: Kind.Identifier, text: 'binder' }));
console.log(fromField(named));
console.log(fromCall(named));
console.log(fromOptionalChain(named));
console.log(fromCall({ name: undefined, declaration: { kind: Kind.Function, name: 'createSourceFile' } }));
console.log(fromCountedCall(named) + ' ' + finds);
const explicit: Holder = { name: undefined, declaration: { kind: Kind.Function, name: 'bind' } };
const missing: Holder = { name: { kind: Kind.Identifier, text: 'emit' } };
console.log(fromField(missing) + ' ' + fromCall(explicit) + ' ' + (explicit.name === undefined) + ' ' + (missing.declaration === undefined));
