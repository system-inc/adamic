// @types/node 25.3.3 crypto.d.ts, selected members and their encoding aliases.
// BinaryLike is restricted to the census's string input; stream options and
// non-string ArrayBufferView inputs belong to a later host surface.
declare module 'node:crypto' {
    export type BinaryToTextEncoding = "base64" | "base64url" | "hex" | "binary";
    export type CharacterEncoding = "utf8" | "utf-8" | "utf16le" | "utf-16le" | "latin1";
    export type LegacyCharacterEncoding = "ascii" | "binary" | "ucs2" | "ucs-2";
    export type Encoding = BinaryToTextEncoding | CharacterEncoding | LegacyCharacterEncoding;
    export type BinaryLike = string;
    export interface Hash {
        update(data: BinaryLike): Hash;
        update(data: string, inputEncoding: Encoding): Hash;
        digest(encoding: BinaryToTextEncoding): string;
    }
    export function createHash(algorithm: string): Hash;
}
