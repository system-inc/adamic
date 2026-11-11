import { type AccessExpression, EmitFlags, type EmitHelper, type Node, type SourceFile, type SourceMapRange, SyntaxKind, type SynthesizedComment, type TextRange } from "../_namespaces/ts.js";
/**
 * Clears any `EmitNode` entries from parse-tree nodes.
 * @param sourceFile A source file.
 */
export declare function disposeEmitNodes(sourceFile: SourceFile | undefined): void;
/**
 * Sets flags that control emit behavior of a node.
 */
export declare function setEmitFlags<T extends Node>(node: T, emitFlags: EmitFlags): T;
/**
 * Gets a custom text range to use when emitting source maps.
 */
export declare function getSourceMapRange(node: Node): SourceMapRange;
/**
 * Sets a custom text range to use when emitting source maps.
 */
export declare function setSourceMapRange<T extends Node>(node: T, range: SourceMapRange | undefined): T;
/**
 * Gets the TextRange to use for source maps for a token of a node.
 */
export declare function getTokenSourceMapRange(node: Node, token: SyntaxKind): SourceMapRange | undefined;
/**
 * Sets the TextRange to use for source maps for a token of a node.
 */
export declare function setTokenSourceMapRange<T extends Node>(node: T, token: SyntaxKind, range: SourceMapRange | undefined): T;
/**
 * Gets a custom text range to use when emitting comments.
 */
export declare function getCommentRange(node: Node): TextRange;
/**
 * Sets a custom text range to use when emitting comments.
 */
export declare function setCommentRange<T extends Node>(node: T, range: TextRange): T;
export declare function getSyntheticLeadingComments(node: Node): SynthesizedComment[] | undefined;
export declare function setSyntheticLeadingComments<T extends Node>(node: T, comments: SynthesizedComment[] | undefined): T;
export declare function addSyntheticLeadingComment<T extends Node>(node: T, kind: SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia, text: string, hasTrailingNewLine?: boolean): T;
export declare function getSyntheticTrailingComments(node: Node): SynthesizedComment[] | undefined;
export declare function setSyntheticTrailingComments<T extends Node>(node: T, comments: SynthesizedComment[] | undefined): T;
export declare function addSyntheticTrailingComment<T extends Node>(node: T, kind: SyntaxKind.SingleLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia, text: string, hasTrailingNewLine?: boolean): T;
export declare function moveSyntheticComments<T extends Node>(node: T, original: Node): T;
/**
 * Gets the constant value to emit for an expression representing an enum.
 */
export declare function getConstantValue(node: AccessExpression): string | number | undefined;
/**
 * Sets the constant value to emit for an expression.
 */
export declare function setConstantValue(node: AccessExpression, value: string | number): AccessExpression;
/**
 * Adds an EmitHelper to a node.
 */
export declare function addEmitHelper<T extends Node>(node: T, helper: EmitHelper): T;
/**
 * Add EmitHelpers to a node.
 */
export declare function addEmitHelpers<T extends Node>(node: T, helpers: EmitHelper[] | undefined): T;
/**
 * Removes an EmitHelper from a node.
 */
export declare function removeEmitHelper(node: Node, helper: EmitHelper): boolean;
/**
 * Gets the EmitHelpers of a node.
 */
export declare function getEmitHelpers(node: Node): EmitHelper[] | undefined;
/**
 * Moves matching emit helpers from a source node to a target node.
 */
export declare function moveEmitHelpers(source: Node, target: Node, predicate: (helper: EmitHelper) => boolean): void;
