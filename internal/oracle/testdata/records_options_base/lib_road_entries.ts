// #6c4swvt's string[] head, Runtime's road proof (17:11Z): a 1 at lib[1], written under lib through OptionsBase, read
// back by Object.entries over the record, the value seen as string[]. Node reads the 1; no backend may read it as a
// string: it stops at compile time or at the read.
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
function libKey(): string {
    return "lib";
}
const wide: (string | number)[] = ["dom", 1];
const options: CompilerOptions = { outDir: "out" };
substitute(options, libKey(), wide);
console.log("before");
for (const [key, value] of Object.entries(options)) {
    if (key === "lib" && Array.isArray(value)) {
        const lib = value as string[];
        console.log(typeof lib[1]);
    }
}
