// Control x1, tsc's injectClassNamedEvaluationHelperBlockIfMissing (transformers/namedEvaluation.ts:173 at
// the stage 3 pin): a generic overload promising Extract<ClassLikeDeclaration, Pick<T, "kind">> over a
// non-generic implementation that returns its parameter after a helper's test, tests a member, and
// reassigns the parameter through an isClassDeclaration conditional. Its three callers are tsc's own.
enum SyntaxKind {
	Unknown = 0,
	ClassExpression = 231,
	ClassDeclaration = 263,
}

interface Node {
	readonly kind: SyntaxKind;
}

interface ClassDeclaration extends Node {
	readonly kind: SyntaxKind.ClassDeclaration;
	readonly name: string;
	readonly members: number;
}

interface ClassExpression extends Node {
	readonly kind: SyntaxKind.ClassExpression;
	readonly name: string | undefined;
	readonly members: number;
}

type ClassLikeDeclaration = ClassDeclaration | ClassExpression;

function injectClassNamedEvaluationHelperBlockIfMissing<T extends ClassLikeDeclaration>(node: T, assignedName: string): Extract<ClassLikeDeclaration, Pick<T, "kind">>;
function injectClassNamedEvaluationHelperBlockIfMissing(node: ClassLikeDeclaration, assignedName: string) {
	if (classHasExplicitlyAssignedName(node)) {
		return node;
	}
	let members = node.members + 1;
	if (node.name !== undefined) {
		members = members + assignedName.length;
	}
	node = isClassDeclaration(node) ? updateClassDeclaration(node, members) : updateClassExpression(node, members);
	return node;
}

function classHasExplicitlyAssignedName(node: ClassLikeDeclaration): boolean {
	return node.members > 9;
}

function isClassDeclaration(node: Node): node is ClassDeclaration {
	return node.kind === SyntaxKind.ClassDeclaration;
}

function updateClassDeclaration(node: ClassDeclaration, members: number): ClassDeclaration {
	return { kind: SyntaxKind.ClassDeclaration, name: node.name, members };
}

function updateClassExpression(node: ClassExpression, members: number): ClassExpression {
	return { kind: SyntaxKind.ClassExpression, name: node.name, members };
}

function either(pickDeclaration: boolean): ClassLikeDeclaration {
	return pickDeclaration ? declaration : expression;
}

const declaration: ClassDeclaration = { kind: SyntaxKind.ClassDeclaration, name: "Declared", members: 2 };
const expression: ClassExpression = { kind: SyntaxKind.ClassExpression, name: undefined, members: 3 };

// esDecorators.ts:1062: a ClassDeclaration variable reassigned with the narrow result.
let declared: ClassDeclaration = declaration;
declared = injectClassNamedEvaluationHelperBlockIfMissing(declared, "C");
// esDecorators.ts:676: a ClassLikeDeclaration variable reassigned with its union.
let like = either(false);
like = injectClassNamedEvaluationHelperBlockIfMissing(like, "default");
// namedEvaluation.ts:250: a ClassExpression whose result is read as one.
const injected = injectClassNamedEvaluationHelperBlockIfMissing(expression, "E");

console.log(`${declared.kind} ${declared.name} ${declared.members}`);
console.log(`${like.kind} ${like.members}`);
console.log(`${injected.kind} ${injected.name === undefined ? "anonymous" : injected.name} ${injected.members}`);
