// #6c4swvt's record-to-fixed direction, a lie: the record holds a string under strict, which Options declares
// boolean. Node reads it; both backends stop at the view's named read, after the line before it.
type Value = string | number | boolean | undefined;
interface Options {
    strict?: boolean;
    [option: string]: Value;
}
function parsed(): Record<string, Value> {
    const bag: Record<string, Value> = {};
    bag["strict"] = "yes";
    return bag;
}
const options: Options = parsed();
console.log("before");
console.log(String(options.strict));
