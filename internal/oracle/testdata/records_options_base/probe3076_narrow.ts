// Scratch probe, never cut: tsc's 3076 shape with a narrower record. A WatchOptions-like record goes through the
// same worker, comes back cast as CompilerOptions, and the cast-and-write puts a Source under a key the record's own
// index reads as Value; then the original is read by that key.
type Value = string | number | boolean | undefined;
interface Source {
    fileName: string;
}
interface OptionsBase {
    [option: string]: Value | Source;
}
interface CompilerOptions {
    outDir?: string;
    configFile?: Source;
    [option: string]: Value | Source;
}
interface WatchOptions {
    watchFile?: number;
    [option: string]: Value;
}
function substitute(options: Readonly<OptionsBase> | undefined): Readonly<OptionsBase> | undefined {
    return options;
}
const watch: WatchOptions = { watchFile: 4 };
const result = substitute(watch) as CompilerOptions;
result.configFile = { fileName: "tsconfig.json" };
const read = watch["configFile"];
console.log(typeof read + ":" + String(typeof read === "object"));
