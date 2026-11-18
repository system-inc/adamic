// #6c4swvt's string[] head, a lie: a string written under lib through OptionsBase. Node iterates its characters; both
// backends stop at the named read, whose narrow to an array finds another member of the slot's union.
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
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["li", "b"].join(""), "dom");
console.log("before");
const lib = options.lib;
if (lib) {
    for (const name of lib) {
        console.log(name);
    }
}
