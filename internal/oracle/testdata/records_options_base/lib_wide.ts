// #6c4swvt's string[] head: a (string | number)[] holding only strings written under lib through OptionsBase. Its elements
// are boxed as the wider array's, and each one the read hands out is checked as a string and passes, as Node reads it.
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
const wide: (string | number)[] = ["dom", "es2022"];
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["li", "b"].join(""), wide);
const lib = options.lib;
if (lib) {
    for (const name of lib) {
        console.log(name.toUpperCase());
    }
}
