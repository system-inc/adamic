// Control x3, tsc's getParseTreeNode (utilitiesPublic.ts:817 at the stage 3 pin), body whole: a generic
// overload promising T | undefined over a non-generic implementation that returns its parameter before
// reassigning it, then walks original and returns a node only its optional nodeTest vouches for. The
// promise is the predicate's, not a kind's. With nodeTest passed as undefined the overload still
// resolves, T is the argument's type, and the walk returns an original of another kind untested: the
// last call reads a Block's size off an Identifier, undefined on Node.
enum SyntaxKind {
	Unknown = 0,
	Identifier = 80,
	Block = 242,
}

enum NodeFlags {
	None = 0,
	Synthesized = 16,
}

interface Node {
	readonly kind: SyntaxKind;
	readonly flags: NodeFlags;
	readonly original: Node | undefined;
}

interface Identifier extends Node {
	readonly kind: SyntaxKind.Identifier;
	readonly text: string;
}

interface Block extends Node {
	readonly kind: SyntaxKind.Block;
	readonly size: number;
}

function isParseTreeNode(node: Node): boolean {
	return (node.flags & NodeFlags.Synthesized) === 0;
}

function getParseTreeNode(node: Node | undefined): Node | undefined;
function getParseTreeNode<T extends Node>(node: T | undefined, nodeTest?: (node: Node) => node is T): T | undefined;
function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
	if (node === undefined || isParseTreeNode(node)) {
		return node;
	}

	node = node.original;
	while (node) {
		if (isParseTreeNode(node)) {
			return !nodeTest || nodeTest(node) ? node : undefined;
		}
		node = node.original;
	}
}

function isIdentifier(node: Node): node is Identifier {
	return node.kind === SyntaxKind.Identifier;
}

function isBlock(node: Node): node is Block {
	return node.kind === SyntaxKind.Block;
}

const parsed: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.None, original: undefined, text: "parsed" };
const made: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.Synthesized, original: parsed, text: "made" };
const block: Block = { kind: SyntaxKind.Block, flags: NodeFlags.Synthesized, original: parsed, size: 1 };

const found = getParseTreeNode(made, isIdentifier);
console.log(found === undefined ? "none" : found.text);
const refused = getParseTreeNode(block, isBlock);
console.log(refused === undefined ? "none" : `${refused.size}`);
const plain = getParseTreeNode(block);
console.log(plain === undefined ? "none" : `${plain.kind}`);
const untested = getParseTreeNode(block, undefined);
console.log(untested === undefined ? "none" : `${untested.kind} ${untested.size}`);
