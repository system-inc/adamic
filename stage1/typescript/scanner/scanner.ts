// Port of cohere/TypeScript/tsc/internal/scanner/scanner.go.
// Positions inside the scanner are UTF-16; byte offsets are supplied by the driver.
import { panic } from 'adamic';
import { digit, isIdentifierPart, isIdentifierStart, isLineBreak, isSpace } from './characters.ts';
import { keywords, punctuators } from './tokens.ts';

export interface ScanErrorInterface {
    readonly code: number;
    readonly start: number;
    readonly length: number;
}

// Binary and octal bigint spellings need arbitrary precision, but only text, not bigint values.
function decimalDigits(text: string, base: number): string {
    let result = '0';
    for(const character of text) {
        let carry = digit(character.charCodeAt(0));
        let next = '';
        for(let index = result.length - 1; index >= 0; index--) {
            const value = (result.charCodeAt(index) - 48) * base + carry;
            next = String.fromCharCode(48 + (value % 10)) + next;
            carry = Math.floor(value / 10);
        }
        while(carry > 0) {
            next = String.fromCharCode(48 + (carry % 10)) + next;
            carry = Math.floor(carry / 10);
        }
        result = next;
    }
    let first = 0;
    while(first < result.length - 1 && result.charCodeAt(first) === 48) {
        first++;
    }
    return result.slice(first);
}

function numberValue(text: string): string {
    if(text === '') {
        return '0';
    }
    const prefix = text.slice(0, 2);
    const value =
        prefix === '0x'
            ? Number.parseInt(text.slice(2), 16)
            : prefix === '0b'
              ? Number.parseInt(text.slice(2), 2)
              : prefix === '0o'
                ? Number.parseInt(text.slice(2), 8)
                : Number.parseFloat(text);
    return `${value}`;
}

export class Scanner {
    readonly text: string;
    pos = 0;
    start = 0;
    fullStart = 0;
    kind = 'Unknown';
    value = '';
    flags = 0;
    readonly errors: ScanErrorInterface[] = [];
    constructor(text: string) {
        this.text = text;
    }
    code(offset = 0): number {
        const unit = this.text.charCodeAt(this.pos + offset);
        // Most reads need one BMP unit. Only a high surrogate can start a pair.
        if(unit >= 0xd800 && unit <= 0xdbff) {
            return this.text.codePointAt(this.pos + offset) ?? -1;
        }
        return Number.isNaN(unit) ? -1 : unit;
    }
    advance(code: number): void {
        this.pos += code > 0xffff ? 2 : 1;
    }
    error(code: number, start: number = this.pos, length = 0): void {
        this.errors.push({ code, start, length });
    }

    // Decimal fragments set ContainsInvalidSeparator; radix fragments only report errors.
    fragment(base: number, separators: boolean, minimum = 0, maximum: number = Number.MAX_SAFE_INTEGER): string {
        let result = '';
        let allowed = false;
        let previous = false;
        let count = 0;
        while(count < maximum) {
            const code = this.code();
            const value = digit(code);
            if(value >= 0 && value < base) {
                result += String.fromCharCode(code).toLowerCase();
                count++;
                allowed = separators;
                previous = false;
            }
            else if(separators && code === 95) {
                this.flags |= 512;
                if(allowed) {
                    allowed = false;
                    previous = true;
                }
                else {
                    if(base === 10) {
                        this.flags |= 16384;
                    }
                    this.error(previous ? 6189 : 6188, this.pos, 1);
                }
            }
            else {
                break;
            }
            this.pos++;
        }
        if(previous) {
            if(base === 10) {
                this.flags |= 16384;
            }
            this.error(6188, this.pos - 1, 1);
        }
        return count < minimum ? '' : result;
    }

    unicode(report: boolean): number {
        this.pos += 2;
        const start = this.pos;
        const extended = this.code() === 123;
        if(extended) {
            this.pos++;
        }
        else {
            this.flags |= 1024;
        }
        const digits = this.fragment(16, false, extended ? 1 : 4, extended ? Number.MAX_SAFE_INTEGER : 4);
        if(digits === '') {
            this.flags |= 2048;
            if(report) {
                this.error(1125);
            }
            return -1;
        }
        const value = Math.min(Number.parseInt(digits, 16), 0x7fffffff);
        if(extended) {
            let invalid = false;
            if(value > 0x10ffff) {
                if(report) {
                    this.error(1198, start + 1, this.pos - start - 1);
                }
                invalid = true;
            }
            if(this.pos >= this.text.length) {
                if(report) {
                    this.error(1126);
                }
                invalid = true;
            }
            else if(this.code() === 125) {
                this.pos++;
            }
            else {
                if(report) {
                    this.error(1199);
                }
                invalid = true;
            }
            if(invalid) {
                this.flags |= 2048;
                return -1;
            }
            this.flags |= 8;
        }
        return value;
    }

    escape(report: boolean): string {
        const start = this.pos;
        this.pos++;
        const code = this.code();
        if(code < 0) {
            this.error(1126);
            return '';
        }
        this.advance(code);
        if(code >= 48 && code <= 55) {
            if(code === 48 && !(this.code() >= 48 && this.code() <= 57)) {
                return '\u0000';
            }
            if(code <= 51 && this.code() >= 48 && this.code() <= 55) {
                this.pos++;
            }
            if(this.code() >= 48 && this.code() <= 55) {
                this.pos++;
            }
            this.flags |= 2048;
            if(report) {
                this.error(1487, start, this.pos - start);
                return String.fromCharCode(Number.parseInt(this.text.slice(start + 1, this.pos), 8));
            }
            return this.text.slice(start, this.pos);
        }
        if(code === 56 || code === 57) {
            this.flags |= 2048;
            if(report) {
                this.error(1488, start, this.pos - start);
                return String.fromCharCode(code);
            }
            return this.text.slice(start, this.pos);
        }
        if(code === 117) {
            this.pos = start;
            const point = this.unicode(report);
            return point < 0 ? this.text.slice(start, this.pos) : String.fromCodePoint(point);
        }
        if(code === 120) {
            const digits = this.fragment(16, false, 2, 2);
            if(digits === '') {
                this.flags |= 2048;
                if(report) {
                    this.error(1125);
                }
                return this.text.slice(start, this.pos);
            }
            this.flags |= 4096;
            return String.fromCharCode(Number.parseInt(digits, 16));
        }
        switch(code) {
            case 98:
                return '\b';
            case 116:
                return '\t';
            case 110:
                return '\n';
            case 118:
                return '\v';
            case 102:
                return '\f';
            case 114:
                return '\r';
            case 13:
                if(this.code() === 10) {
                    this.pos++;
                }
                return '';
            case 10:
            case 0x2028:
            case 0x2029:
                return '';
            default:
                return String.fromCodePoint(code);
        }
    }

    identifier(prefix: number): boolean {
        const start = this.pos;
        this.pos += prefix;
        let result = '';
        let segment = start;
        let escaped = false;
        let first = true;
        for(;;) {
            let code = this.code();
            if(code === 92 && this.code(1) === 117) {
                const saved = this.pos;
                const flags = this.flags;
                code = this.unicode(false);
                if(!(first ? isIdentifierStart(code) : isIdentifierPart(code))) {
                    this.pos = saved;
                    this.flags = flags;
                    break;
                }
                result += this.text.slice(segment, saved);
                result += String.fromCodePoint(code);
                segment = this.pos;
                escaped = true;
            }
            else if(first ? isIdentifierStart(code) : isIdentifierPart(code)) {
                this.advance(code);
            }
            else {
                break;
            }
            first = false;
        }
        if(first) {
            return false;
        }
        this.value = escaped ? result + this.text.slice(segment, this.pos) : this.text.slice(start, this.pos);
        return true;
    }

    string(): void {
        const quote = this.text.charCodeAt(this.pos);
        if(quote === 39) {
            this.flags |= 65536;
        }
        this.pos++;
        let start = this.pos;
        const pieces: string[] = [];
        for(;;) {
            const code = this.text.charCodeAt(this.pos);
            if(this.pos >= this.text.length || code === 10 || code === 13 || code === quote) {
                pieces.push(this.text.slice(start, this.pos));
                if(code === quote) {
                    this.pos++;
                }
                else {
                    this.flags |= 4;
                    this.error(1002);
                }
                break;
            }
            if(code === 92) {
                // Gap 1: push accepts one value per call in stage 0.
                pieces.push(this.text.slice(start, this.pos));
                pieces.push(this.escape(true));
                start = this.pos;
            }
            else {
                this.pos++;
            }
        }
        this.kind = 'StringLiteral';
        this.value = pieces.length === 1 ? (pieces[0] ?? panic('missing string piece')) : pieces.join('');
    }

    template(report: boolean): void {
        const backtick = this.text.charCodeAt(this.pos) === 96;
        this.pos++;
        let start = this.pos;
        const pieces: string[] = [];
        for(;;) {
            const code = this.text.charCodeAt(this.pos);
            if(this.pos >= this.text.length || code === 96) {
                pieces.push(this.text.slice(start, this.pos));
                if(code === 96) {
                    this.pos++;
                }
                else {
                    this.flags |= 4;
                    this.error(1160);
                }
                this.kind = backtick ? 'NoSubstitutionTemplateLiteral' : 'TemplateTail';
                break;
            }
            if(code === 36 && this.text.charCodeAt(this.pos + 1) === 123) {
                pieces.push(this.text.slice(start, this.pos));
                this.pos += 2;
                this.kind = backtick ? 'TemplateHead' : 'TemplateMiddle';
                break;
            }
            if(code === 92) {
                pieces.push(this.text.slice(start, this.pos));
                pieces.push(this.escape(report));
                start = this.pos;
            }
            else if(code === 13) {
                pieces.push(this.text.slice(start, this.pos));
                pieces.push('\n');
                this.pos++;
                if(this.text.charCodeAt(this.pos) === 10) {
                    this.pos++;
                }
                start = this.pos;
            }
            else {
                this.pos++;
            }
        }
        this.value = pieces.length === 1 ? (pieces[0] ?? panic('missing string piece')) : pieces.join('');
    }

    bigint(): void {
        if(this.code() === 110) {
            if((this.flags & 384) !== 0) {
                this.value = decimalDigits(this.value.slice(2), (this.flags & 128) !== 0 ? 2 : 8);
            }
            this.value += 'n';
            this.pos++;
            this.kind = 'BigIntLiteral';
        }
        else {
            this.value = numberValue(this.value);
            this.kind = 'NumericLiteral';
        }
    }

    number(): void {
        const start = this.pos;
        let fixed: string;
        if(this.code() === 48) {
            this.pos++;
            if(this.code() === 95) {
                this.flags |= 512 | 16384;
                this.error(6188, this.pos, 1);
                this.pos = start;
                fixed = this.fragment(10, true);
            }
            else {
                const digits = this.fragment(10, false);
                if(digits === '') {
                    fixed = '0';
                }
                else {
                    let octal = true;
                    for(const ch of digits) {
                        if(ch === '8' || ch === '9') {
                            octal = false;
                        }
                    }
                    if(octal) {
                        // Go ParseInt saturates at MaxInt64 for long legacy octal literals.
                        const exact = decimalDigits(digits, 8);
                        const saturated =
                            exact.length > 19 || (exact.length === 19 && exact > '9223372036854775807')
                                ? '9223372036854775807'
                                : exact;
                        this.value = saturated;
                        this.flags |= 32;
                        this.error(
                            1121,
                            this.kind === 'MinusToken' ? start - 1 : start,
                            this.pos - start + (this.kind === 'MinusToken' ? 1 : 0),
                        );
                        this.kind = 'NumericLiteral';
                        return;
                    }
                    this.flags |= 8192;
                    fixed = digits;
                }
            }
        }
        else {
            fixed = this.fragment(10, true);
        }
        const fixedEnd = this.pos;
        let fraction = '';
        let preamble = '';
        let exponent = '';
        if(this.code() === 46) {
            this.pos++;
            fraction = this.fragment(10, true);
        }
        let end = this.pos;
        if(this.code() === 69 || this.code() === 101) {
            this.pos++;
            this.flags |= 16;
            if(this.code() === 43 || this.code() === 45) {
                this.pos++;
            }
            const numericStart = this.pos;
            exponent = this.fragment(10, true);
            if(exponent === '') {
                this.error(1124);
            }
            else {
                preamble = this.text.slice(end, numericStart);
                end = this.pos;
            }
        }
        this.value =
            (this.flags & 512) !== 0
                ? fixed + (fraction !== '' ? `.${fraction}` : '') + (exponent !== '' ? preamble + exponent : '')
                : this.text.slice(start, end);
        if((this.flags & 8192) !== 0) {
            this.error(1489, start, this.pos - start);
            this.value = numberValue(this.value);
            this.kind = 'NumericLiteral';
            return;
        }
        if(fixedEnd === this.pos) {
            this.bigint();
        }
        else {
            this.value = numberValue(this.value);
            this.kind = 'NumericLiteral';
        }
        if(isIdentifierStart(this.code())) {
            const idStart = this.pos;
            const tokenValue = this.value;
            this.identifier(0);
            const idEnd = this.pos;
            const id = this.value;
            this.value = tokenValue;
            if(this.kind !== 'BigIntLiteral' && id === 'n' && ((this.flags & 16) !== 0 || fixedEnd < idStart)) {
                this.error((this.flags & 16) !== 0 ? 1352 : 1353, start, idEnd - start);
            }
            else {
                this.error(1351, idStart, idEnd - idStart);
                this.pos = idStart;
            }
        }
    }

    scan(): string {
        this.fullStart = this.pos;
        this.flags = 0;
        for(;;) {
            this.start = this.pos;
            const code = this.code();
            if(code < 0) {
                this.kind = 'EndOfFile';
                return this.kind;
            }
            if(isSpace(code)) {
                this.advance(code);
                continue;
            }
            if(isLineBreak(code)) {
                this.flags |= 1;
                this.advance(code);
                continue;
            }
            if(code === 47 && this.code(1) === 47) {
                this.pos += 2;
                while(this.pos < this.text.length) {
                    const current = this.text.charCodeAt(this.pos);
                    if(isLineBreak(current)) {
                        break;
                    }
                    this.pos++;
                }
                continue;
            }
            if(code === 47 && this.code(1) === 42) {
                this.pos += 2;
                const jsdoc = this.code() === 42 && this.code(1) !== 47;
                while(this.pos < this.text.length) {
                    const current = this.text.charCodeAt(this.pos);
                    if(current === 42 && this.text.charCodeAt(this.pos + 1) === 47) {
                        break;
                    }
                    if(isLineBreak(current)) {
                        this.flags |= 1;
                    }
                    this.pos++;
                }
                if(this.pos === this.text.length) {
                    this.error(1010);
                }
                else {
                    this.pos += 2;
                }
                if(jsdoc) {
                    this.flags |= 2;
                    const comment = this.text.slice(this.start, this.pos);
                    const parts = comment.split('@');
                    for(const part of parts.slice(1)) {
                        for(const tag of ['deprecated', 'see', 'link', 'linkcode', 'linkplain']) {
                            const ending = part.charCodeAt(tag.length);
                            if(
                                part.startsWith(tag) &&
                                (part.length === tag.length ||
                                    ending === 32 ||
                                    ending === 9 ||
                                    ending === 10 ||
                                    ending === 13 ||
                                    ending === 125 ||
                                    ending === 42)
                            ) {
                                this.flags |= tag === 'deprecated' ? 131072 : 262144;
                            }
                        }
                    }
                }
                continue;
            }
            if(code === 35 && this.code(1) === 33) {
                this.pos += 2;
                if(this.start === 0) {
                    while(this.pos < this.text.length) {
                        const current = this.text.charCodeAt(this.pos);
                        if(isLineBreak(current)) {
                            break;
                        }
                        this.pos++;
                    }
                    continue;
                }
                this.error(18026, this.start, 2);
                this.kind = 'Unknown';
                return this.kind;
            }
            if(
                (code === 60 || code === 62 || code === 61 || code === 124) &&
                (this.start === 0 || isLineBreak(this.text.charCodeAt(this.start - 1))) &&
                this.text.slice(this.start, this.start + 7) === String.fromCharCode(code).repeat(7) &&
                this.start + 7 < this.text.length &&
                (code === 61 || this.code(7) === 32)
            ) {
                this.error(1185, this.pos, 7);
                this.pos += 7;
                if(code === 60 || code === 62) {
                    while(this.pos < this.text.length) {
                        const current = this.text.charCodeAt(this.pos);
                        if(isLineBreak(current)) {
                            break;
                        }
                        this.pos++;
                    }
                }
                else {
                    while(this.pos < this.text.length) {
                        const current = this.text.charCodeAt(this.pos);
                        if(
                            (current === 61 || current === 62) &&
                            current !== code &&
                            (this.pos === 0 || isLineBreak(this.text.charCodeAt(this.pos - 1))) &&
                            this.text.slice(this.pos, this.pos + 7) === String.fromCharCode(current).repeat(7) &&
                            (current === 61 || this.code(7) === 32)
                        ) {
                            break;
                        }
                        this.pos++;
                    }
                }
                continue;
            }
            if(code === 39 || code === 34) {
                this.string();
                return this.kind;
            }
            if(code === 96) {
                this.template(false);
                return this.kind;
            }
            if((code >= 48 && code <= 57) || (code === 46 && this.code(1) >= 48 && this.code(1) <= 57)) {
                const prefix = this.text.slice(this.pos, this.pos + 2).toLowerCase();
                if(prefix === '0x' || prefix === '0b' || prefix === '0o') {
                    this.pos += 2;
                    const base = prefix === '0x' ? 16 : prefix === '0b' ? 2 : 8;
                    let digits = this.fragment(base, true, 1);
                    if(digits === '') {
                        this.error(base === 16 ? 1125 : base === 2 ? 1177 : 1178);
                        digits = '0';
                    }
                    this.value = prefix + digits;
                    this.flags |= base === 16 ? 64 : base === 2 ? 128 : 256;
                    this.bigint();
                }
                else {
                    this.number();
                }
                return this.kind;
            }
            if(code === 35) {
                if(!this.identifier(1)) {
                    this.error(1127, this.pos - 1, 1);
                    this.value = '#';
                }
                this.kind = 'PrivateIdentifier';
                return this.kind;
            }
            if((isIdentifierStart(code) || code === 92) && this.identifier(0)) {
                this.kind = keywords.get(this.value) ?? 'Identifier';
                return this.kind;
            }
            if(code === 0xfffd) {
                this.error(1490, 0);
                this.pos = this.text.length;
                this.kind = 'NonTextFileMarkerTrivia';
                return this.kind;
            }
            const spelling = this.punctuation(code);
            const kind = punctuators.get(spelling);
            if(kind !== undefined) {
                this.pos += spelling.length;
                this.kind = kind;
                return kind;
            }
            this.error(1127, this.pos, code > 0xffff ? 2 : 1);
            this.advance(code);
            this.kind = 'Unknown';
            return this.kind;
        }
    }

    // Character dispatch avoids allocating speculative slices for every punctuation token.
    punctuation(code: number): string {
        switch(code) {
            case 33:
                if(this.code(1) === 61 && this.code(2) === 61) {
                    return '!==';
                }
                if(this.code(1) === 61) {
                    return '!=';
                }
                return '!';
            case 37:
                if(this.code(1) === 61) {
                    return '%=';
                }
                return '%';
            case 38:
                if(this.code(1) === 38 && this.code(2) === 61) {
                    return '&&=';
                }
                if(this.code(1) === 61) {
                    return '&=';
                }
                if(this.code(1) === 38) {
                    return '&&';
                }
                return '&';
            case 42:
                if(this.code(1) === 42 && this.code(2) === 61) {
                    return '**=';
                }
                if(this.code(1) === 61) {
                    return '*=';
                }
                if(this.code(1) === 42) {
                    return '**';
                }
                return '*';
            case 43:
                if(this.code(1) === 61) {
                    return '+=';
                }
                if(this.code(1) === 43) {
                    return '++';
                }
                return '+';
            case 45:
                if(this.code(1) === 61) {
                    return '-=';
                }
                if(this.code(1) === 45) {
                    return '--';
                }
                return '-';
            case 46:
                if(this.code(1) === 46 && this.code(2) === 46) {
                    return '...';
                }
                return '.';
            case 47:
                if(this.code(1) === 61) {
                    return '/=';
                }
                return '/';
            case 60:
                if(this.code(1) === 60 && this.code(2) === 61) {
                    return '<<=';
                }
                if(this.code(1) === 61) {
                    return '<=';
                }
                if(this.code(1) === 60) {
                    return '<<';
                }
                return '<';
            case 61:
                if(this.code(1) === 61 && this.code(2) === 61) {
                    return '===';
                }
                if(this.code(1) === 61) {
                    return '==';
                }
                if(this.code(1) === 62) {
                    return '=>';
                }
                return '=';
            case 62:
                return '>';
            case 63:
                if(this.code(1) === 63 && this.code(2) === 61) {
                    return '??=';
                }
                if(this.code(1) === 63) {
                    return '??';
                }
                if(this.code(1) === 46 && !(this.code(2) >= 48 && this.code(2) <= 57)) {
                    return '?.';
                }
                return '?';
            case 94:
                if(this.code(1) === 61) {
                    return '^=';
                }
                return '^';
            case 124:
                if(this.code(1) === 124 && this.code(2) === 61) {
                    return '||=';
                }
                if(this.code(1) === 61) {
                    return '|=';
                }
                if(this.code(1) === 124) {
                    return '||';
                }
                return '|';
            case 126:
                return '~';
            case 64:
                return '@';
            case 40:
                return '(';
            case 41:
                return ')';
            case 91:
                return '[';
            case 93:
                return ']';
            case 123:
                return '{';
            case 125:
                return '}';
            case 59:
                return ';';
            case 58:
                return ':';
            case 44:
                return ',';
            default:
                return '';
        }
    }

    rescanGreater(): void {
        if(this.kind !== 'GreaterThanToken') {
            return;
        }
        for(let length = 4; length >= 1; length--) {
            const spelling = this.text.slice(this.start, this.start + length);
            if(spelling.length !== length) {
                continue;
            }
            const kind = punctuators.get(spelling);
            if(kind !== undefined) {
                this.kind = kind;
                this.pos = this.start + length;
                return;
            }
        }
    }
    rescanTemplate(): void {
        this.pos = this.start;
        this.template(true);
    }

    // Boundary scan with diagnostics disabled, as ReScanSlashToken(false). Unterminated errors remain.
    rescanSlash(): void {
        if(this.kind !== 'SlashToken' && this.kind !== 'SlashEqualsToken') {
            return;
        }
        const body = this.start + 1;
        let position = body;
        let escaped = false;
        let characterClass = false;
        for(;;) {
            const code = this.text.charCodeAt(position);
            if(position >= this.text.length || code === 10 || code === 13) {
                this.flags |= 4;
                break;
            }
            if(escaped) {
                escaped = false;
            }
            else if(code === 47 && !characterClass) {
                break;
            }
            else if(code === 91) {
                characterClass = true;
            }
            else if(code === 92) {
                escaped = true;
            }
            else if(code === 93) {
                characterClass = false;
            }
            position++;
        }
        if((this.flags & 4) !== 0) {
            const end = position;
            position = body;
            escaped = false;
            let classes = 0;
            let groups = 0;
            let quantifier = false;
            while(position < end) {
                const code = this.text.charCodeAt(position);
                if(escaped) {
                    escaped = false;
                }
                else if(code === 92) {
                    escaped = true;
                }
                else if(code === 91) {
                    classes++;
                }
                else if(code === 93 && classes !== 0) {
                    classes--;
                }
                else if(classes === 0) {
                    if(code === 123) {
                        quantifier = true;
                    }
                    else if(code === 125 && quantifier) {
                        quantifier = false;
                    }
                    else if(!quantifier) {
                        if(code === 40) {
                            groups++;
                        }
                        else if(code === 41 && groups !== 0) {
                            groups--;
                        }
                        else if(code === 41 || code === 93 || code === 125) {
                            break;
                        }
                    }
                }
                position++;
            }
            while(position > body) {
                const code = this.text.charCodeAt(position - 1);
                if(isSpace(code) || isLineBreak(code) || code === 59) {
                    position--;
                }
                else {
                    break;
                }
            }
            this.error(1161, this.start, position - this.start);
        }
        else {
            position++;
            while(position < this.text.length && isIdentifierPart(this.text.codePointAt(position) ?? -1)) {
                position += (this.text.codePointAt(position) ?? -1) > 0xffff ? 2 : 1;
            }
        }
        this.pos = position;
        this.value = this.text.slice(this.start, position);
        this.kind = 'RegularExpressionLiteral';
    }

    // Extend the already scanned identifier with JSX's dash name parts.
    scanJsxIdentifier(): void {
        if(this.kind !== 'Identifier' && !this.kind.endsWith('Keyword') && this.kind !== 'PrivateIdentifier') {
            return;
        }
        const start = this.pos;
        while(this.code() === 45 || isIdentifierPart(this.code())) {
            this.advance(this.code());
        }
        this.value += this.text.slice(start, this.pos);
        this.kind = keywords.get(this.value) ?? 'Identifier';
    }
    // Attribute strings preserve backslashes, entities and line endings.
    scanJsxAttributeValue(): void {
        this.fullStart = this.pos;
        while(isSpace(this.code()) || isLineBreak(this.code())) {
            this.advance(this.code());
        }
        this.start = this.pos;
        const quote = this.code();
        if(quote !== 34 && quote !== 39) {
            this.scan();
            return;
        }
        if(quote === 39) {
            this.flags |= 65536;
        }
        this.pos++;
        const start = this.pos;
        while(this.pos < this.text.length && this.code() !== quote) {
            this.advance(this.code());
        }
        this.value = this.text.slice(start, this.pos);
        if(this.pos === this.text.length) {
            this.flags |= 4;
            this.error(1002);
        }
        else {
            this.pos++;
        }
        this.kind = 'StringLiteral';
    }

    scanJsx(): string {
        this.start = this.pos;
        this.fullStart = this.pos;
        const code = this.code();
        if(code < 0) {
            this.kind = 'EndOfFile';
        }
        else if(code === 60) {
            this.pos += this.code(1) === 47 ? 2 : 1;
            this.kind = this.text.charCodeAt(this.start + 1) === 47 ? 'LessThanSlashToken' : 'LessThanToken';
        }
        else if(code === 123) {
            this.pos++;
            this.kind = 'OpenBraceToken';
        }
        else {
            let first = 0;
            while(this.pos < this.text.length && this.code() !== 60 && this.code() !== 123) {
                const current = this.code();
                if(current === 62 || current === 125) {
                    this.error(current === 62 ? 1382 : 1381, this.pos, 1);
                }
                if(isLineBreak(current) && first === 0) {
                    first = -1;
                }
                else if(!isSpace(current) && !isLineBreak(current)) {
                    first = this.pos;
                }
                this.advance(current);
            }
            this.value = this.text.slice(this.start, this.pos);
            this.kind = first !== -1 ? 'JsxTextAllWhiteSpaces' : 'JsxText';
        }
        return this.kind;
    }
}
