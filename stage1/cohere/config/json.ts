// Typed JSON tree. Child indices keep recursion in data rather than unknown or an unchecked cast.
// Strict syntax diagnostics follow pinned Go encoding/json; JSONC permits comments and trailing commas.
import { goIsPrint } from './quote_table.ts';
import { panic, utf8At } from 'adamic';

export class JsonNode {
	kind = 'null';
	text = '';
	raw = '';
	readonly keys: string[] = [];
	readonly children: number[] = [];
}

export function quoted(text: string): string {
	let result = '"';
	for (const character of text) {
		const point = character.codePointAt(0) ?? 0;
		if (character === '"' || character === '\\') { result += '\\' + character; }
		else if (character === '\n') { result += '\\n'; }
		else if (character === '\r') { result += '\\r'; }
		else if (character === '\t') { result += '\\t'; }
		else if (character === '\b') { result += '\\b'; }
		else if (character === '\f') { result += '\\f'; }
		else if (point === 7) { result += '\\a'; }
		else if (point === 11) { result += '\\v'; }
		else if (point < 32 || point === 127) { result += '\\x' + point.toString(16).padStart(2, '0'); }
		else if (!goIsPrint(point)) { result += (point < 65536 ? '\\u' : '\\U') + point.toString(16).padStart(point < 65536 ? 4 : 8, '0'); }
		else { result += character; }
	}
	return result + '"';
}

function replaceUnpairedSurrogates(text: string): string {
    let result = '';
    for (let index = 0; index < text.length; index++) {
        const unit = text.charCodeAt(index);
        if (unit >= 0xd800 && unit <= 0xdbff) {
            const next = text.charCodeAt(index + 1);
            if (next >= 0xdc00 && next <= 0xdfff) { result += text.slice(index, index + 2); index++; }
            else { result += '�'; }
        } else if (unit >= 0xdc00 && unit <= 0xdfff) { result += '�'; }
        else { result += text[index] ?? ''; }
    }
    return result;
}

export class JsonDocument {
	readonly nodes: JsonNode[] = [];
	error = '';
	position = 0;
	readonly text: string;
	readonly comments: boolean;
	constructor(text: string, comments: boolean) {
		this.text = text;
		this.comments = comments;
	}
	node(index: number): JsonNode { return this.nodes[index] ?? panic('JSON: missing node'); }
	get(index: number, key: string): number {
		if (index < 0) { return -1; }
		const node = this.node(index);
		for (let at = node.keys.length - 1; at >= 0; at--) {
			if (node.keys[at] === key) { return node.children[at] ?? -1; }
		}
		return -1;
	}
	string(index: number): string { return index < 0 || this.node(index).kind === 'null' ? '' : this.node(index).text; }
	kind(index: number): string { return index < 0 ? 'missing' : this.node(index).kind; }
	strings(index: number): string[] {
		const result: string[] = [];
		if (index >= 0 && this.kind(index) === 'array') {
			for (const child of this.node(index).children) { result.push(this.string(child)); }
		}
		return result;
	}
	space(): void {
		while (this.position < this.text.length) {
			const character = this.text[this.position] ?? '';
			if (character === ' ' || character === '\t' || character === '\r' || character === '\n' || (this.comments && character === '\ufeff')) { this.position++; continue; }
			if (this.comments && this.text.slice(this.position, this.position + 2) === '//') {
				while (this.position < this.text.length && this.text[this.position] !== '\n') { this.position++; }
				continue;
			}
			if (this.comments && this.text.slice(this.position, this.position + 2) === '/*') {
				const end = this.text.indexOf('*/', this.position + 2);
				if (end < 0) { this.error = 'unexpected end of JSON input'; return; }
				this.position = end + 2;
				continue;
			}
			break;
		}
	}
	invalid(context: string): void {
		if (this.error !== '') { return; }
		const character = this.text[this.position];
		if (character === undefined) { this.error = 'unexpected end of JSON input'; return; }
		const byte = String.fromCharCode(utf8At(character, 0));
		const escaped = byte === "'" ? "\\'" : quoted(byte).slice(1, -1);
		this.error = `invalid character '${escaped}' ${context}`;
	}
	readString(): string {
		this.position++;
		let result = '';
		while (this.position < this.text.length && this.error === '') {
			const character = this.text[this.position] ?? '';
			if (character === '"') { this.position++; return replaceUnpairedSurrogates(result); }
			if (character.charCodeAt(0) < 32) { this.invalid('in string'); return ''; }
			this.position++;
			if (character !== '\\') { result += character; continue; }
			const escape = this.text[this.position] ?? '';
			const escapeStart = this.position - 1;
			if (escape === 'u') {
				this.position++;
				let code = 0;
				for (let digit = 0; digit < 4; digit++) {
					const character = this.text[this.position] ?? '';
					const value = '0123456789abcdef'.indexOf(character.toLowerCase());
					if (character === '' || value < 0) { this.error = this.text.length < escapeStart + 6 ? 'unexpected end of JSON input' : `invalid escape sequence \`${this.text.slice(escapeStart, escapeStart + 6)}\` in string`; return ''; }
					code = code * 16 + value;
					this.position++;
				}
				result += String.fromCharCode(code);
			} else {
				if (escape === '"' || escape === '\\' || escape === '/') { result += escape; }
				else if (escape === 'n') { result += '\n'; }
				else if (escape === 'r') { result += '\r'; }
				else if (escape === 't') { result += '\t'; }
				else if (escape === 'b') { result += '\b'; }
				else if (escape === 'f') { result += '\f'; }
				else { this.error = escape === '' ? 'unexpected end of JSON input' : `invalid escape sequence \`${this.text.slice(escapeStart, escapeStart + 2)}\` in string`; return ''; }
				this.position++;
			}
		}
		if (this.error === '') { this.error = 'unexpected end of JSON input'; }
		return '';
	}
	value(): number {
		this.space();
		const start = this.position;
		const index = this.nodes.length;
		const node = new JsonNode();
		this.nodes.push(node);
		const first = this.text[this.position] ?? '';
		if (first === '{' || first === '[') {
			node.kind = first === '{' ? 'object' : 'array';
			const end = first === '{' ? '}' : ']';
			this.position++;
			this.space();
			if (this.text[this.position] !== end) {
				while (this.error === '') {
					if (first === '{') {
						if (this.text[this.position] !== '"') { this.invalid('looking for beginning of object key string'); break; }
						node.keys.push(this.readString());
						this.space();
						if (this.text[this.position] !== ':') { this.invalid('after object key'); break; }
						this.position++;
					}
					node.children.push(this.value());
					this.space();
					if (this.text[this.position] === end) { break; }
					if (this.text[this.position] !== ',') { this.invalid(first === '{' ? 'after object key:value pair' : 'after array element'); break; }
					this.position++;
					this.space();
					if (this.comments && this.text[this.position] === end) { break; }
				}
			}
			if (this.error === '') { this.position++; }
		} else if (first === '"') { node.kind = 'string'; node.text = this.readString(); }
		else if (first === 't' || first === 'f' || first === 'n') {
			const literal = first === 't' ? 'true' : first === 'f' ? 'false' : 'null';
			node.kind = literal === 'null' ? 'null' : 'bool';
			node.text = literal;
			for (const character of literal) {
				if (this.text[this.position] !== character) { this.invalid(`in literal ${literal} (expecting '${character}')`); break; }
				this.position++;
			}
		} else if (first === '-' || (first >= '0' && first <= '9')) {
			node.kind = 'number';
			if (first === '-') { this.position++; }
			const digit = this.text[this.position] ?? '';
			if (digit === '0') { this.position++; }
			else if (digit >= '1' && digit <= '9') { while (this.digit()) { this.position++; } }
			else { this.invalid('in numeric literal'); }
			if (this.text[this.position] === '.') {
				this.position++;
				if (!this.digit()) { this.invalid('in numeric literal'); }
				while (this.digit()) { this.position++; }
			}
			if (this.text[this.position] === 'e' || this.text[this.position] === 'E') {
				this.position++;
				if (this.text[this.position] === '+' || this.text[this.position] === '-') { this.position++; }
				if (!this.digit()) { this.invalid('in numeric literal'); }
				while (this.digit()) { this.position++; }
			}
			node.text = this.text.slice(start, this.position);
		} else { this.invalid('looking for beginning of value'); }
		node.raw = this.text.slice(start, this.position);
		return index;
	}
	digit(): boolean { const value = this.text[this.position] ?? ''; return value !== '' && value >= '0' && value <= '9'; }
	parse(): void {
		this.value();
		this.space();
		if (this.error === '' && this.position < this.text.length) { this.invalid('after top-level value'); }
	}
}

export function parseJson(text: string, comments: boolean): JsonDocument {
	const document = new JsonDocument(text, comments);
	document.parse();
	return document;
}

// Like json.Compact, retains numeric spellings and string escapes, stripping only whitespace.
export function compact(raw: string): string {
	let result = '';
	let inString = false;
	let escaped = false;
	for (const character of raw) {
		if (inString || (character !== ' ' && character !== '\n' && character !== '\r' && character !== '\t')) { result += character; }
		if (inString && escaped) { escaped = false; }
		else if (inString && character === '\\') { escaped = true; }
		else if (character === '"') { inString = !inString; }
	}
	return result;
}
