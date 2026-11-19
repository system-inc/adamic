// #6c4swvt, Language's layout pin (J-jp1b3bx c01): a TsConfigSourceFile of its fields in the other order, text
// first, written under configFile through OptionsBase, then read as a TsConfigSourceFile. fileName must be read by
// name, never at TsConfigSourceFile's own offset, where this value holds text.
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
interface Reordered {
    text: string;
    fileName: string;
}
const reordered: Reordered = { text: "{\"a\":1}", fileName: "f.json" };
substitute(options, configKey(), reordered);
const config = options.configFile;
if (config) {
    console.log(config.fileName + ":" + String(config.text.length));
}
