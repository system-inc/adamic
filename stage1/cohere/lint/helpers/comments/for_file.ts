import type { Parser } from '../../../../typescript/parser/parser.ts';
import { all } from './all.ts';
import type { Comment } from './comment.ts';
// One cache object per immutable parsed file, shared by all its rule contexts.
// No global table retains files. An undefined cache keeps Go's uncached contract.
export class FileCommentCache {
    readonly parser: Parser | undefined;
    readonly root: number;
    result: Comment[] = [];
    ready = false;
    constructor(parser: Parser | undefined, root: number) {
        this.parser = parser;
        this.root = root;
    }
    get(): Comment[] {
        if(!this.ready) {
            this.result = all(this.parser, this.root);
            this.ready = true;
        }
        return this.result;
    }
}
export function forFile(parser: Parser | undefined, root: number, cache: FileCommentCache | undefined): Comment[] {
    if(cache === undefined) {
        return all(parser, root);
    }
    return cache.get();
}
