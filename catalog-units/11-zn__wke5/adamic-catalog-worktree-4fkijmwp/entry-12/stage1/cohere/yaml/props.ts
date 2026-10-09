// Property token resolution from yaml 2.9.0; missing token links are -1.
import type { ComposeError } from './composeError.ts';
export class Props {
    warnings: ComposeError[] = [];
    comma = -1;
    found = -1;
    spaceBefore = false;
    comment = '';
    hasNewline = false;
    anchor = -1;
    tag = -1;
    newlineAfterProp = -1;
    end = 0;
    start = 0;
}
