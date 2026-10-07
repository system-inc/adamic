// Cohere's tokenize.go, dist/tokenize.js, walking the library's UTF-16 units.
export class Token {
    readonly kind: string;
    readonly value: string;
    readonly rest: readonly number[];
    constructor(kind: string, value: string, rest: readonly number[]) {
        this.kind = kind;
        this.value = value;
        this.rest = rest;
    }
    at(index: number): number | undefined {
        return this.rest[index - 2];
    }
}

function whitespace(code: number): boolean {
    return code === 10 || code === 32 || code === 9 || code === 13 || code === 12;
}
function combinator(code: number): boolean {
    return code === 43 || code === 62 || code === 126 || code === 124;
}
function atEnd(code: number): boolean {
    return (
        code === 32 ||
        code === 10 ||
        code === 9 ||
        code === 13 ||
        code === 123 ||
        code === 40 ||
        code === 41 ||
        code === 39 ||
        code === 34 ||
        code === 92 ||
        code === 59 ||
        code === 47
    );
}
function wordEnd(css: string, index: number): boolean {
    const code = css.charCodeAt(index);
    return (
        code === 32 ||
        code === 10 ||
        code === 9 ||
        code === 13 ||
        code === 40 ||
        code === 41 ||
        code === 42 ||
        code === 58 ||
        code === 59 ||
        code === 64 ||
        code === 33 ||
        code === 38 ||
        code === 39 ||
        code === 34 ||
        combinator(code) ||
        code === 44 ||
        code === 91 ||
        code === 93 ||
        code === 92 ||
        (code === 47 && css.charCodeAt(index + 1) === 42)
    );
}

export function tokenize(css: string): Token[] {
    const tokens: Token[] = [];
    let offset = -1;
    let line = 1;
    for(let pos = 0; pos < css.length; pos++) {
        let code = css.charCodeAt(pos);
        if(code === 10) {
            offset = pos;
            line++;
        }
        let next = pos;
        if(whitespace(code)) {
            do {
                next++;
                code = css.charCodeAt(next);
                if(code === 10) {
                    offset = next;
                    line++;
                }
            } while(whitespace(code));
            tokens.push(new Token('space', css.slice(pos, next), [line, pos - offset, pos]));
            pos = next - 1;
        }
        else if(combinator(code)) {
            do {
                next++;
            } while(combinator(css.charCodeAt(next)));
            tokens.push(new Token('combinator', css.slice(pos, next), [line, pos - offset, pos]));
            pos = next - 1;
        }
        else if(
            code === 42 ||
            code === 38 ||
            code === 44 ||
            code === 91 ||
            code === 93 ||
            code === 58 ||
            code === 59 ||
            code === 40 ||
            code === 41
        ) {
            const value = css.slice(pos, pos + 1);
            tokens.push(new Token(value, value, [line, pos - offset, pos]));
        }
        else if(code === 39 || code === 34) {
            const quote = css.slice(pos, pos + 1);
            while(true) {
                next = css.indexOf(quote, next + 1);
                if(next === -1) {
                    throw new Error('Unclosed quote');
                }
                let escapePos = next;
                let escaped = false;
                while(css.charCodeAt(escapePos - 1) === 92) {
                    escapePos--;
                    escaped = !escaped;
                }
                if(!escaped) {
                    break;
                }
            }
            tokens.push(new Token('string', css.slice(pos, next + 1), [line, pos - offset, line, next - offset, pos]));
            pos = next;
        }
        else if(code === 64) {
            next = pos + 1;
            while(next < css.length && !atEnd(css.charCodeAt(next))) {
                next++;
            }
            next--;
            tokens.push(new Token('at-word', css.slice(pos, next + 1), [line, pos - offset, line, next - offset, pos]));
            pos = next;
        }
        else if(code === 92) {
            let escape = true;
            while(css.charCodeAt(next + 1) === 92) {
                next++;
                escape = !escape;
            }
            code = css.charCodeAt(next + 1);
            if(escape && code !== 47 && !whitespace(code)) {
                next++;
            }
            tokens.push(new Token('word', css.slice(pos, next + 1), [line, pos - offset, line, next - offset, pos]));
            pos = next;
        }
        else if(code === 47 && css.charCodeAt(pos + 1) === 42) {
            next = css.indexOf('*/', pos + 2) + 1;
            if(next === 0) {
                throw new Error('Unclosed comment');
            }
            const content = css.slice(pos, next + 1);
            const lines = content.split('\n');
            const last = lines.length - 1;
            const nextLine = line + last;
            const nextOffset = last > 0 ? next - (lines[last] ?? '').length : offset;
            tokens.push(new Token('comment', content, [line, pos - offset, nextLine, next - nextOffset, pos]));
            offset = nextOffset;
            line = nextLine;
            pos = next;
        }
        else {
            next = pos + 1;
            while(next < css.length && !wordEnd(css, next)) {
                next++;
            }
            next--;
            tokens.push(new Token('word', css.slice(pos, next + 1), [line, pos - offset, line, next - offset, pos]));
            pos = next;
        }
    }
    return tokens;
}
