// Language's g4 p04 on #6c4swvt (08:38Z): a class whose read takes CompilerOptions, seen as a Reader, whose read takes
// OptionsBase. No contract is dropped through a parameter, so it stops not yet at the view.
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
class StrictReader implements Reader {
    read(o: CompilerOptions): string {
        return String(o.strict);
    }
}
const r: Reader = new StrictReader();
console.log(r.read(base));
