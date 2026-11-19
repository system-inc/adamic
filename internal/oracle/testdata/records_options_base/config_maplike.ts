// #6c4swvt's dictionary and fixed-object head, a lie: a MapLike written under configFile through OptionsBase. Node
// reads its missing fileName; both backends stop at the named read, whose narrow to a fixed object finds a dictionary.
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
substitute(options, configKey(), { k: ["a"] });
console.log("before");
const config = options.configFile;
if (config) {
    console.log(String(config.fileName));
}
