// #6c4swvt's dictionary and fixed-object head: a fixed object of a wider shape than TsConfigSourceFile, assignable to
// it, written under configFile through OptionsBase, then read as a TsConfigSourceFile. Its field reads must agree with
// Node, or stop.
interface TsConfigSourceFile {
    fileName: string;
    text: string;
}
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    readonly configFile?: TsConfigSourceFile;
    outDir?: string;
    [option: string]: CompilerOptionsValue | TsConfigSourceFile | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | TsConfigSourceFile | undefined;
}
function substitute(options: OptionsBase, name: string, value: CompilerOptionsValue | TsConfigSourceFile): void {
    options[name] = value;
}
function configKey(): string {
    return "configFile";
}
const options: CompilerOptions = { outDir: "out" };
interface Parsed {
    kind: number;
    fileName: string;
    text: string;
}
const parsed: Parsed = { kind: 7, fileName: "tsconfig.json", text: "{}" };
substitute(options, configKey(), parsed);
const config = options.configFile;
if (config) {
    console.log(config.fileName + ":" + String(config.text.length));
}
