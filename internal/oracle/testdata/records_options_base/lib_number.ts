// #6c4swvt's string[] head, a lie: a (string | number)[] holding a 1 written under lib through OptionsBase. Node reads the
// 1; both backends stop at the element read that hands it out, after the line before it, never reading it as a string.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    lib?: string[];
    outDir?: string;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}
function substitute(options: OptionsBase, name: string, value: CompilerOptionsValue): void {
    options[name] = value;
}
const wide: (string | number)[] = ["dom", 1];
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["li", "b"].join(""), wide);
const lib = options.lib;
if (lib) {
    console.log("before");
    const second = lib[1];
    console.log(typeof second);
}
