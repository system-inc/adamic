// @types/node 25.3.3, buffer.d.ts and buffer.buffer.d.ts, selected members.
// Lowering admits only the input forms and literal encodings in the census.
declare module 'node:buffer' {
    export type WithImplicitCoercion<T> =
        | T
        | { valueOf(): T }
        | (T extends string ? { [Symbol.toPrimitive](hint: "string"): T } : never);
    export type BufferEncoding =
        | "ascii" | "utf8" | "utf-8" | "utf16le" | "utf-16le"
        | "ucs2" | "ucs-2" | "base64" | "base64url" | "latin1" | "binary" | "hex";
    export interface Buffer<TArrayBuffer extends ArrayBufferLike = ArrayBufferLike> extends Uint8Array<TArrayBuffer> {
        toString(encoding?: BufferEncoding, start?: number, end?: number): string;
    }
    export interface BufferConstructor {
        from(array: WithImplicitCoercion<ArrayLike<number>>): Buffer<ArrayBuffer>;
        from(string: WithImplicitCoercion<string>, encoding?: BufferEncoding): Buffer<ArrayBuffer>;
        from(arrayOrString: WithImplicitCoercion<ArrayLike<number> | string>): Buffer<ArrayBuffer>;
    }
    export var Buffer: BufferConstructor;
}
