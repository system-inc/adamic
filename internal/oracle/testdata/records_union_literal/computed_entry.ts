// #6c4swvt (c): Language's g5 p06 shape, reachable on main without a view. A record literal's computed entry puts an
// object literal under paths, declared MapLike<string[]>, which recordComputedEntry leaves to the named read's checked
// narrow to the slot union's one dictionary. The literal's slot is the index signature's union, which holds that
// dictionary and no fixed object the literal fits, so it's built as the dictionary, and the read agrees with Node.
// Built fixed, the narrow stopped natively on an honest value.
interface MapLike<T> {
    [index: string]: T;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | null | undefined;
interface CompilerOptions {
    paths?: MapLike<string[]>;
    outDir?: string;
    [option: string]: CompilerOptionsValue | undefined;
}
function pathsKey(): string {
    return "paths";
}
const options: CompilerOptions = { outDir: "out", [pathsKey()]: { k: ["a", "b"] } };
const paths = options.paths;
if (paths) {
    const list = paths["k"];
    console.log(list ? String(list.length) + ":" + (list[1] ?? "none") : "missing");
    console.log(paths["z"] ? "present" : "missing");
}
