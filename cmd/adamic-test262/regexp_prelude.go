// Adapted from test262's regExpUtils.js.
// Copyright (C) 2017 Mathias Bynens. All rights reserved.
// BSD license: testdata/regexp/LICENSE.
package main

// regExpUtils.js without apply, array length writes or implicit truthiness.
// Chunk flushing is unobservable: every supplied code point is appended in order.
// Keep the upstream whole-string fast path and per-symbol fallback, rather than
// changing the coverage or accepting a partial match as a new test expectation.
const regexpPrelude = `function buildString(args: { readonly loneCodePoints: readonly number[]; readonly ranges: readonly (readonly number[])[] }): string {
	let points: number[] = [];
	const chunks: string[] = [];
	for (const point of args.loneCodePoints) {
		points.push(point);
		if (points.length >= 10000) { chunks.push(String.fromCodePoint(...points)); points = []; }
	}
	for (const range of args.ranges) {
		const start = range[0];
		const end = range[1];
		if (start === undefined || end === undefined) {
			throw new Error("invalid code point range");
		}
		for (let point = start; point <= end; point += 1) {
			points.push(point);
			if (points.length >= 10000) { chunks.push(String.fromCodePoint(...points)); points = []; }
		}
	}
	chunks.push(String.fromCodePoint(...points));
	return chunks.join("");
}

function printCodePoint(point: number): string {
	return "U+" + point.toString(16).toUpperCase().padStart(6, "0");
}

function printStringCodePoints(text: string): string {
	const formatted: string[] = [];
	for (const symbol of text) {
		const point = symbol.codePointAt(0);
		if (point === undefined) { throw new Error("empty symbol"); }
		formatted.push(printCodePoint(point));
	}
	return formatted.join(" ");
}

function testPropertyEscapes(regExp: RegExp, text: string, expression: string): void {
	if (!regExp.test(text)) {
		for (const symbol of text) {
			const point = symbol.codePointAt(0);
			if (point === undefined) { throw new Error("empty symbol"); }
			assert(regExp.test(symbol), expression + " should match " + printCodePoint(point));
		}
	}
}

function testPropertyOfStrings(args: { readonly regExp: RegExp; readonly expression: string; readonly matchStrings: readonly string[]; readonly nonMatchStrings: readonly string[] | undefined }): void {
	const regExp = args.regExp;
	if (!regExp.test(args.matchStrings.join(""))) {
		for (const text of args.matchStrings) {
			assert(regExp.test(text), args.expression + " should match " + text + " " + printStringCodePoints(text));
		}
	}
	const negatives = args.nonMatchStrings;
	if (negatives === undefined) { return; }
	if (regExp.test(negatives.join(""))) {
		for (const text of negatives) {
			assert(!regExp.test(text), args.expression + " should not match " + text + " " + printStringCodePoints(text));
		}
	}
}

const testExtendedCharacterClass = testPropertyOfStrings;

function matchValidator(expectedEntries: readonly (string | undefined)[], expectedIndex: number, expectedInput: string): (match: RegExpExecArray | null) => void {
	return (match: RegExpExecArray | null): void => {
		if (match === null) { throw new Error("missing match"); }
		assertCompareArray<string | undefined>(match, expectedEntries, "Match entries");
		assertSameValue(match.index, expectedIndex, "Match index");
		assertSameValue(match.input, expectedInput, "Match input");
	};
}
`
