// A port of cohere's internal/format/graphql/block_string.go to Adamic 0.1: graphql-js 17.0.2,
// language/blockString.js, dedentBlockStringLines, which the lexer calls for every block string.

import { isWhiteSpace } from './characterClasses.ts';

function leadingWhitespace(text: string): number {
	let index = 0;
	while (index < text.length && isWhiteSpace(text.charCodeAt(index))) {
		index++;
	}
	return index;
}

// dedentBlockStringLines produces the value of a block string from its parsed raw value, similar to
// CoffeeScript's block string, Python's docstring trim or Ruby's strip_heredoc.
//
// This implements the GraphQL spec's BlockStringValue() static algorithm.
export function dedentBlockStringLines(lines: readonly string[]): string[] {
	let commonIndent = Number.MAX_SAFE_INTEGER;
	let firstNonEmptyLine = -1; // null
	let lastNonEmptyLine = -1;

	for (let index = 0; index < lines.length; index++) {
		const line = lines[index] ?? '';
		const indent = leadingWhitespace(line);

		if (indent === line.length) {
			continue; // skip empty lines
		}

		if (firstNonEmptyLine === -1) {
			firstNonEmptyLine = index;
		}
		lastNonEmptyLine = index;

		if (index !== 0 && indent < commonIndent) {
			commonIndent = indent;
		}
	}

	// Remove common indentation from all lines but first, then leading and trailing blank lines.
	// slice clamps an index past the end, so a blank line shorter than the indent, or an indent still
	// MAX_SAFE_INTEGER, gives ''.
	const dedented: string[] = [];
	for (let index = 0; index < lines.length; index++) {
		const line = lines[index] ?? '';
		dedented.push(index === 0 ? line : line.slice(commonIndent));
	}
	return dedented.slice(firstNonEmptyLine === -1 ? 0 : firstNonEmptyLine, lastNonEmptyLine + 1);
}
