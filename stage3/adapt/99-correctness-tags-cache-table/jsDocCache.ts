import type { JSDoc, JSDocTag } from "./types.js";

const cache = new Map<JSDoc[], readonly JSDocTag[]>();

export function getCachedJSDocTags(comments: JSDoc[]): readonly JSDocTag[] | undefined {
    return cache.get(comments);
}

export function setCachedJSDocTags(comments: JSDoc[], tags: readonly JSDocTag[]): void {
    cache.set(comments, tags);
}

export function clearCachedJSDocTags(comments: JSDoc[]): void {
    cache.delete(comments);
}
