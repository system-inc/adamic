// #6c4swvt, Language's layout pin (J-jp1b3bx c03): a class instance holding TsConfigSourceFile's fields, written under
// configFile through OptionsBase, then read as a TsConfigSourceFile. Its fields must be read by name from the class's
// own layout, which puts a field of its own before them.
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
class File {
    version = 2;
    fileName: string;
    text: string;
    constructor(fileName: string, text: string) {
        this.fileName = fileName;
        this.text = text;
    }
}
substitute(options, configKey(), new File("c.json", "{}x"));
const config = options.configFile;
if (config) {
    console.log(config.fileName + ":" + String(config.text.length));
}
