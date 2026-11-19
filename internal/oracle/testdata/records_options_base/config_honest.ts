// #6c4swvt's dictionary and fixed-object head: an honest TsConfigSourceFile written under configFile, a readonly fixed-
// object member CompilerOptions declares, through OptionsBase. Its read out of the slot is a checked heap narrow to a
// fixed object, the slot union's only one, so the view drops it, and the read agrees with Node.
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
const file: TsConfigSourceFile = { fileName: "tsconfig.json", text: "{}" };
substitute(options, configKey(), file);
const config = options.configFile;
if (config) {
    console.log(config.fileName + ":" + String(config.text.length));
}
