// Control x2, tsc's transformNamedEvaluation (transformers/namedEvaluation.ts:469 at the stage 3 pin): a
// generic overload promising Extract<NamedEvaluation, Pick<T, "kind" | keyof T & "operatorToken" | keyof
// T & "name">> over a non-generic implementation whose switch on node.kind returns each case's helper,
// each a factory update of the narrowed node typed as the plain node, not as NamedEvaluation's member.
// NamedEvaluation's members are intersections, as tsc's are, so the promise restricts more than the kind.
enum SyntaxKind {
	Unknown = 0,
	StringLiteral = 11,
	EqualsToken = 64,
	BarBarEqualsToken = 76,
	Identifier = 80,
	BinaryExpression = 226,
	VariableDeclaration = 260,
	PropertyAssignment = 303,
}

interface Node {
	readonly kind: SyntaxKind;
}

interface Identifier extends Node {
	readonly kind: SyntaxKind.Identifier;
	readonly text: string;
}

interface StringLiteral extends Node {
	readonly kind: SyntaxKind.StringLiteral;
	readonly text: string;
}

interface EqualsToken extends Node {
	readonly kind: SyntaxKind.EqualsToken;
}

interface BarBarEqualsToken extends Node {
	readonly kind: SyntaxKind.BarBarEqualsToken;
}

interface PropertyAssignment extends Node {
	readonly kind: SyntaxKind.PropertyAssignment;
	readonly name: Identifier | StringLiteral;
	readonly initializer: string;
}

interface VariableDeclaration extends Node {
	readonly kind: SyntaxKind.VariableDeclaration;
	readonly name: Identifier | StringLiteral;
	readonly initializer: string;
}

interface BinaryExpression extends Node {
	readonly kind: SyntaxKind.BinaryExpression;
	readonly left: Identifier | StringLiteral;
	readonly operatorToken: EqualsToken | BarBarEqualsToken;
	readonly right: string;
}

type NamedEvaluation =
	| PropertyAssignment & { readonly name: Identifier; }
	| VariableDeclaration & { readonly name: Identifier; }
	| BinaryExpression & { readonly left: Identifier; readonly operatorToken: EqualsToken; }
	| BinaryExpression & { readonly left: Identifier; readonly operatorToken: BarBarEqualsToken; };

function transformNamedEvaluation<T extends NamedEvaluation>(node: T, assignedName?: string): Extract<NamedEvaluation, Pick<T, "kind" | keyof T & "operatorToken" | keyof T & "name">>;
function transformNamedEvaluation(node: NamedEvaluation, assignedName?: string) {
	switch (node.kind) {
		case SyntaxKind.PropertyAssignment:
			return transformNamedEvaluationOfPropertyAssignment(node, assignedName);
		case SyntaxKind.VariableDeclaration:
			return transformNamedEvaluationOfVariableDeclaration(node, assignedName);
		case SyntaxKind.BinaryExpression:
			return transformNamedEvaluationOfAssignmentExpression(node, assignedName);
	}
}

function transformNamedEvaluationOfPropertyAssignment(node: NamedEvaluation & PropertyAssignment, assignedName: string | undefined) {
	return updatePropertyAssignment(node, node.name, finishTransformNamedEvaluation(node.initializer, assignedName ?? node.name.text));
}

function transformNamedEvaluationOfVariableDeclaration(node: NamedEvaluation & VariableDeclaration, assignedName: string | undefined) {
	return updateVariableDeclaration(node, node.name, finishTransformNamedEvaluation(node.initializer, assignedName ?? node.name.text));
}

function transformNamedEvaluationOfAssignmentExpression(node: NamedEvaluation & BinaryExpression, assignedName: string | undefined) {
	return updateBinaryExpression(node, node.left, node.operatorToken, finishTransformNamedEvaluation(node.right, assignedName ?? node.left.text));
}

function finishTransformNamedEvaluation(expression: string, assignedName: string): string {
	return `${expression} named ${assignedName}`;
}

function updatePropertyAssignment(node: PropertyAssignment, name: Identifier | StringLiteral, initializer: string): PropertyAssignment {
	return node.name === name && node.initializer === initializer ? node : { kind: SyntaxKind.PropertyAssignment, name, initializer };
}

function updateVariableDeclaration(node: VariableDeclaration, name: Identifier | StringLiteral, initializer: string): VariableDeclaration {
	return node.name === name && node.initializer === initializer ? node : { kind: SyntaxKind.VariableDeclaration, name, initializer };
}

function updateBinaryExpression(node: BinaryExpression, left: Identifier | StringLiteral, operatorToken: EqualsToken | BarBarEqualsToken, right: string): BinaryExpression {
	return node.left === left && node.operatorToken === operatorToken && node.right === right ? node : { kind: SyntaxKind.BinaryExpression, left, operatorToken, right };
}

const name: Identifier = { kind: SyntaxKind.Identifier, text: "x" };
const equals: EqualsToken = { kind: SyntaxKind.EqualsToken };
const orEquals: BarBarEqualsToken = { kind: SyntaxKind.BarBarEqualsToken };

// Every tsc caller reassigns a node-specific variable with the result.
let property: PropertyAssignment & { readonly name: Identifier; } = { kind: SyntaxKind.PropertyAssignment, name, initializer: "class {}" };
property = transformNamedEvaluation(property);
let variable: VariableDeclaration & { readonly name: Identifier; } = { kind: SyntaxKind.VariableDeclaration, name, initializer: "function () {}" };
variable = transformNamedEvaluation(variable, "v");
let assignment: BinaryExpression & { readonly left: Identifier; readonly operatorToken: BarBarEqualsToken; } = { kind: SyntaxKind.BinaryExpression, left: name, operatorToken: orEquals, right: "() => 1" };
assignment = transformNamedEvaluation(assignment);
let plain: BinaryExpression & { readonly left: Identifier; readonly operatorToken: EqualsToken; } = { kind: SyntaxKind.BinaryExpression, left: name, operatorToken: equals, right: "class {}" };
plain = transformNamedEvaluation(plain, "");

console.log(`${property.kind} ${property.name.text} ${property.initializer}`);
console.log(`${variable.kind} ${variable.name.text} ${variable.initializer}`);
console.log(`${assignment.kind} ${assignment.operatorToken.kind} ${assignment.left.text} ${assignment.right}`);
console.log(`${plain.kind} ${plain.operatorToken.kind} ${plain.right}`);
