// #6c4swvt (a): a dictionary put under configFile, a fixed-object member CompilerOptions declares, by a computed entry
// of a record literal, which recordComputedEntry leaves to the read (readGuardedInUnionSlot). Node reads its missing
// fileName; the read's narrow to a fixed object stops on both backends alike, the JavaScript backend telling the record
// by the mark every record it makes carries.
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
function configKey(): string {
    return "configFile";
}
const paths: MapLike<string[]> = { k: ["a"] };
const options: CompilerOptions = { outDir: "out", [configKey()]: paths };
console.log("before");
const config = options.configFile;
if (config) {
    console.log(String(config.fileName));
}
