// Language's condition on #6c4swvt (2026-11-06 03:08Z): a named field narrower than the index, strict: boolean under
// CompilerOptionsValue | undefined, is stored at the index's representation, so a read through OptionsBase reads
// its bytes as they're stored, by name and by a computed key, and the record read by name agrees after it.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    strict: boolean;
    target?: number;
    [option: string]: CompilerOptionsValue | undefined;
}
interface OptionsBase {
    [option: string]: CompilerOptionsValue | undefined;
}

function read(options: OptionsBase, name: string): string {
    const value = options[name];
    if (typeof value === "boolean" || typeof value === "number" || typeof value === "string") {
        return `${name} ${typeof value} ${String(value)}`;
    }
    return `${name} ${value === undefined ? "undefined" : "other"}`;
}
function strictThroughView(options: OptionsBase): string {
    const value = options["strict"];
    return typeof value === "boolean" ? `strict ${value ? "on" : "off"}` : "strict missing";
}

const strictName: string = ["str", "ict"].join("");
const on: CompilerOptions = { strict: true, target: 9 };
const off: CompilerOptions = { strict: false };
for (const options of [on, off]) {
    console.log(read(options, strictName));
    console.log(read(options, "target"));
    console.log(strictThroughView(options));
    console.log(`${options.strict}:${options.target}:${options.strict === true}`);
}
