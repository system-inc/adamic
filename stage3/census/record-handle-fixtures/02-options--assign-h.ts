// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type Fn = (path: string) => string;
const fn: Fn = path => path;
const k = "k";
interface PluginImport {
    name: string;
}
interface ProjectReference {
    /** A normalized path on disk */
    path: string;
    /** The path as the user originally wrote it */
    originalPath?: string;
    /** @deprecated */
    prepend?: boolean;
    /** True if it is intended that this reference form a circularity */
    circular?: boolean;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined; // eslint-disable-line no-restricted-syntax
interface TsConfigSourceFile { kind: number; pos: number; end: number; }
interface CompilerOptions {
 all?: boolean;
 allowJs?: boolean;
 strict?: boolean;
 outDir?: string;
 rootDirs?: string[];
 paths?: MapLike<string[]>;
 plugins?: PluginImport[];
 readonly configFile?: TsConfigSourceFile;
 [option: string]: CompilerOptionsValue | TsConfigSourceFile | undefined;
}

type Holding = CompilerOptions;
const h: Holding = {};
Object.assign(h, { [k]: fn });
const receiver = h;
const value = receiver.k;
if (value === undefined) throw new Error("missing k");
const read: Fn = value;
console.log(read("called"));
