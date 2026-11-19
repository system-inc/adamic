// #6c4swvt's record-to-fixed direction one field deep, tsc's commandLineParser.ts:2129: a parse result whose
// options are held as a record returned as one whose options are Options. Its reads agree with Node, and the lie
// under strict stops at its read.
type Value = string | number | boolean | undefined;
interface OptionsBase {
    [option: string]: Value;
}
interface Options {
    strict?: boolean;
    outDir?: string;
    [option: string]: Value;
}
interface BaseParsed {
    options: OptionsBase;
    fileNames: string[];
}
interface Parsed {
    options: Options;
    fileNames: string[];
}
function worker(strict: Value): BaseParsed {
    const options: OptionsBase = {};
    options["strict"] = strict;
    options["outDir"] = "out";
    return { options, fileNames: ["a.ts"] };
}
function parse(strict: Value): Parsed {
    return worker(strict);
}
const honest = parse(true);
console.log(String(honest.options.strict) + ":" + String(honest.options.outDir) + ":" + honest.fileNames.join(","));
const lie = parse("yes");
console.log("before");
console.log(String(lie.options.strict));
