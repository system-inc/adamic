export declare const versionMajorMinor = "6.0";
/** The version of the TypeScript compiler release */
export declare const version: string;
/**
 * Type of objects whose values are all of the same type.
 * The `in` and `for-in` operators can *not* be safely used,
 * since `Object.prototype` may be modified by outside code.
 */
export interface MapLike<T> {
    [index: string]: T;
}
export interface SortedReadonlyArray<T> extends ReadonlyArray<T> {
    " __sortedArrayBrand": undefined;
}
export interface SortedArray<T> extends Array<T> {
    " __sortedArrayBrand": undefined;
}
