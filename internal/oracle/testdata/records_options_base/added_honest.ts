// #6c4swvt's record-to-fixed direction: a record seen as Options, which declares strict and outDir that the
// record's own values don't promise. Each named read through the view is checked by its declared type, so the
// view adds them (addedContractRead, recordAddedNameChecked): what the record holds reads as Node reads it, and a
// name it doesn't hold reads undefined.
type Value = string | number | boolean | undefined;
interface Options {
    strict?: boolean;
    outDir?: string;
    target?: number;
    [option: string]: Value;
}
function parsed(): Record<string, Value> {
    const bag: Record<string, Value> = {};
    bag["strict"] = true;
    bag["outDir"] = "out";
    return bag;
}
const options: Options = parsed();
console.log(String(options.strict) + ":" + String(options.outDir) + ":" + String(options.target));
options.target = 9;
console.log(String(options.target) + ":" + Object.keys(options).join(","));
