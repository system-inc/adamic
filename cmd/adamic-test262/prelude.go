package main

// prelude is the harness test262's assert.js and sta.js provide, rewritten so stage 0 can lower it.
// harness/compareArray.js is empty upstream: compareArray lives in assert.js, and tests that list
// compareArray.js still call it.
//
// What was adapted, and why:
//
//   - assert, assert.sameValue, assert.notSameValue, assert.throws, assert.compareArray and
//     assert._isSameValue are free functions (assert, assertSameValue, ...). 0.1 refuses expandos,
//     so a function cannot carry properties. rewriteHarnessCalls renames the calls in the test body.
//     Both Node and the native binary run that one rewritten program.
//   - SameValue covers string, number (NaN equals NaN, and -0 is not 0), boolean and undefined.
//     Anything else is not a member of that union, so the checker refuses the call instead of the
//     runner comparing it wrong.
//   - assert.throws is assertThrows(name, fn). It checks that an Error was caught, not that the
//     constructor is the one named: stage 0 only builds `new Error`, and only `instanceof Error`
//     on a catch lowers. The name is what the failure says when nothing was thrown. This is weaker
//     than the harness, on both sides equally. A library call that throws on Node panics natively
//     (docs/0.1.md), and a panic is not caught, so those tests come out fail rather than pass.
//   - compareArray is generic over string, number or boolean. An array of a union does not lower;
//     one copy per element type does. SameValue is the element comparison, as in assert.js.
//   - Test262Error is a class with a message, not a prototype assignment. `throw new Test262Error`
//     is still refused (only `throw new Error` lowers). $DONOTEVALUATE throws an Error.
//   - "use strict" directives are dropped. Every Adamic file is a module, and a string used as a
//     statement does not lower.
const prelude = `function requiredCodePoint(point: number | undefined): number {
	if (point === undefined) { throw new Error("missing code point"); }
	return point;
}

function requiredRegexExec(match: RegExpExecArray | null): RegExpExecArray {
	if (match === null) { throw new Error("missing match result"); }
	return match;
}

function requiredRegexMatch(match: RegExpMatchArray | null): RegExpMatchArray {
	if (match === null) { throw new Error("missing match result"); }
	return match;
}

class Test262Error {
	message: string;
	constructor(message?: string) {
		this.message = message ?? "";
	}
	toString(): string {
		return "Test262Error: " + this.message;
	}
}

function $DONOTEVALUATE(): void {
	throw new Error("Test262: This statement should not be evaluated.");
}

function sameValue(left: string | number | boolean | undefined, right: string | number | boolean | undefined): boolean {
	if (typeof left === "number" && typeof right === "number") {
		if (Number.isNaN(left) && Number.isNaN(right)) {
			return true;
		}
		if (left === 0 && right === 0) {
			return 1 / left === 1 / right;
		}
		return left === right;
	}
	if (typeof left === "string" && typeof right === "string") {
		return left === right;
	}
	if (typeof left === "boolean" && typeof right === "boolean") {
		return left === right;
	}
	return left === undefined && right === undefined;
}

function assert(mustBeTrue: boolean, message?: string): void {
	if (mustBeTrue === true) {
		return;
	}
	throw new Error(message ?? "Expected true but got a non-true value");
}

function assertSameValue(actual: string | number | boolean | undefined, expected: string | number | boolean | undefined, message?: string): void {
	if (sameValue(actual, expected)) {
		return;
	}
	throw new Error(message ?? "Expected SameValue");
}

function assertNotSameValue(actual: string | number | boolean | undefined, unexpected: string | number | boolean | undefined, message?: string): void {
	if (!sameValue(actual, unexpected)) {
		return;
	}
	throw new Error(message ?? "Expected NotSameValue");
}

function assertIsSameValue(left: string | number | boolean | undefined, right: string | number | boolean | undefined): boolean {
	return sameValue(left, right);
}

function compareArray<T extends string | number | boolean | undefined>(actual: readonly T[] | null, expected: readonly T[]): boolean {
	if (actual === null) { return false; }
	if (actual.length !== expected.length) {
		return false;
	}
	for (let index = 0; index < actual.length; index += 1) {
		if (!sameValue(actual[index], expected[index])) {
			return false;
		}
	}
	return true;
}

function assertCompareArray<T extends string | number | boolean | undefined>(actual: readonly T[] | null, expected: readonly T[], message?: string): void {
	if (compareArray(actual, expected)) {
		return;
	}
	throw new Error(message ?? "arrays differ");
}

function assertThrows(expectedName: string, func: () => void, message?: string): void {
	let threw = false;
	try {
		func();
	} catch (caught) {
		threw = true;
		if (!(caught instanceof Error)) {
			throw new Error((message ?? "") + "thrown value was not an Error");
		}
	}
	if (!threw) {
		throw new Error((message ?? "") + "Expected " + expectedName + " to be thrown but no exception was thrown at all");
	}
}
`

// program is one Adamic source: the adapted harness, then the test with its frontmatter removed and
// its harness calls rewritten. Node runs the same text as a module (.mts); adamic reads it as .ts.
func program(testBody string) string {
	return prelude + "\n" + stripUseStrict(testBody)
}
