import { type HasDecorators, type HasModifiers, type Node, type TextRange } from "../_namespaces/ts.js";
export declare function setTextRange<T extends TextRange>(range: T, location: Readonly<TextRange> | undefined): T;
export declare function canHaveModifiers(node: Node): node is HasModifiers;
export declare function canHaveDecorators(node: Node): node is HasDecorators;
