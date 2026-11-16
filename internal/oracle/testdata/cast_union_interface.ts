// A represented union cast to one interface some of its members satisfy (#b5w3ycg 2d): tsc's
// `node as NamedDeclaration` and `input as Declaration` of a DeclarationDiagnosticProducing, whose members are each
// a declaration or not. The tag test keeps the members that upcast to the interface; every cast here holds one, so
// the program prints what Node prints.
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
