// Scratch probe, never cut: tsc's 3076 shape. A CompilerOptions-like record goes through a substitution worker
// typed Readonly<OptionsBase> in and out, comes back cast as CompilerOptions, and is written and read.
type Value = string | number | boolean | undefined;
interface Source {
    fileName: string;
}
interface OptionsBase {
    [option: string]: Value | Source;
}
interface CompilerOptions {
    outDir?: string;
    configFilePath?: string;
    [option: string]: Value | Source;
}
function substitute(options: Readonly<OptionsBase> | undefined): Readonly<OptionsBase> | undefined {
    return options;
}
const given: CompilerOptions = { outDir: "out" };
const result = substitute(given) as CompilerOptions;
result.configFilePath = "x";
console.log(String(result.outDir) + ":" + String(result.configFilePath) + ":" + String(given.configFilePath));
