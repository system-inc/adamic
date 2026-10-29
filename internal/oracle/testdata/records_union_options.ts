// tsc's CompilerOptions in miniature: declared options over an index signature whose union holds null, undefined, a
// dictionary and a fixed object, built through computed keys as commandLineParser builds it and through declared keys.
interface MapLike<T> {
    [index: string]: T;
}
interface SourceFileLike {
    fileName: string;
}
type OptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    outFile?: string;
    sourceMap?: boolean;
    target?: number;
    strict?: boolean;
    noImplicitAny?: boolean;
    paths?: MapLike<string[]>;
    configFile?: SourceFileLike;
    [option: string]: OptionsValue | SourceFileLike | undefined;
}
type StrictOptionName = "noImplicitAny" | "strict";

function strictOption(options: CompilerOptions, flag: StrictOptionName): boolean {
    return options[flag] === undefined ? options.strict !== false : !!options[flag];
}
function describe(options: CompilerOptions): string {
    return `${options.outFile ?? "none"}:${options.sourceMap ? "map" : "nomap"}:${options.target ?? -1}:${strictOption(options, "noImplicitAny")}`;
}
function presence(options: CompilerOptions, key: string): string {
    const value = options[key];
    return `${key} ${key in options} ${value === null} ${value === undefined} ${typeof value} ${typeof value === "object" ? "-" : String(value)}`;
}

const outFile: string = ["out", "File"].join("");
const rootDir: string = ["root", "Dir"].join("");
const declarationDir: string = ["declaration", "Dir"].join("");
const parsed: CompilerOptions = {
    [outFile]: ["out", "js"].join("."),
    sourceMap: true,
    [["tar", "get"].join("")]: 99,
    [rootDir]: null,
    [declarationDir]: undefined,
    paths: { "*": ["types/*"] },
    configFile: { fileName: "tsconfig.json" },
    lib: ["es2022", "dom"],
};
const declared: CompilerOptions = { outFile: "a.js", strict: false, noImplicitAny: true };

console.log(describe(parsed));
console.log(describe(declared));
console.log(describe({}));
console.log(`${strictOption(declared, "strict")}:${strictOption(parsed, "strict")}`);
for (const key of [rootDir, declarationDir, "missing", outFile, "sourceMap", "lib"]) {
    console.log(presence(parsed, key));
}
console.log(`${typeof parsed.paths}:${typeof parsed.configFile}:${parsed.configFile === undefined}`);
console.log(Object.keys(parsed).join(","));
console.log(Object.keys(declared).join(","));
const visited: string[] = [];
for (const key in parsed) {
    visited.push(key);
}
console.log(visited.join(","));
console.log(`${delete parsed[rootDir]}:${rootDir in parsed}:${parsed[rootDir] === undefined}`);
console.log(Object.keys(parsed).join(","));
