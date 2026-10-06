import type { Directives } from './directives.ts';
import type { ComposeError } from './composeError.ts';
import { ScalarResolution } from './scalarResolution.ts';
export class ComposedDocument {
    directives: Directives;
    contents = -1;
    range: number[] = [];
    comment = '';
    commentBefore = '';
    diagnostics = new ScalarResolution();
    warnings: ComposeError[] = [];
    constructor(directives: Directives) {
        this.directives = directives;
    }
}
