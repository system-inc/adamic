// Road r3, a getter-backed kind: each class's kind is a get accessor typed with one literal, and it
// counts its reads. The switch reads it once per call, so Node prints 2 reads. A check at the call that
// reads the result's kind (and the argument's) runs the getter again, and the count tells.
enum SyntaxKind {
	Unknown = 0,
	Alpha = 1,
	Beta = 2,
}

let reads = 0;

class AlphaNode {
	readonly alpha: string;
	constructor(alpha: string) {
		this.alpha = alpha;
	}
	get kind(): SyntaxKind.Alpha {
		reads = reads + 1;
		return SyntaxKind.Alpha;
	}
}

class BetaNode {
	readonly beta: number;
	constructor(beta: number) {
		this.beta = beta;
	}
	get kind(): SyntaxKind.Beta {
		reads = reads + 1;
		return SyntaxKind.Beta;
	}
}

type Shape = AlphaNode | BetaNode;

function renew<T extends Shape>(node: T): Extract<Shape, Pick<T, "kind">>;
function renew(node: Shape) {
	switch (node.kind) {
		case SyntaxKind.Alpha:
			return new AlphaNode(`${node.alpha}!`);
		case SyntaxKind.Beta:
			return new BetaNode(node.beta + 1);
	}
}

const renewedAlpha = renew(new AlphaNode("a"));
console.log(renewedAlpha.alpha);
const renewedBeta = renew(new BetaNode(1));
console.log(`${renewedBeta.beta}`);
console.log(`${reads}`);
