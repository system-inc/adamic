// The lie: the member that isn't a NamedDeclaration, cast as one. Its tag fails the test with the cast's message.
// Node runs on without a check and prints undefined.
enum Kind { Variable = 1, Export = 2 }
interface NamedDeclaration { readonly name: string }
interface VariableDeclaration extends NamedDeclaration { readonly kind: Kind.Variable }
interface ExportAssignment { readonly kind: Kind.Export; readonly expression: string }
function cast(node: VariableDeclaration | ExportAssignment): NamedDeclaration { return node as NamedDeclaration; }
console.log('casting');
console.log(typeof cast({ kind: Kind.Export, expression: 'e' }).name);
