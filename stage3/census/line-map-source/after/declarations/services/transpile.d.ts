import { CompilerOptions, CustomTransformers, Diagnostic, JSDocParsingMode, MapLike } from "./_namespaces/ts.js";
export interface TranspileOptions {
    compilerOptions?: CompilerOptions;
    fileName?: string;
    reportDiagnostics?: boolean;
    moduleName?: string;
    renamedDependencies?: MapLike<string>;
    transformers?: CustomTransformers;
    jsDocParsingMode?: JSDocParsingMode;
}
export interface TranspileOutput {
    outputText: string;
    diagnostics?: Diagnostic[];
    sourceMapText?: string;
}
export declare function transpileModule(input: string, transpileOptions: TranspileOptions): TranspileOutput;
export declare function transpileDeclaration(input: string, transpileOptions: TranspileOptions): TranspileOutput;
export declare function transpile(input: string, compilerOptions?: CompilerOptions, fileName?: string, diagnostics?: Diagnostic[], moduleName?: string): string;
