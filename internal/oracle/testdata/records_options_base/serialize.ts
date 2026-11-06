// tsc's serializeOptionBaseObject (commandLineParser.ts:2795) and its callers at 2788 and 2792: a CompilerOptions and
// a WatchOptions, records whose named options OptionsBase drops, read through OptionsBase by for...in, an own-key test
// and options[name], into a fresh list. The view only reads the object it's handed, keys in insertion order. The
// object-kind options are index entries here, not declared: a declared one keeps the view not yet (paths_*.ts).
interface MapLike<T> {
    [index: string]: T;
}
enum WatchFileKind {
    FixedPollingInterval = 0,
    UseFsEvents = 4,
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    strict?: boolean;
    outDir?: string;
    target?: number;
    [option: string]: CompilerOptionsValue | undefined;
}
interface WatchOptions {
    watchFile?: WatchFileKind;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;

function text(value: CompilerOptionsValue): string {
    if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
        return `${typeof value} ${String(value)}`;
    }
    if (value === null || value === undefined) {
        return value === null ? "null" : "undefined";
    }
    return typeof value;
}
function serializeOptionBaseObject(options: OptionsBase): string {
    const result: string[] = [];
    for (const name in options) {
        if (hasOwnProperty.call(options, name)) {
            result.push(`${name}=${text(options[name])}`);
        }
    }
    return result.join(" ");
}

const extra: string = ["no", "Emit"].join("");
const compilerOptions: CompilerOptions = { outDir: "out", strict: true, lib: ["es2022", "dom"], target: 9, paths: { "*": ["types/*"] }, [extra]: false };
const watchOptions: WatchOptions = { watchFile: WatchFileKind.UseFsEvents, excludeFiles: ["a.ts"] };
console.log(serializeOptionBaseObject(compilerOptions));
console.log(serializeOptionBaseObject(watchOptions));
console.log(serializeOptionBaseObject({}));
console.log(`${compilerOptions.strict}:${compilerOptions.outDir}:${watchOptions.watchFile}`);
