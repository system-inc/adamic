// #6c4swvt piece 4, tsc's commandLineParser.ts:3257 shape: handleOptionConfigDirTemplateSubstitution reads its
// options through Readonly<OptionsBase> and returns the original object when nothing is substituted. The caller's
// cast back to WatchOptions is what would hand the readonly view back as a writable one, and the cast rule
// refuses it, so the view never leaks a writable alias of the record.
type Value = string | number | boolean | string[] | undefined;
interface WatchOptions {
    watchFile?: number;
    [option: string]: Value;
}
interface OptionsBase {
    [option: string]: Value;
}
function substitute(options: Readonly<OptionsBase> | undefined, base: string): Readonly<OptionsBase> | undefined {
    let result: OptionsBase | undefined;
    if (options && options["outDir"] === "${configDir}") {
        result = {};
        result["outDir"] = base;
    }
    return result || options;
}
const watch: WatchOptions = { watchFile: 4 };
const back = substitute(watch, "/b") as WatchOptions | undefined;
console.log(String(back ? back.watchFile : "none"));
