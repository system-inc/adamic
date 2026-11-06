// Language's g4 p02 on #6c4swvt (08:38Z): a function over CompilerOptions as Reader's read, which takes OptionsBase,
// through method bivariance. No contract is dropped through a parameter, so it stops not yet at the view.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    strict?: boolean;
    outDir?: string;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}
function readStrict(o: CompilerOptions): string {
    return String(o.strict);
}
const base: OptionsBase = {};
base[["str", "ict"].join("")] = ["y", "es"].join("");
interface Reader {
    read(o: OptionsBase): string;
}
const r: Reader = { read: readStrict };
console.log(r.read(base));
