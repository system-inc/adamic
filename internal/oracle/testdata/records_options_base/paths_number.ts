// Language's g5 p03 on #6c4swvt (08:38Z): a number written under paths, an object-kind field, through OptionsBase. An
// object-kind field is never dropped, so the view stops not yet, before the write.
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
substitute(options, ["pa", "ths"].join(""), 5);
const p = options.paths;
if (p) {
    for (const key in p) { console.log(key); }
    console.log("after");
}
