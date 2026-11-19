// #6c4swvt piece 4, Language's q06: the way back through a field. A holder whose readonly options field is a
// readonly view, seen as one whose options is writable, is refused as the readonly-to-writable field.
type Value = string | number | boolean | undefined;
interface Source {
    fileName: string;
}
interface WatchOptions {
    watchFile?: number;
    [option: string]: Value;
}
interface OptionsBase {
    [option: string]: Value | Source;
}
interface Held {
    readonly options: Readonly<OptionsBase>;
}
interface Writable {
    options: OptionsBase;
}
const watch: WatchOptions = { watchFile: 4 };
const held: Held = { options: watch };
const writable: Writable = held;
writable.options["k"] = { fileName: "x" };
console.log(typeof watch["k"]);
