// Two interfaces share one kind (Compiler's pin on #b5w3ycg slice 4), cast to a union holding the narrower one: a
// MethodDeclaration whose name is an Identifier passes the kind test of PrivateIdentifierMethodDeclaration |
// PropertyDeclaration, since the tag proves the kind, not the layout. The read the narrower interface promises and the
// value lacks stops at the read through the view, never reading through. Node runs on and prints undefined.
// cast_kind_shared_fails.ts is the one-kind form.
enum SyntaxKind { Identifier = 80, PrivateIdentifier = 81, PropertyDeclaration = 172, MethodDeclaration = 174 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
interface PrivateIdentifier extends Node { readonly kind: SyntaxKind.PrivateIdentifier; readonly escapedText: string; readonly hash: number }
interface MethodDeclaration extends Node { readonly kind: SyntaxKind.MethodDeclaration; readonly name: Identifier | PrivateIdentifier }
interface PrivateIdentifierMethodDeclaration extends MethodDeclaration { readonly name: PrivateIdentifier }
interface PropertyDeclaration extends Node { readonly kind: SyntaxKind.PropertyDeclaration; readonly name: Identifier | PrivateIdentifier }
function hashOf(node: Node): number {
	const member = node as PrivateIdentifierMethodDeclaration | PropertyDeclaration;
	if (member.kind === SyntaxKind.MethodDeclaration) {
		return member.name.hash;
	}
	return 0;
}
const method: MethodDeclaration = { kind: SyntaxKind.MethodDeclaration, pos: 1, name: { kind: SyntaxKind.Identifier, pos: 2, escapedText: 'visit' } };
console.log('casting');
console.log(typeof hashOf(method));
