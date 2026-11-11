import { type Node, type NodeFactory, type SourceMapSource } from "../_namespaces/ts.js";
export declare const factory: NodeFactory;
declare let SourceMapSource: new (fileName: string, text: string, skipTrivia?: (pos: number) => number) => SourceMapSource;
/**
 * Create an external source map source file reference
 */
export declare function createSourceMapSource(fileName: string, text: string, skipTrivia?: (pos: number) => number): SourceMapSource;
export declare function setOriginalNode<T extends Node>(node: T, original: Node | undefined): T;
export {};
