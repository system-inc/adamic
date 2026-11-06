// tsc's toIncrementalBuildInfoCompilerOptions (builder.ts:1465): CompilerOptions handed to getOwnKeys<T>(map: MapLike<T>)
// (core.ts:1297), T inferred from its index signature, which reads it by for...in and an own-key test and returns a
// fresh array. The caller reads each option by its key on its own CompilerOptions.
interface MapLike<T> {
    [index: string]: T;
}
interface SourceFileLike {
    fileName: string;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    strict?: boolean;
    outDir?: string;
    lib?: string[];
    target?: number;
    readonly configFile?: SourceFileLike;
    [option: string]: CompilerOptionsValue | SourceFileLike | undefined;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;

function getOwnKeys<T>(map: MapLike<T>): string[] {
    const keys: string[] = [];
    for (const key in map) {
        if (hasOwnProperty.call(map, key)) {
            keys.push(key);
        }
    }
    return keys;
}
function text(value: CompilerOptionsValue | SourceFileLike): string {
    if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
        return `${typeof value} ${String(value)}`;
    }
    if (value === null || value === undefined) {
        return value === null ? "null" : "undefined";
    }
    return typeof value;
}
function toBuildInfo(options: CompilerOptions): string[] {
    const result: string[] = [];
    for (const name of getOwnKeys(options).sort((a, b) => (a < b ? -1 : a > b ? 1 : 0))) {
        if (name !== "configFile") {
            result.push(`${name}=${text(options[name])}`);
        }
    }
    return result;
}

const declared: CompilerOptions = { target: 9, strict: false, outDir: "out", lib: ["es2022"], configFile: { fileName: "tsconfig.json" } };
console.log(getOwnKeys(declared).join(","));
console.log(toBuildInfo(declared).join(" "));
console.log(`${declared.strict}:${declared.target}:${declared.configFile === undefined ? "none" : declared.configFile.fileName}`);
console.log(`${toBuildInfo({}).length}`);
