// tsc's handleOptionConfigDirTemplateSubstitution (commandLineParser.ts:3278), called at 3095 with a CompilerOptions
// seen as OptionsBase | undefined: it reads options[name], writes only into a fresh copy, assign({}, options), and
// returns that copy or, when nothing was substituted, the object it was handed, the same object. tsc's caller casts
// it back as CompilerOptions (3098), which the cast rule refuses (back_cast.ts), so this caller reads the result
// through OptionsBase, and the CompilerOptions it handed by name.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    strict?: boolean;
    outDir?: string;
    rootDir?: string;
    lib?: string[];
    target?: number;
    configFilePath?: string | undefined;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;
const configDir = "${configDir}";

function assign(target: OptionsBase, source: OptionsBase): OptionsBase {
    for (const p in source) {
        if (hasOwnProperty.call(source, p)) {
            target[p] = source[p];
        }
    }
    return target;
}
function handleOptionConfigDirTemplateSubstitution(options: OptionsBase | undefined, names: readonly string[], basePath: string): OptionsBase | undefined {
    if (!options) return options;
    let result: OptionsBase | undefined;
    for (const name of names) {
        const value = options[name];
        if (typeof value === "string" && value.startsWith(configDir)) {
            if (result === undefined) {
                result = assign({}, options);
            }
            result[name] = basePath + value.slice(configDir.length);
        }
    }
    return result || options;
}
function text(value: CompilerOptionsValue): string {
    return typeof value === "string" || typeof value === "boolean" || typeof value === "number" ? String(value) : value === undefined ? "undefined" : "other";
}
function parse(existing: CompilerOptions, fileName: string): string {
    existing.configFilePath = fileName;
    const options = handleOptionConfigDirTemplateSubstitution(existing, ["outDir", "rootDir"], "/project");
    if (options === undefined) {
        return "none";
    }
    return `${options === existing}:${text(options["outDir"])}:${text(options["rootDir"])}:${text(options["strict"])}:${text(options["configFilePath"])}:${Object.keys(options).join(",")}`;
}

const substituted: CompilerOptions = { outDir: configDir + "/out", rootDir: "src", strict: true, lib: ["es2022"] };
console.log(parse(substituted, "tsconfig.json"));
console.log(`${substituted.outDir}:${substituted.rootDir}:${substituted.strict}:${substituted.configFilePath}`);
const kept: CompilerOptions = { outDir: "out", target: 9 };
console.log(parse(kept, "kept.json"));
console.log(`${kept.outDir}:${kept.target}:${kept.strict}:${kept.configFilePath}`);
console.log(`${handleOptionConfigDirTemplateSubstitution(undefined, ["outDir"], "/project") === undefined}`);
