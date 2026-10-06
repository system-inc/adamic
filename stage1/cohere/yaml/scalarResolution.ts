import type { Token } from './cst.ts';
import { ComposeError } from './composeError.ts';
export class ScalarResolution {
    value = '';
    type = '';
    comment = '';
    range: number[] = [];
    errors: ComposeError[] = [];
    error(offset: number, code: string, message: string): void {
        this.errors.push(new ComposeError(offset, offset + 1, code, message));
    }
    tokenError(token: Token, code: string, message: string): void {
        this.errors.push(new ComposeError(token.offset, token.offset + token.source.length, code, message));
    }
}
