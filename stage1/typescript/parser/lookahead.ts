// Speculation saves and restores scanner state without building a tree.
import type { Scanner } from '../scanner/scanner.ts';
import { precedence } from './grammar.ts';

export interface ParserStateInterface {
    readonly pos: number;
    readonly start: number;
    readonly fullStart: number;
    readonly kind: string;
    readonly value: string;
    readonly flags: number;
    readonly errors: number;
    readonly nodes: number;
}

function kind(scanner: Scanner): string {
    return scanner.kind;
}
class Speculation {
    readonly scanner: Scanner;
    readonly pos: number;
    readonly start: number;
    readonly fullStart: number;
    readonly kind: string;
    readonly value: string;
    readonly flags: number;
    readonly errors: number;
    constructor(scanner: Scanner) {
        this.scanner = scanner;
        this.pos = scanner.pos;
        this.start = scanner.start;
        this.fullStart = scanner.fullStart;
        this.kind = scanner.kind;
        this.value = scanner.value;
        this.flags = scanner.flags;
        this.errors = scanner.errors.length;
    }
    restore(): void {
        this.scanner.pos = this.pos;
        this.scanner.start = this.start;
        this.scanner.fullStart = this.fullStart;
        this.scanner.kind = this.kind;
        this.scanner.value = this.value;
        this.scanner.flags = this.flags;
        this.scanner.errors.splice(this.errors);
    }
}

export function arrowAhead(scanner: Scanner, allowReturn: boolean): boolean {
    const state = new Speculation(scanner);
    let result = false;
    if(kind(scanner) === 'AsyncKeyword') {
        scanner.scan();
        if((scanner.flags & 1) !== 0) {
            state.restore();
            return false;
        }
        if(kind(scanner) === 'Identifier') {
            scanner.scan();
            result = kind(scanner) === 'EqualsGreaterThanToken' && (scanner.flags & 1) === 0;
            state.restore();
            return result;
        }
    }
    if(kind(scanner) === 'LessThanToken') {
        let depth = 0;
        while(kind(scanner) !== 'EndOfFile') {
            if(kind(scanner) === 'LessThanToken') {
                depth++;
            }
            if(kind(scanner) === 'GreaterThanToken') {
                depth--;
            }
            scanner.scan();
            if(depth === 0) {
                break;
            }
        }
    }
    if(kind(scanner) === 'OpenParenToken') {
        let depth = 0;
        let typed = false;
        let head = 0;
        while(kind(scanner) !== 'EndOfFile') {
            if(kind(scanner) === 'OpenParenToken') {
                depth++;
            }
            if(kind(scanner) === 'CloseParenToken') {
                depth--;
            }
            if(depth === 1 && kind(scanner) !== 'OpenParenToken') {
                head++;
                if(
                    (head === 2 && (kind(scanner) === 'ColonToken' || kind(scanner) === 'QuestionToken')) ||
                    (head === 1 && kind(scanner) === 'DotDotDotToken')
                ) {
                    typed = true;
                }
            }
            scanner.scan();
            if(depth === 0) {
                break;
            }
        }
        if(kind(scanner) === 'EqualsGreaterThanToken' && (scanner.flags & 1) === 0) {
            result = true;
        }
        else if(kind(scanner) === 'ColonToken' && (allowReturn || typed)) {
            scanner.scan();
            let braces = 0;
            while(kind(scanner) !== 'EndOfFile' && kind(scanner) !== 'EqualsGreaterThanToken') {
                if(kind(scanner) === 'SemicolonToken' && braces === 0) {
                    break;
                }
                if(kind(scanner) === 'OpenBraceToken') {
                    braces++;
                }
                if(kind(scanner) === 'CloseBraceToken') {
                    if(braces === 0) {
                        break;
                    }
                    braces--;
                }
                scanner.scan();
            }
            result = kind(scanner) === 'EqualsGreaterThanToken';
        }
    }
    state.restore();
    return result;
}

export function typeArgumentsAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    let depth = 0;
    let braces = 0;
    let parens = 0;
    let brackets = 0;
    let closed = false;
    let conditionalType = false;
    while(kind(scanner) !== 'EndOfFile') {
        if(
            (kind(scanner) === 'CloseParenToken' && parens === 0) ||
            (kind(scanner) === 'CloseBraceToken' && braces === 0) ||
            (kind(scanner) === 'CloseBracketToken' && brackets === 0)
        ) {
            break;
        }
        if(kind(scanner) === 'ExtendsKeyword') {
            conditionalType = true;
        }
        if(kind(scanner) === 'QuestionToken' && !conditionalType && braces === 0 && parens === 0 && brackets === 0) {
            break;
        }
        if(kind(scanner) === 'OpenBracketToken') {
            brackets++;
        }
        if(kind(scanner) === 'CloseBracketToken') {
            brackets--;
        }
        if(kind(scanner) === 'SemicolonToken' && braces === 0 && parens === 0) {
            break;
        }
        if(kind(scanner) === 'OpenBraceToken') {
            braces++;
        }
        if(kind(scanner) === 'CloseBraceToken') {
            braces--;
        }
        if(kind(scanner) === 'OpenParenToken') {
            parens++;
        }
        if(kind(scanner) === 'CloseParenToken') {
            parens--;
        }
        if(kind(scanner) === 'LessThanToken') {
            depth++;
        }
        if(kind(scanner) === 'GreaterThanToken') {
            depth--;
        }
        if(
            kind(scanner) === 'PlusToken' ||
            kind(scanner) === 'MinusToken' ||
            kind(scanner) === 'SlashToken' ||
            kind(scanner) === 'BarBarToken' ||
            kind(scanner) === 'AmpersandAmpersandToken'
        ) {
            break;
        }
        scanner.scan();
        if(depth === 0) {
            closed = true;
            break;
        }
    }
    const after = kind(scanner);
    const result =
        closed &&
        (after === 'OpenParenToken' ||
            after === 'NoSubstitutionTemplateLiteral' ||
            after === 'TemplateHead' ||
            (after !== 'LessThanToken' &&
                after !== 'GreaterThanToken' &&
                after !== 'PlusToken' &&
                after !== 'MinusToken' &&
                ((scanner.flags & 1) !== 0 ||
                    precedence(after) >= 0 ||
                    after === 'SemicolonToken' ||
                    after === 'CloseParenToken' ||
                    after === 'CloseBracketToken' ||
                    after === 'CommaToken' ||
                    after === 'EndOfFile' ||
                    after === 'DotToken' ||
                    after === 'EqualsToken')));
    state.restore();
    return result;
}
