// A port of cohere's internal/format/css/values/parser.go to Adamic 0.1: postcss-values-parser 2.0.1,
// lib/parser.js, with lib/errors/ParserError.js. Each piece names the Go it reads as.
//
// Where the port differs from the Go, and why:
//
//   - Throws. Upstream throws; the Go panics with an *Error and Parse recovers it. 0.1 has no
//     exceptions, so a throw here records upstream's String(error) in #thrown and returns at once, and
//     every caller that can see one returns too, so nothing upstream wouldn't run after a throw runs
//     here. The loop stops at the first. The TypeErrors upstream hits reading a property of undefined
//     are thrown where its expression reads them, in its order of evaluation.
//   - Offsets. The Go converts every sourceIndex to a byte offset; the port keeps upstream's UTF-16
//     index. The test converts the Go's back with cohere's own utf16IndexOf, as cohere's oracle test
//     does before comparing with the library.
//   - The methods come callee first, not in the Go's order: stage 0 takes a call to a method declared
//     further down for a void one (gitignore's GAPS.md, gap 3).
//   - The Go's unexported fields and methods are #private members (gap 4 in GAPS.md is closed).

import { panic } from 'adamic';
import { alphaNum, type Token } from './tokenize.ts';
import {
	leaf,
	newAtWord,
	newComment,
	newFunc,
	newNumber,
	newParen,
	newRoot,
	newString,
	newValue,
	newWord,
	Position,
	Source,
	type ValueNode,
	ValueTree,
} from './nodes.ts';

// What a parse gives: the Root, or upstream's String(error) for what it threw.
export type Parsed = { readonly kind: 'Ok'; readonly root: ValueTree } | { readonly kind: 'Error'; readonly message: string };

// parser.go: tokenSource, the { start, end } most node kinds take straight from their token.
function tokenSource(token: Token): Source {
	return new Source(new Position(token.startLine, token.startColumn), new Position(token.endLine, token.endColumn));
}

// parser.go: commentDelimiters, value.replace(/\/\*|\*\//g, ""): every "/*" and "*/", left to right,
// not overlapping.
function commentDelimiters(value: string): string {
	let result = '';
	for (let index = 0; index < value.length; index++) {
		const pair = value.slice(index, index + 2);
		if (pair === '/*' || pair === '*/') {
			index++;
			continue;
		}
		result += value[index] ?? '';
	}
	return result;
}

// isDigit is whether the unit at index is an ASCII digit; NaN past the end is not.
function isDigit(text: string, index: number): boolean {
	const code = text.charCodeAt(index);
	return code >= 0x30 && code <= 0x39;
}

// parser.go: rNumber, /^[\+\-]?((\d+(\.\d*)?)|(\.\d+))([eE][\+\-]?\d+)?/: the length of the match at
// the start of text, or -1.
function rNumber(text: string): number {
	let index = 0;
	if (text.startsWith('+') || text.startsWith('-')) {
		index++;
	}
	const digits = (from: number): number => {
		let end = from;
		while (isDigit(text, end)) {
			end++;
		}
		return end;
	};
	const integerEnd = digits(index);
	if (integerEnd > index) {
		index = integerEnd;
		if (text[index] === '.') {
			index = digits(index + 1);
		}
	} else if (text[index] === '.' && digits(index + 1) > index + 1) {
		index = digits(index + 1);
	} else {
		return -1;
	}
	if (text[index] === 'e' || text[index] === 'E') {
		let exponent = index + 1;
		if (text[exponent] === '+' || text[exponent] === '-') {
			exponent++;
		}
		const exponentEnd = digits(exponent);
		if (exponentEnd > exponent) {
			index = exponentEnd;
		}
	}
	return index;
}

// parser.go: rNoFollow, /^(?!\#([a-z0-9]+))[\#\{\}]/gi, tested once on a fresh regex: a word opening with
// '#', '{' or '}', unless it is '#' and an ASCII letter or digit.
function rNoFollow(word: string): boolean {
	if (!word.startsWith('#') && !word.startsWith('{') && !word.startsWith('}')) {
		return false;
	}
	return !(word.startsWith('#') && alphaNum(word.charCodeAt(1)));
}

// parser.go: isHex, /^#(.+)/: '#' and at least one character that is not a line terminator. The
// terminator test can't be reached: the tokenizer starts a word with '#' only when an ASCII letter or
// digit follows it, and makes any other '#' a token of its own, which splitWord never joins to what
// follows. A mutant without it agrees with the library on every case, so the test has none.
function isHex(value: string): boolean {
	const second = value.charCodeAt(1);
	return value.startsWith('#') && value.length > 1 && second !== 0x0a && second !== 0x0d && second !== 0x2028 && second !== 0x2029;
}

// parser.go: isColor, /^#([0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$/i, ASCII only.
function isColor(value: string): boolean {
	if (!value.startsWith('#')) {
		return false;
	}
	const digits = value.length - 1;
	if (digits !== 3 && digits !== 4 && digits !== 6 && digits !== 8) {
		return false;
	}
	for (let index = 1; index < value.length; index++) {
		const unit = value.charCodeAt(index);
		if (!((unit >= 0x30 && unit <= 0x39) || (unit >= 0x61 && unit <= 0x66) || (unit >= 0x41 && unit <= 0x46))) {
			return false;
		}
	}
	return true;
}

// replaceFirst is value.replace(search, ''), with a string pattern: the first occurrence, which is not
// always the suffix. 0.1's library has no replace.
function replaceFirst(value: string, search: string): string {
	const found = value.indexOf(search);
	return found === -1 ? value : value.slice(0, found) + value.slice(found + search.length);
}

// A leaf's children in its tree. tree() starts from it and replaces it for a container, where it would
// write `children ?? []`, since stage 0 doesn't lower an empty array literal as a default yet (gap 5 in
// GAPS.md).
const noChildren: readonly ValueTree[] = [];

// parser.go: parser.
export class Parser {
	// cache needs to be an array for values with more than 1 level of function nesting
	readonly #cache: ValueNode[] = [];
	readonly #loose: boolean;
	#position = 0;
	// upstream also keeps an `unbalanced` count on the parser that nothing reads; the count that
	// matters lives on the Value and FunctionNode nodes.
	readonly #root: ValueNode;
	#current: ValueNode;
	// Each container's children, by its id: the Go's Container nodes field, kept here while the parse
	// runs (nodes.ts says why).
	readonly #children: ValueNode[][] = [];
	readonly #tokens: readonly Token[];
	#spaces = '';
	// What upstream threw, as String(error), or '' while nothing has.
	#thrown = '';

	// parser.go: newParser. The Go tokenizes here; the port is handed the tokens, so that a tokenizer's
	// throw needn't come out of a constructor (values.ts).
	constructor(loose: boolean, tokens: readonly Token[]) {
		this.#loose = loose;
		this.#children.push([]);
		this.#root = newRoot(0);

		this.#children.push([]);
		const value = newValue(1);

		(this.#children[0] ?? panic('the root has children')).push(value);

		this.#current = value;
		this.#tokens = tokens;
	}

	// parser.go: tokenAt; currToken, nextToken and prevToken. undefined where upstream's index is out of
	// range.
	#tokenAt(index: number): Token | undefined {
		return this.#tokens[index];
	}

	// The current token, which every caller reads only while position is inside the tokens, as
	// upstream's do.
	#currToken(): Token {
		return this.#tokenAt(this.#position) ?? panic(`no token at ${this.#position} of ${this.#tokens.length}`);
	}

	#nextToken(): Token | undefined {
		return this.#tokenAt(this.#position + 1);
	}

	#prevToken(): Token | undefined {
		return this.#tokenAt(this.#position - 1);
	}

	// parser.go: error, which throws a ParserError.
	#error(message: string, token: Token): void {
		this.#thrown = `ParserError: ${message} at line: ${token.startLine}, column ${token.startColumn}`;
	}

	// parser.go: typeError, V8's TypeError for reading a property of undefined.
	#typeError(key: string): void {
		this.#thrown = `TypeError: Cannot read properties of undefined (reading '${key}')`;
	}

	// A container's children, as the Go's nodes field holds them.
	#childrenOf(container: ValueNode): ValueNode[] {
		return this.#children[container.id] ?? panic(`no children for a ${container.type}`);
	}

	// A new container's id, with no children yet.
	#nextId(): number {
		this.#children.push([]);
		return this.#children.length - 1;
	}

	// parser.go: last, Container's `get last`: the current node's last child, or undefined.
	#last(): ValueNode | undefined {
		return this.#childrenOf(this.#current).at(-1);
	}

	// parser.go: newNode.
	#newNode(node: ValueNode): void {
		if (this.#spaces !== '') {
			node.raws.before += this.#spaces;
			this.#spaces = '';
		}

		this.#childrenOf(this.#current).push(node);
	}

	// parser.go: colon.
	#colon(): void {
		const token = this.#currToken();

		this.#newNode(leaf('colon', token.value, tokenSource(token), token.index));

		this.#position++;
	}

	// parser.go: comma.
	#comma(): void {
		const token = this.#currToken();

		this.#newNode(leaf('comma', token.value, tokenSource(token), token.index));

		this.#position++;
	}

	// parser.go: comment.
	#comment(): void {
		let inline = false;
		let value = commentDelimiters(this.#currToken().value);

		if (this.#loose && value.startsWith('//')) {
			value = value.slice(2);
			inline = true;
		}

		const node = newComment(value, inline, tokenSource(this.#currToken()), this.#currToken().index);

		this.#newNode(node);
		this.#position++;
	}

	// parser.go: space.
	#space(): void {
		const token = this.#currToken();
		// Handle space before and after the selector
		// Upstream compares the next token's type with ',' and ')', but comma tokens are typed 'comma',
		// so only ')' and the last position ever match. The kind is widened to a string to say so, since no
		// TokenKind is ','.
		const next = this.#nextToken();
		const nextKind: string = next === undefined ? '' : next.kind;
		if (this.#position === this.#tokens.length - 1 || nextKind === ',' || nextKind === ')') {
			const currentLast = this.#last();
			if (currentLast === undefined) {
				this.#typeError('raws');
				return;
			}
			currentLast.raws.after += token.value;
			this.#position++;
		} else {
			this.#spaces = token.value;
			this.#position++;
		}
	}

	// parser.go: unicodeRange.
	#unicodeRange(): void {
		const token = this.#currToken();

		this.#newNode(leaf('unicode-range', token.value, tokenSource(token), token.index));

		this.#position++;
	}

	// parser.go: string.
	#string(): void {
		const token = this.#currToken();
		let value = token.value;
		// rQuote is /^(\"|\')/
		const quoted = value.startsWith('"') || value.startsWith("'");
		let quote = '';

		if (quoted) {
			quote = value.slice(0, 1);
			// set value to the string within the quotes
			// quotes are stored in raws
			value = value.slice(1, value.length - 1);
		}

		const node = newString(value, tokenSource(token), token.index, quoted);

		node.raws.quote = quote;

		this.#newNode(node);
		this.#position++;
	}

	// parser.go: splitWord.
	#splitWord(): void {
		let nextToken = this.#nextToken();
		let word = this.#currToken().value;

		// treat css-like groupings differently so they can be inspected,
		// but don't address them as anything but a word, but allow hex values
		// to pass through.
		if (!rNoFollow(word)) {
			while (nextToken !== undefined && nextToken.kind === 'word') {
				this.#position++;

				word += this.#currToken().value;

				nextToken = this.#nextToken();
			}
		}

		// hasAt = indexesOf(word, '@'); indices = sortAscending(uniq(flatten([[0], hasAt]))): 0 and every
		// '@', ascending, once each.
		const hasAt: number[] = [];
		for (let index = 0; index < word.length; index++) {
			if (word[index] === '@') {
				hasAt.push(index);
			}
		}
		const indices: number[] = [0];
		for (const index of hasAt) {
			if (index !== 0) {
				indices.push(index);
			}
		}

		const currToken = this.#currToken();
		for (let i = 0; i < indices.length; i++) {
			const ind = indices[i] ?? panic(`index ${i} out of range`);
			const following = indices[i + 1];
			const index = following !== undefined && following !== 0 ? following : word.length;
			const value = word.slice(ind, index);
			let node: ValueNode;

			const nodeSource = new Source(new Position(currToken.startLine, currToken.startColumn + ind), new Position(currToken.endLine, currToken.startColumn + (index - 1)));
			const nodeSourceIndex = currToken.index + ind;

			if (hasAt.includes(ind)) {
				node = newAtWord(value.slice(1), nodeSource, nodeSourceIndex, this.#nextId());
			} else if (rNumber(currToken.value) >= 0) {
				const matched = rNumber(value);
				const unit = matched >= 0 ? value.slice(matched) : value;

				// value.replace(unit, ''): the first occurrence, which is not always the suffix.
				node = newNumber(replaceFirst(value, unit), nodeSource, nodeSourceIndex, unit);
			} else if (nextToken !== undefined && nextToken.kind === '(') {
				node = newFunc(value, nodeSource, nodeSourceIndex, this.#nextId());
				this.#cache.push(this.#current);
			} else {
				node = newWord(value, nodeSource, nodeSourceIndex, isHex(value), isColor(value));
			}

			this.#newNode(node);
		}

		this.#position++;
	}

	// parser.go: word.
	#word(): void {
		this.#splitWord();
	}

	// parser.go: operator.
	#operator(): void {
		// if a +|- operator is followed by a non-word character (. is allowed) and
		// is preceded by a non-word character. (5+5)
		const char = this.#currToken().value;

		if (char === '+' || char === '-') {
			// only inspect if the operator is not the first token, and we're only
			// within a calc() function: the only spec-valid place for math expressions
			if (!this.#loose) {
				if (this.#position > 0) {
					const previous = this.#prevToken() ?? panic('no token before a position past the first');
					const next = this.#nextToken();
					if (this.#current.type === 'func' && this.#current.value === 'calc') {
						// allow operators to be proceeded by spaces and opening parens
						if (previous.kind !== 'space' && previous.kind !== '(') {
							this.#error('Syntax Error', this.#currToken());
							return;
						}
						if (next === undefined) {
							this.#typeError('0');
							return;
						}
						if (next.kind !== 'space' && next.kind !== 'word') {
							// valid: calc(1 - +2)
							// invalid: calc(1 -+2)
							this.#error('Syntax Error', this.#currToken());
							return;
						}
						if (next.kind === 'word') {
							const currentLast = this.#last();
							if (currentLast === undefined) {
								this.#typeError('type');
								return;
							}
							if (currentLast.type !== 'operator' && currentLast.value !== '(') {
								// valid: calc(1 - +2)
								// valid: calc(-0.5 + 2)
								// invalid: calc(1 -2)
								this.#error('Syntax Error', this.#currToken());
								return;
							}
						}
					} else {
						if (next === undefined) {
							this.#typeError('0');
							return;
						}
						if (next.kind === 'space' || next.kind === 'operator' || previous.kind === 'operator') {
							// if we're not in a function and someone has doubled up on operators,
							// or they're trying to perform a calc outside of a calc
							// eg. +-4px or 5+ 5, throw an error
							this.#error('Syntax Error', this.#currToken());
							return;
						}
					}
				}
			}

			if (!this.#loose) {
				const next = this.#nextToken();
				if (next === undefined) {
					this.#typeError('0');
					return;
				}
				if (next.kind === 'word') {
					this.#word();
					return;
				}
			} else {
				const currentLast = this.#last();
				if (this.#childrenOf(this.#current).length === 0 || (currentLast !== undefined && currentLast.type === 'operator')) {
					const next = this.#nextToken();
					if (next === undefined) {
						this.#typeError('0');
						return;
					}
					if (next.kind === 'word') {
						this.#word();
						return;
					}
				}
			}
		}

		const token = this.#currToken();
		// Upstream ends the operator where it starts, and takes its sourceIndex from token[4], which is
		// the end line rather than the index.
		const node = leaf('operator', token.value, new Source(new Position(token.startLine, token.startColumn), new Position(token.startLine, token.startColumn)), token.endLine);

		this.#position++;

		this.#newNode(node);
	}

	// parser.go: parenOpen.
	#parenOpen(): void {
		let unbalancedCount = 1;
		let pos = this.#position + 1;
		const token = this.#currToken();

		// check for balanced parens
		while (pos < this.#tokens.length && unbalancedCount !== 0) {
			const tkn = this.#tokens[pos] ?? panic(`no token at ${pos}`);

			if (tkn.kind === '(') {
				unbalancedCount++;
			}
			if (tkn.kind === ')') {
				unbalancedCount--;
			}
			pos++;
		}

		if (unbalancedCount !== 0) {
			this.#error('Expected closing parenthesis', token);
			return;
		}

		// ok, all parens are balanced. continue on

		const currentLast = this.#last();

		if (currentLast !== undefined && currentLast.type === 'func' && currentLast.unbalanced < 0) {
			currentLast.unbalanced = 0; // ok we're ready to add parens now
			this.#current = currentLast;
		}

		this.#current.unbalanced++;

		this.#newNode(newParen(token.value, tokenSource(token), token.index));

		this.#position++;

		// url functions get special treatment, and anything between the function
		// parens get treated as one word, if the contents aren't not a string.
		const afterParen = this.#currToken();
		if (
			this.#current.type === 'func' &&
			this.#current.unbalanced !== 0 &&
			this.#current.value === 'url' &&
			afterParen.kind !== 'string' &&
			afterParen.kind !== ')' &&
			!this.#loose
		) {
			let nextToken = this.#nextToken();
			let value = afterParen.value;
			const start = new Position(afterParen.startLine, afterParen.startColumn);

			while (nextToken !== undefined && nextToken.kind !== ')' && this.#current.unbalanced !== 0) {
				this.#position++;
				value += this.#currToken().value;
				nextToken = this.#nextToken();
			}

			if (this.#position !== this.#tokens.length - 1) {
				// skip the following word definition, or it'll be a duplicate
				this.#position++;

				// Constructed directly, so unlike splitWord's words it carries no isHex or isColor.
				const word = this.#currToken();
				this.#newNode(leaf('word', value, new Source(start, new Position(word.endLine, word.endColumn)), word.index));
			}
		}
	}

	// parser.go: parenClose.
	#parenClose(): void {
		const token = this.#currToken();

		this.#newNode(newParen(token.value, tokenSource(token), token.index));

		this.#position++;

		if (this.#position >= this.#tokens.length - 1 && this.#current.unbalanced === 0) {
			return;
		}

		this.#current.unbalanced--;

		if (this.#current.unbalanced < 0) {
			this.#error('Expected opening parenthesis', token);
			return;
		}

		const cached = this.#cache.at(-1);
		if (this.#current.unbalanced === 0 && cached !== undefined) {
			this.#current = cached;
			this.#cache.pop();
		}
	}

	// parser.go: parseTokens.
	#parseTokens(): void {
		switch (this.#currToken().kind) {
			case 'space':
				this.#space();
				break;
			case 'colon':
				this.#colon();
				break;
			case 'comma':
				this.#comma();
				break;
			case 'comment':
				this.#comment();
				break;
			case '(':
				this.#parenOpen();
				break;
			case ')':
				this.#parenClose();
				break;
			case 'atword':
			case 'word':
				this.#word();
				break;
			case 'operator':
				this.#operator();
				break;
			case 'string':
				this.#string();
				break;
			case 'unicoderange':
				this.#unicodeRange();
				break;
			default:
				this.#word();
				break;
		}
	}

	// The tree of a node and everything under it, as the library returns it.
	#tree(node: ValueNode): ValueTree {
		let nodes: readonly ValueTree[] = noChildren;
		if (node.id >= 0) {
			nodes = this.#childrenOf(node).map((child) => this.#tree(child));
		}
		return new ValueTree(node, nodes);
	}

	// parser.go: loop.
	#loop(): Parsed {
		while (this.#position < this.#tokens.length) {
			this.#parseTokens();
			if (this.#thrown !== '') {
				return { kind: 'Error', message: this.#thrown };
			}
		}

		const currentLast = this.#last();
		if (currentLast === undefined) {
			if (this.#spaces !== '') {
				this.#current.raws.before += this.#spaces;
			}
		} else if (this.#spaces !== '') {
			currentLast.raws.after += this.#spaces;
		}

		this.#spaces = '';

		return { kind: 'Ok', root: this.#tree(this.#root) };
	}

	// parser.go: parse.
	parse(): Parsed {
		return this.#loop();
	}
}
