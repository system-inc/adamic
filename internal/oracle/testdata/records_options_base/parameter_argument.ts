// Language's g4 p03 on #6c4swvt (08:38Z): { read: readStrict } handed where a Reader is taken, read over
// CompilerOptions, Reader's over OptionsBase. No contract is dropped through a parameter, so it stops not yet.
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
function run(reader: Reader): string {
    return reader.read(base);
}
console.log(run({ read: readStrict }));
