// A port of cohere's internal/format/css/values/values.go to Adamic 0.1: postcss-values-parser 2.0.1,
// the parser Prettier's language-css hands declaration values and some at-rule params to.
//
// Prettier constructs it in src/language-css/parse/parse-value.js as
// `new PostcssValuesParser(value, { loose: true }).parse()`, importing lib/parser.js directly, so the
// library's index.js is not part of the port. loose mirrors that options object: its default is false,
// and Prettier passes true. It relaxes the operator checks, keeps url() arguments as tokens, and reads
// `//` as an inline comment.
//
// sourceIndex is upstream's UTF-16 index, where the Go's is a byte offset; source lines and columns are
// upstream's in both. The Go's utf16Units and itoa, which stand between Go strings and JavaScript's,
// have nothing to do here.

import { Parser, type Parsed } from './parser.ts';
import { tokenize } from './tokenize.ts';

// values.go: Parse, `new Parser(value, options).parse()`: the Root, holding one Value, holding the
// nodes. Everything the library throws comes back as an Error with upstream's String(error), which the
// Go marks printing.Syntax and Prettier turns into a value-unknown node. The constructor tokenizes, so
// a TokenizeError comes before the parse, as it does upstream.
export function parse(value: string, loose: boolean): Parsed {
	const tokenized = tokenize(value, loose);
	if (tokenized.kind === 'Error') {
		return tokenized;
	}
	return new Parser(loose, tokenized.tokens).parse();
}
