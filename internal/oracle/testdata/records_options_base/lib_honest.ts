// #6c4swvt's string[] head: an honest string[] written under lib, a string[] member CompilerOptions declares, through
// OptionsBase. The view drops lib, since its named read out of the slot is an element-checked view origin (#y852c3c),
// so it lowers, and each element the read hands out is checked as a string.
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
const options: CompilerOptions = { outDir: "out", lib: ["es2020"] };
substitute(options, ["li", "b"].join(""), ["dom", "es2022"]);
const lib = options.lib;
if (lib) {
    for (const name of lib) {
        console.log(name.toUpperCase());
    }
    console.log(String(lib.length) + ":" + (lib[1] ?? "none"));
}
