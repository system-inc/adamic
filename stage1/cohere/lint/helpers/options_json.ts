// Strict encoding/json grammar. Flat indices keep the value arena acyclic.
import { panic } from 'adamic';
export interface OptionValue {
    readonly kind: string;
    readonly text: string;
    readonly number: number;
    readonly keys: readonly string[];
    readonly children: readonly number[];
}
export class OptionsJson {
    readonly nodes: OptionValue[] = [];
    position = 0;
    root = -1;
    error = '';
    readonly source: string;
    constructor(source: string) { this.source = source; }
    node(index: number): OptionValue { return this.nodes[index] ?? panic('missing option value'); }
    add(kind: string, text: string, number: number, keys: string[], children: number[]): number {
        const index = this.nodes.length;
        this.nodes.push({kind, text, number, keys, children});
        return index;
    }
    skip(): void {
        while(this.position < this.source.length && ' \t\r\n'.includes(this.source.slice(this.position, this.position + 1))) { this.position++; }
    }
    string(): string {
        this.position++;
        let result = '';
        while(this.position < this.source.length) {
            const char = this.source.slice(this.position, this.position + 1);
            this.position++;
            if(char === '"') {
                // Go replaces unpaired UTF-16 escape halves with U+FFFD.
                let normalized = '';
                for(let i = 0; i < result.length; i++) {
                    const code = result.charCodeAt(i); const next = result.charCodeAt(i + 1);
                    if(code >= 0xd800 && code <= 0xdbff && next >= 0xdc00 && next <= 0xdfff) { normalized += result.slice(i, i + 2); i++; }
                    else { normalized += code >= 0xd800 && code <= 0xdfff ? '\ufffd' : result.slice(i, i + 1); }
                }
                return normalized;
            }
            if(char.charCodeAt(0) < 32) { this.error = 'control character'; return ''; }
            if(char !== '\\') { result += char; continue; }
            const escape = this.source.slice(this.position, this.position + 1);
            this.position++;
            if(escape === 'u') {
                const digits = this.source.slice(this.position, this.position + 4);
                if(digits.length !== 4 || ![...digits].every(c => '0123456789abcdefABCDEF'.includes(c))) { this.error = 'unicode escape'; return ''; }
                result += String.fromCharCode(Number.parseInt(digits, 16));
                this.position += 4;
            } else {
                const names = ['"', '\\', '/', 'b', 'f', 'n', 'r', 't'];
                const texts = ['"', '\\', '/', '\b', '\f', '\n', '\r', '\t'];
                const index = names.indexOf(escape);
                if(index < 0) { this.error = 'escape'; return ''; }
                result += texts[index] ?? panic('missing escape');
            }
        }
        this.error = 'unterminated string'; return '';
    }
    digit(): boolean {
        const char = this.source.slice(this.position, this.position + 1);
        return char !== '' && '0123456789'.includes(char);
    }
    value(depth: number): number {
        if(depth > 512) { this.error = 'option nesting exceeds 512'; return -1; }
        this.skip();
        const char = this.source.slice(this.position, this.position + 1);
        if(char === '"') { const text = this.string(); return this.add('string', text, 0, [], []); }
        if(char === '[' || char === '{') {
            const object = char === '{'; const close = object ? '}' : ']';
            this.position++; this.skip();
            const keys: string[] = []; const children: number[] = [];
            if(this.source.slice(this.position, this.position + 1) === close) { this.position++; return this.add(object ? 'object' : 'array', '', 0, keys, children); }
            while(this.error === '') {
                let key = '';
                if(object) {
                    if(this.source.slice(this.position, this.position + 1) !== '"') { this.error = 'object key'; break; }
                    key = this.string(); this.skip();
                    if(this.source.slice(this.position, this.position + 1) !== ':') { this.error = 'colon'; break; }
                    this.position++;
                }
                const value = this.value(depth + 1);
                if(this.error !== '') { break; }
                // encoding/json keeps the last duplicate key.
                const previous = object ? keys.indexOf(key) : -1;
                if(previous >= 0) { children[previous] = value; }
                else { if(object) { keys.push(key); } children.push(value); }
                this.skip();
                const separator = this.source.slice(this.position, this.position + 1); this.position++;
                if(separator === close) { return this.add(object ? 'object' : 'array', '', 0, keys, children); }
                if(separator !== ',') { this.error = 'separator'; break; }
                this.skip();
            }
            return -1;
        }
        for(const keyword of ['true', 'false', 'null']) {
            if(this.source.slice(this.position, this.position + keyword.length) === keyword) {
                this.position += keyword.length; return this.add(keyword === 'null' ? 'null' : 'boolean', keyword, 0, [], []);
            }
        }
        const start = this.position;
        if(char === '-') { this.position++; }
        if(this.source.slice(this.position, this.position + 1) === '0') { this.position++; }
        else {
            if(!this.digit()) { this.error = 'value'; return -1; }
            while(this.digit()) { this.position++; }
        }
        if(this.source.slice(this.position, this.position + 1) === '.') {
            this.position++; if(!this.digit()) { this.error = 'fraction'; return -1; }
            while(this.digit()) { this.position++; }
        }
        const exponent = this.source.slice(this.position, this.position + 1);
        if(exponent === 'e' || exponent === 'E') {
            this.position++;
            const sign = this.source.slice(this.position, this.position + 1);
            if(sign === '+' || sign === '-') { this.position++; }
            if(!this.digit()) { this.error = 'exponent'; return -1; }
            while(this.digit()) { this.position++; }
        }
        const text = this.source.slice(start, this.position);
        return this.add('number', text, Number.parseFloat(text), [], []);
    }
    parse(): number {
        const root = this.value(0); this.skip();
        if(this.position !== this.source.length) { this.error = 'trailing JSON'; }
        this.root = this.error === '' ? root : -1; return this.root;
    }
    field(index: number, name: string): number {
        const value = this.node(index); const offset = value.keys.indexOf(name);
        return offset < 0 ? -1 : value.children[offset] ?? panic('missing option member');
    }
    equal(left: number, other: OptionsJson, right: number): boolean {
        const a = this.node(left); const b = other.node(right);
        if(a.kind !== b.kind) { return false; }
        if(a.kind === 'number') { return Number.isFinite(a.number) && Number.isFinite(b.number) && a.number === b.number; }
        if(a.kind === 'string' || a.kind === 'boolean' || a.kind === 'null') { return a.text === b.text; }
        if(a.children.length !== b.children.length) { return false; }
        for(let i = 0; i < a.children.length; i++) {
            const j = a.kind === 'object' ? b.keys.indexOf(a.keys[i] ?? panic('missing key')) : i;
            if(j < 0 || !this.equal(a.children[i] ?? panic('missing value'), other, b.children[j] ?? panic('missing value'))) { return false; }
        }
        return true;
    }
}
