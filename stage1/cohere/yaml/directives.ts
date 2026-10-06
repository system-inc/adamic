import type { Token } from './cst.ts';
import { ComposeError } from './composeError.ts';
function parts(text: string): string[] {
    const result: string[] = [];
    const source = text.trim();
    let start = 0;
    for(let index = 0; index < source.length; index++) {
        const character = source.slice(index, index + 1);
        if(character !== ' ' && character !== '\t') continue;
        result.push(source.slice(start, index));
        while(source.slice(index, index + 1) === ' ' || source.slice(index, index + 1) === '\t') index++;
        start = index;
        index--;
    }
    result.push(source.slice(start));
    return result;
}
function versionPattern(text: string): boolean {
    const groups = text.split('.');
    if(groups.length !== 2) return false;
    for(const group of groups) {
        if(group === '') return false;
        for(let index = 0; index < group.length; index++) {
            const code = group.charCodeAt(index);
            if(code < 48 || code > 57) return false;
        }
    }
    return true;
}
function hexByte(source: string, index: number): number {
    if(source.slice(index, index + 1) !== '%' || index + 2 >= source.length) return -1;
    const high = '0123456789abcdef'.indexOf(source.slice(index + 1, index + 2).toLowerCase());
    const low = '0123456789abcdef'.indexOf(source.slice(index + 2, index + 3).toLowerCase());
    if(high < 0 || low < 0) return -1;
    return high * 16 + low;
}
// ECMAScript decodeURIComponent, including rejection of overlong and surrogate UTF-8.
function decode(source: string): string[] {
    const result: string[] = [];
    for(let index = 0; index < source.length; index++) {
        if(source.slice(index, index + 1) !== '%') {
            result.push(source.slice(index, index + 1));
            continue;
        }
        const first = hexByte(source, index);
        if(first < 0) return [];
        index += 2;
        if(first < 128) {
            result.push(String.fromCodePoint(first));
            continue;
        }
        const length =
            first >= 194 && first <= 223 ? 2 : first >= 224 && first <= 239 ? 3 : first >= 240 && first <= 244 ? 4 : 0;
        if(length === 0) return [];
        let code = first & (length === 2 ? 31 : length === 3 ? 15 : 7);
        for(let count = 1; count < length; count++) {
            index++;
            const octet = hexByte(source, index);
            if(octet < 128 || octet > 191) return [];
            code = code * 64 + (octet & 63);
            index += 2;
        }
        if(
            code < (length === 2 ? 128 : length === 3 ? 2048 : 65536) ||
            code > 0x10ffff ||
            (code >= 0xd800 && code <= 0xdfff)
        )
            return [];
        result.push(String.fromCodePoint(code));
    }
    // A final empty part distinguishes valid empty text from malformed text.
    result.push('');
    return result;
}
export class Directives {
    docStart = false;
    docEnd = false;
    explicit = false;
    version = '1.2';
    tags = new Map<string, string>();
    atNextDocument = false;
    errors: ComposeError[] = [];
    warnings: ComposeError[] = [];
    constructor() {
        this.tags.set('!!', 'tag:yaml.org,2002:');
    }
    atDocument(): Directives {
        const result = new Directives();
        result.explicit = this.explicit;
        result.version = this.version;
        for(const [handle, prefix] of this.tags) result.tags.set(handle, prefix);
        if(this.version === '1.1') this.atNextDocument = true;
        else {
            this.atNextDocument = false;
            this.explicit = false;
            this.version = '1.2';
            this.tags = new Map<string, string>();
            this.tags.set('!!', 'tag:yaml.org,2002:');
        }
        return result;
    }
    error(token: Token, relative: number, message: string, warning: boolean): void {
        const error = new ComposeError(
            token.offset + relative,
            token.offset + token.source.length,
            'BAD_DIRECTIVE',
            message,
        );
        if(warning) this.warnings.push(error);
        else this.errors.push(error);
    }
    add(token: Token): void {
        if(this.atNextDocument) {
            this.explicit = false;
            this.version = '1.1';
            this.tags = new Map<string, string>();
            this.tags.set('!!', 'tag:yaml.org,2002:');
            this.atNextDocument = false;
        }
        const words = parts(token.source);
        const name = words[0] ?? '';
        if(name === '%TAG') {
            if(words.length !== 3) {
                this.error(token, 0, '%TAG directive should contain exactly two parts', false);
                if(words.length < 3) return;
            }
            this.tags.set(words[1] ?? '', words[2] ?? '');
        }
        else if(name === '%YAML') {
            this.explicit = true;
            if(words.length !== 2) {
                this.error(token, 0, '%YAML directive should contain exactly one part', false);
                return;
            }
            const version = words[1] ?? '';
            if(version === '1.1' || version === '1.2') this.version = version;
            else this.error(token, 6, `Unsupported YAML version ${version}`, versionPattern(version));
        }
        else this.error(token, 0, `Unknown directive ${name}`, true);
    }
    tagError(token: Token, message: string): void {
        this.errors.push(
            new ComposeError(token.offset, token.offset + token.source.length, 'TAG_RESOLVE_FAILED', message),
        );
    }
    tagName(token: Token): string {
        if(token.source === '!') return '!';
        if(!token.source.startsWith('!')) {
            this.tagError(token, `Not a valid tag: ${token.source}`);
            return '';
        }
        if(token.source.slice(1, 2) === '<') {
            const verbatim = token.source.slice(2, -1);
            if(verbatim === '!' || verbatim === '!!') {
                this.tagError(token, `Verbatim tags aren't resolved, so ${token.source} is invalid.`);
                return '';
            }
            if(!token.source.endsWith('>')) this.tagError(token, 'Verbatim tags must end with a >');
            return verbatim;
        }
        const last = token.source.lastIndexOf('!');
        const handle = token.source.slice(0, last + 1);
        const suffix = token.source.slice(last + 1);
        if(suffix === '') this.tagError(token, `The ${token.source} tag has no suffix`);
        const prefix = this.tags.get(handle) ?? '';
        if(prefix !== '') {
            const decoded = decode(suffix);
            if(decoded.length === 0) {
                this.tagError(token, 'URIError: URI malformed');
                return '';
            }
            return prefix + decoded.join('');
        }
        if(handle === '!') return token.source;
        this.tagError(token, `Could not resolve tag: ${token.source}`);
        return '';
    }
}
