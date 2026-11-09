// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
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


// The value whose static receiver type is being tested is the record itself.
type Receiver = (path: string) => string;
type Holding = CompilerOptions;
const h: Holding = {};
const erased: {} = h;
const receiver: Receiver = erased;
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("call result:", receiver("called"));
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
