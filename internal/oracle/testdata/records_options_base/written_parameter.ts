// Language's condition on #6c4swvt (2026-11-06 03:08Z): the check sits on every named read of the record's value, not
// on the view, so it follows the value. A write through OptionsBase puts the string "yes" under strict, declared
// boolean, in the same object the program holds as CompilerOptions, and a parameter typed CompilerOptions that reads it there stops on it.
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

function substitute(options: OptionsBase, name: string): void {
    options[name] = ["y", "es"].join("");
}
const options: CompilerOptions = { outDir: "out" };
substitute(options, ["str", "ict"].join(""));
console.log(`written ${Object.keys(options).join(",")}`);
function show(compilerOptions: CompilerOptions): string {
    return String(compilerOptions.strict);
}
console.log(show(options));
