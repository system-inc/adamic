// Language's must, g5 p06 on #6c4swvt (08:38Z): an honest MapLike written under paths through OptionsBase, which
// natively read as another member of the slot's union. An object-kind field is never dropped: the view stops not yet.
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
substitute(options, ["pa", "ths"].join(""), { k: ["a"] });
const p = options.paths;
if (p) {
    const v = p["k"];
    if (v) { console.log(v.join(",")); }
}
