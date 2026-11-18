// #6c4swvt's string[] head: an honest string[] written under lib through OptionsBase, then a number pushed into it
// through a (string | number)[] alias the writer kept. The view can't check what arrives later at the read, so the push
// stops at compile time, as every write into an array of a program with a slot origin does (V3's demanded reach).
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
const wide: (string | number)[] = ["dom"];
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["li", "b"].join(""), wide);
wide.push(1);
const lib = options.lib;
if (lib) {
    console.log(typeof lib[1]);
}
