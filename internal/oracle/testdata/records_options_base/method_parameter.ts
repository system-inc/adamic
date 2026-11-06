// Language's g4 p06 on #6c4swvt (08:38Z): a method over CompilerOptions written in a literal where Reader's read takes
// OptionsBase. A literal is made as its slot, so no view rule sees the method: it lowers, on main too, and its read stops.
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
const r: Reader = { read(o: CompilerOptions): string { return o.strict ? "on" : "off"; } };
console.log(r.read(base));
