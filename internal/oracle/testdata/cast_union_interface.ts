// A represented union cast to one interface some of its members satisfy (#b5w3ycg 2d): tsc's
// `node as NamedDeclaration` and `input as Declaration` of a DeclarationDiagnosticProducing, whose members are each
// a declaration or not. The tag test keeps the members that upcast to the interface; every cast here holds one, so
// the program prints what Node prints. The kept members hold name at different places in their objects (name
// second, name last, name among more fields), and the read after the cast goes through the cast's checked view
// by name, never by one member's offset (Lower's question).
enum Kind { Variable = 1, Parameter = 2, Property = 3, Export = 4 }
interface Declaration { readonly pos: number }
interface NamedDeclaration extends Declaration { readonly name: string }
interface VariableDeclaration extends NamedDeclaration { readonly kind: Kind.Variable; readonly initializer: number }
interface ParameterDeclaration extends NamedDeclaration { readonly kind: Kind.Parameter; readonly optional: boolean }
interface PropertyDeclaration extends NamedDeclaration { readonly kind: Kind.Property }
interface ExportAssignment { readonly kind: Kind.Export; readonly expression: string }
type DiagnosticProducing = VariableDeclaration | ParameterDeclaration | PropertyDeclaration | ExportAssignment;

function nameOf(node: DiagnosticProducing): string {
	return (node as NamedDeclaration).name;
}
function isDeclaration(input: DiagnosticProducing): string {
	const declaration = input as Declaration;
	return String(declaration.pos);
}
console.log(nameOf({ kind: Kind.Variable, name: 'x', initializer: 1, pos: 7 }));
console.log(nameOf({ kind: Kind.Parameter, name: 'p', optional: true, pos: 7 }));
console.log(nameOf({ kind: Kind.Property, name: 'size', pos: 7 }));
console.log(isDeclaration({ kind: Kind.Property, name: 'size', pos: 7 }));
const nodes: DiagnosticProducing[] = [
	{ pos: 1, optional: false, kind: Kind.Parameter, name: 'last' },
	{ initializer: 2, pos: 3, kind: Kind.Variable, name: 'after' },
	{ name: 'first', kind: Kind.Property, pos: 4 },
];
for (const node of nodes) {
	console.log(nameOf(node) + ' ' + isDeclaration(node));
}
// Language's width roads (p02, p03, p08): a value of the member NamedDeclaration doesn't name, holding a name anyway.
// The cast claims the shape, the value has it, and Node and both backends read it.
interface NamedExport extends ExportAssignment { readonly name: string }
const named: NamedExport = { kind: Kind.Export, expression: 'e', name: 'hidden' };
const held: ExportAssignment = named;
console.log(nameOf(held));
function viaParameter(assignment: ExportAssignment): string {
	return nameOf(assignment);
}
console.log(viaParameter({ kind: Kind.Export, expression: 'p', name: 'param' } as NamedExport));
class BoxedExport {
	readonly kind = Kind.Export as const;
	readonly expression = 'b';
	readonly name = 'boxed';
}
console.log(nameOf(new BoxedExport()));
