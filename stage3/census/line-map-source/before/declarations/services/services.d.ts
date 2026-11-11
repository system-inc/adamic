import { CompilerOptions, CreateSourceFileOptions, DocumentRegistry, EditorOptions, EditorSettings, IScriptSnapshot, LanguageService, LanguageServiceHost, LanguageServiceMode, ScriptKind, ScriptTarget, SourceFile, SymbolDisplayPart, TextChangeRange } from "./_namespaces/ts.js";
/** The version of the language service API */
export declare const servicesVersion = "0.8";
export declare function toEditorSettings(options: EditorOptions | EditorSettings): EditorSettings;
export declare function displayPartsToString(displayParts: SymbolDisplayPart[] | undefined): string;
export declare function getDefaultCompilerOptions(): CompilerOptions;
export declare function getSupportedCodeFixes(): readonly string[];
export declare function createLanguageServiceSourceFile(fileName: string, scriptSnapshot: IScriptSnapshot, scriptTargetOrOptions: ScriptTarget | CreateSourceFileOptions, version: string, setNodeParents: boolean, scriptKind?: ScriptKind): SourceFile;
export declare function updateLanguageServiceSourceFile(sourceFile: SourceFile, scriptSnapshot: IScriptSnapshot, version: string, textChangeRange: TextChangeRange | undefined, aggressiveChecks?: boolean): SourceFile;
export declare function createLanguageService(host: LanguageServiceHost, documentRegistry?: DocumentRegistry, syntaxOnlyOrLanguageServiceMode?: boolean | LanguageServiceMode): LanguageService;
/**
 * Get the path of the default library files (lib.d.ts) as distributed with the typescript
 * node package.
 * The functionality is not supported if the ts module is consumed outside of a node module.
 */
export declare function getDefaultLibFilePath(options: CompilerOptions): string;
