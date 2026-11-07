// Cohere's port of postcss 8.5.16 lib/tokenize.js.
import { panic } from 'adamic';
import type { Input } from './input.ts';

export class Token {
    readonly kind: string;
    readonly value: string;
    readonly start: number;
    readonly end: number;
    readonly inline: boolean;
    constructor(kind: string, value: string, start: number, end: number = -1, inline: boolean = false) {
        this.kind = kind;
        this.value = value;
        this.start = start;
        this.end = end;
        this.inline = inline;
    }
    endOrStart(): number {
        return this.end > 0 ? this.end : this.start;
    }
}
function space(code: number): boolean {
    return code === 32 || code === 10 || code === 9 || code === 13 || code === 12;
}
function hex(code: number): boolean {
    return (code >= 48 && code <= 57) || (code >= 65 && code <= 70) || (code >= 97 && code <= 102);
}
export class Tokenizer {
    readonly input: Input;
    readonly scss: boolean;
    pos = 0;
    readonly buffer: Token[] = [];
    readonly returned: Token[] = [];
    lastBadParen = -1;
    constructor(input: Input, scss: boolean = false) {
        this.input = input;
        this.scss = scss;
    }
    code(index: number): number {
        return index < 0 || index >= this.input.css.length ? -1 : this.input.css.charCodeAt(index);
    }
    indexOf(text: string, from: number): number {
        // The prelude only has indexOf(search). Search a suffix and add its origin.
        from = Math.max(from, 0);
        const found = this.input.css.slice(from).indexOf(text);
        return found < 0 ? -1 : from + found;
    }
    back(token: Token): void {
        this.returned.push(token);
    }
    endOfFile(): boolean {
        return this.returned.length === 0 && this.pos >= this.input.css.length;
    }
    interpolation(next: number): number {
        let deep = 1;
        let stringQuote = 0;
        let escaped = false;
        while(deep > 0) {
            next++;
            if(next >= this.input.css.length) {
                throw new Error(this.input.error('Unclosed interpolation', this.pos));
            }
            const code = this.code(next);
            if(stringQuote !== 0) {
                if(!escaped && code === stringQuote) {
                    stringQuote = 0;
                    escaped = false;
                }
                else if(code === 92) {
                    escaped = !escaped;
                }
                else if(escaped) {
                    escaped = false;
                }
            }
            else if(code === 39 || code === 34) {
                stringQuote = code;
            }
            else if(code === 125) {
                deep--;
            }
            else if(code === 35 && this.code(next + 1) === 123) {
                deep++;
            }
        }
        return next;
    }
    nextToken(): Token | undefined {
        if(this.returned.length > 0) {
            return this.returned.pop() ?? panic('missing returned token');
        }
        const length = this.input.css.length;
        const pos = this.pos;
        if(pos >= length) {
            return undefined;
        }
        let code = this.code(pos);
        let next = pos;
        let kind = 'word';
        let start = pos;
        let end = -1;
        let inline = false;
        if(space(code)) {
            do {
                next++;
            } while(space(this.code(next)));
            kind = 'space';
            start = -1;
            next--;
        }
        else if(
            code === 91 ||
            code === 93 ||
            code === 123 ||
            code === 125 ||
            code === 58 ||
            code === 59 ||
            code === 41
        ) {
            kind = this.input.slice(pos, pos + 1);
        }
        else if(this.scss && code === 44) {
            end = pos + 1;
        }
        else if(code === 40) {
            const previous = this.buffer.length > 0 ? (this.buffer.pop() ?? panic('missing buffered token')).value : '';
            const n = this.code(pos + 1);
            if(this.scss && previous === 'url' && n !== 39 && n !== 34) {
                let brackets = 1;
                next = pos + 1;
                while(next < length) {
                    const character = this.code(next);
                    if(character === 40) {
                        brackets++;
                    }
                    else if(character === 41) {
                        brackets--;
                        if(brackets === 0) {
                            break;
                        }
                    }
                    next++;
                }
                kind = 'brackets';
                end = next;
            }
            else if(!this.scss && previous === 'url' && n !== 39 && n !== 34 && !space(n)) {
                do {
                    next = this.indexOf(')', next + 1);
                    if(next < 0) {
                        throw new Error(this.input.error('Unclosed bracket', pos));
                    }
                    let escaped = false;
                    let escapePos = next;
                    while(this.code(escapePos - 1) === 92) {
                        escapePos--;
                        escaped = !escaped;
                    }
                    if(!escaped) {
                        break;
                    }
                } while(true);
                kind = 'brackets';
                end = next;
            }
            else if(!this.scss && pos <= this.lastBadParen) {
                kind = '(';
            }
            else {
                next = this.indexOf(')', pos + 1);
                if(next < 0 || /.[\r\n"'(/\\]/.test(this.input.slice(pos, next + 1))) {
                    this.lastBadParen = next < 0 ? length : next;
                    kind = '(';
                    next = pos;
                }
                else {
                    kind = 'brackets';
                    end = next;
                }
            }
        }
        else if(code === 39 || code === 34) {
            if(this.scss) {
                const quoteCode = code;
                let escaped = false;
                while(next < length) {
                    next++;
                    if(next === length) {
                        throw new Error(this.input.error('Unclosed string', pos));
                    }
                    const character = this.code(next);
                    if(!escaped && character === quoteCode) {
                        break;
                    }
                    else if(character === 92) {
                        escaped = !escaped;
                    }
                    else if(escaped) {
                        escaped = false;
                    }
                    else if(character === 35 && this.code(next + 1) === 123) {
                        next = this.interpolation(next);
                    }
                }
                kind = 'string';
                end = next;
                this.pos = next + 1;
                return new Token(kind, this.input.slice(pos, next + 1), start, end);
            }
            const quote = code === 39 ? "'" : '"';
            do {
                next = this.indexOf(quote, next + 1);
                if(next < 0) {
                    throw new Error(this.input.error('Unclosed string', pos));
                }
                let escaped = false;
                let escapePos = next;
                while(this.code(escapePos - 1) === 92) {
                    escapePos--;
                    escaped = !escaped;
                }
                if(!escaped) {
                    break;
                }
            } while(true);
            kind = 'string';
            end = next;
        }
        else if(code === 64) {
            next = pos + 1;
            while(next < length && !/[\t\n\f\r "#'()/;[\\\]{}]/.test(this.input.slice(next, next + 1))) {
                next++;
            }
            next--;
            kind = 'at-word';
            end = next;
        }
        else if(code === 92) {
            let escape = true;
            while(this.code(next + 1) === 92) {
                next++;
                escape = !escape;
            }
            code = this.code(next + 1);
            if(escape && code !== 47 && !space(code)) {
                next++;
                if(hex(this.code(next))) {
                    while(hex(this.code(next + 1))) {
                        next++;
                    }
                    if(this.code(next + 1) === 32) {
                        next++;
                    }
                }
            }
            end = next;
        }
        else if(this.scss && code === 35 && this.code(pos + 1) === 123) {
            next = this.interpolation(pos);
            end = next;
        }
        else if(this.scss && code === 47 && this.code(pos + 1) === 47) {
            next = pos + 1;
            while(next < length && this.code(next) !== 10 && this.code(next) !== 12 && this.code(next) !== 13) {
                next++;
            }
            next--;
            end = next;
            kind = 'comment';
            inline = true;
        }
        else if(code === 47 && this.code(pos + 1) === 42) {
            next = this.indexOf('*/', pos + 2) + 1;
            if(next === 0) {
                throw new Error(this.input.error('Unclosed comment', pos));
            }
            kind = 'comment';
            end = next;
        }
        else {
            next = pos + 1;
            while(
                next < length &&
                !(this.scss && this.code(next) === 44) &&
                !/[\t\n\f\r !"#'():;@[\\\]{}]/.test(this.input.slice(next, next + 1)) &&
                !(this.code(next) === 47 && this.code(next + 1) === 42)
            ) {
                next++;
            }
            next--;
            end = next;
            this.buffer.push(new Token(kind, this.input.slice(pos, next + 1), start, end));
        }
        this.pos = next + 1;
        return new Token(kind, this.input.slice(pos, next + 1), start, end, inline);
    }
}
