// #6c4swvt piece 4, Language's q02: the way back through a type parameter. A readonly view cast back to its
// writable constraint's instantiation is refused as the same widening, so no writable alias of the record leaks.
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
function seen(options: Readonly<OptionsBase>): Readonly<OptionsBase> {
    return options;
}
function back<T extends OptionsBase>(options: Readonly<T>): T {
    return options as T;
}
const watch: WatchOptions = { watchFile: 4 };
const alias = back<OptionsBase>(seen(watch));
alias["k"] = { fileName: "x" };
console.log(typeof watch["k"]);
