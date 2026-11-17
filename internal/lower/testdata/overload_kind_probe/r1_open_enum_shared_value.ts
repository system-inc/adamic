// Road r1, the kind an open numeric enum with two members sharing a value. Beta and Gamma are both 2,
// so tsc gives them one literal type (getEnumLiteralType keys by enum and value) and Extract by Beta
// keeps Gamma's member too: honest. Loose's kind is the whole enum, so Extract by Alpha leaves Loose out
// though a Loose can carry Alpha's value. The Alpha clause returns such a Loose where the overload
// promises an Alpha: its kind equals the argument's, so a check of kind against the argument's passes
// it, and the caller's read of alpha is undefined on Node. A body proof must take the kind from the
// returned value's type (Loose's is the whole enum), never from the switch's narrowed node.kind.
enum SyntaxKind {
	Unknown = 0,
	Alpha = 1,
	Beta = 2,
	Gamma = 2,
}

interface Node {
	readonly kind: SyntaxKind;
}

interface Alpha extends Node {
	readonly kind: SyntaxKind.Alpha;
	readonly alpha: string;
}

interface Beta extends Node {
	readonly kind: SyntaxKind.Beta;
	readonly beta: number;
}

interface Gamma extends Node {
	readonly kind: SyntaxKind.Gamma;
	readonly gamma: string;
}

interface Loose extends Node {
	readonly kind: SyntaxKind;
	readonly loose: boolean;
}

type Shape = Alpha | Beta | Gamma | Loose;

function rebuild<T extends Shape>(node: T): Extract<Shape, Pick<T, "kind">>;
function rebuild(node: Shape) {
	switch (node.kind) {
		case SyntaxKind.Alpha:
			return makeLoose(node.kind);
		case SyntaxKind.Beta:
			return makeGamma();
	}
	return node;
}

function makeLoose(kind: SyntaxKind): Loose {
	return { kind, loose: true };
}

function makeGamma(): Gamma {
	return { kind: SyntaxKind.Gamma, gamma: "gamma" };
}

const alpha: Alpha = { kind: SyntaxKind.Alpha, alpha: "alpha" };
const beta: Beta = { kind: SyntaxKind.Beta, beta: 2 };
const loose: Loose = { kind: SyntaxKind.Unknown, loose: false };

const rebuiltAlpha = rebuild(alpha);
console.log(`${rebuiltAlpha.kind} ${rebuiltAlpha.alpha}`);
const rebuiltBeta = rebuild(beta);
console.log(`${rebuiltBeta.kind}`);
const rebuiltLoose = rebuild(loose);
console.log(`${rebuiltLoose.kind}`);
