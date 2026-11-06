// Language's g5 p04 on #6c4swvt (08:38Z): a string written under paths, then its keys. An object-kind field is never
// dropped, so the view stops not yet, before the write.
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
substitute(options, ["pa", "ths"].join(""), "zz");
const p = options.paths;
if (p) {
    console.log(Object.keys(p).join(","));
}
