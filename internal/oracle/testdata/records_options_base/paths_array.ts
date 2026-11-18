// Language's g5 p02 on #6c4swvt (08:38Z): a string array written under paths, a dictionary field, through OptionsBase.
// Since #6c4swvt's dictionary head the view drops paths, and the named read's narrow to a dictionary stops on the lie
// on both backends.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    paths?: MapLike<string[]>;
    outDir?: string;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}
function substitute(options: OptionsBase, name: string, value: CompilerOptionsValue): void {
    options[name] = value;
}
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["pa", "ths"].join(""), ["a", "b"]);
const p = options.paths;
if (p) {
    const v = p["0"];
    console.log(typeof v);
    if (v) { console.log(String(v.length)); }
}
