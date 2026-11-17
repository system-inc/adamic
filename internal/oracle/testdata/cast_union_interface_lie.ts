// The lie: a value that isn't a NamedDeclaration, cast as one. The cast is a checked view with no test at the cast
// (#b5w3ycg 2d), so it stops at the first read through it, with the view's own message. Node runs on and prints
// undefined.
enum Kind { Variable = 1, Export = 2 }
interface NamedDeclaration { readonly name: string }
interface VariableDeclaration extends NamedDeclaration { readonly kind: Kind.Variable }
interface ExportAssignment { readonly kind: Kind.Export; readonly expression: string }
function cast(node: VariableDeclaration | ExportAssignment): NamedDeclaration { return node as NamedDeclaration; }
console.log('casting');
console.log(typeof cast({ kind: Kind.Export, expression: 'e' }).name);
