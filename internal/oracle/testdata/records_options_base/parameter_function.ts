// Language's g4 p01 on #6c4swvt (08:38Z): a function over CompilerOptions seen as one over OptionsBase. Callers hand it
// an OptionsBase, so the view through its parameter adds strict: no contract is dropped there, and it stops not yet.
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
const f: (o: OptionsBase) => string = readStrict;
console.log(f(base));
