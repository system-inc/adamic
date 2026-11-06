// The back-cast at tsc's commandLineParser.ts:3098: what handleOptionConfigDirTemplateSubstitution returns, seen as
// OptionsBase, cast back as CompilerOptions. The cast puts named contracts back that the view never held, and the cast
// rule (adamic/no-optional-widening) refuses it ahead of the record view: the reads under the run-time check are
// pinned through the view's writes instead (written_*.ts).
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

function substitute(options: OptionsBase): OptionsBase {
    return options;
}
const original: CompilerOptions = { outDir: "out" };
const options = substitute(original) as CompilerOptions;
console.log(String(options.strict));
