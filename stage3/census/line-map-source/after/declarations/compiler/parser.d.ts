import { type EntityName, JSDocParsingMode, type JsonSourceFile, type Node, type NodeArray, type ResolutionMode, ScriptKind, ScriptTarget, type SourceFile, type TextChangeRange } from "./_namespaces/ts.js";
/**
 * Invokes a callback for each child of the given node. The 'cbNode' callback is invoked for all child nodes
 * stored in properties. If a 'cbNodes' callback is specified, it is invoked for embedded arrays; otherwise,
 * embedded arrays are flattened and the 'cbNode' callback is invoked for each element. If a callback returns
 * a truthy value, iteration stops and that value is returned. Otherwise, undefined is returned.
 *
 * @param node a given node to visit its children
 * @param cbNode a callback to be invoked for all child nodes
 * @param cbNodes a callback to be invoked for embedded array
 *
 * @remarks `forEachChild` must visit the children of a node in the order
 * that they appear in the source code. The language service depends on this property to locate nodes by position.
 */
export declare function forEachChild<T>(node: Node, cbNode: (node: Node) => T | undefined, cbNodes?: (nodes: NodeArray<Node>) => T | undefined): T | undefined;
export interface CreateSourceFileOptions {
    languageVersion: ScriptTarget;
    /**
     * Controls the format the file is detected as - this can be derived from only the path
     * and files on disk, but needs to be done with a module resolution cache in scope to be performant.
     * This is usually `undefined` for compilations that do not have `moduleResolution` values of `node16` or `nodenext`.
     */
    impliedNodeFormat?: ResolutionMode;
    /**
     * Controls how module-y-ness is set for the given file. Usually the result of calling
     * `getSetExternalModuleIndicator` on a valid `CompilerOptions` object. If not present, the default
     * check specified by `isFileProbablyExternalModule` will be used to set the field.
     */
    setExternalModuleIndicator?: (file: SourceFile) => void;
    jsDocParsingMode?: JSDocParsingMode | undefined;
}
export declare function createSourceFile(fileName: string, sourceText: string, languageVersionOrOptions: ScriptTarget | CreateSourceFileOptions, setParentNodes?: boolean, scriptKind?: ScriptKind): SourceFile;
export declare function parseIsolatedEntityName(text: string, languageVersion: ScriptTarget): EntityName | undefined;
/**
 * Parse json text into SyntaxTree and return node and parse errors if any
 * @param fileName
 * @param sourceText
 */
export declare function parseJsonText(fileName: string, sourceText: string): JsonSourceFile;
export declare function isExternalModule(file: SourceFile): boolean;
export declare function updateSourceFile(sourceFile: SourceFile, newText: string, textChangeRange: TextChangeRange, aggressiveChecks?: boolean): SourceFile;
