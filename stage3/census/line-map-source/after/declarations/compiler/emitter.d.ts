import { type CompilerOptions, type ParsedCommandLine, type Printer, type PrinterOptions, type PrintHandlers } from "./_namespaces/ts.js";
export declare function getTsBuildInfoEmitOutputFilePath(options: CompilerOptions): string | undefined;
export declare function getOutputFileNames(commandLine: ParsedCommandLine, inputFileName: string, ignoreCase: boolean): readonly string[];
export declare function createPrinter(printerOptions?: PrinterOptions, handlers?: PrintHandlers): Printer;
