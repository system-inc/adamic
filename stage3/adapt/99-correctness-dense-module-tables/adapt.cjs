#!/usr/bin/env node
'use strict';
// #s59veah, Adamic's ruling: two tables tsc builds at module load wrote past the end of an empty array, leaving holes,
// and under docs/rulings.md:318 natively that write stops the program before it reads any input. Each now builds
// without a hole, with the same value at every read:
// - emitter.ts: createBracketsMap's four rows sat at ListFormat.Braces, Parenthesis, AngleBrackets and SquareBrackets,
//   1024, 2048, 4096 and 8192, so a dense fill would push 8,189 undefined slots to hold four rows. It's a
//   Map<number, readonly [string, string]> instead, and its two readers, getOpeningBracket and getClosingBracket, read
//   it by `.get` with the same masked key and the same slot. A key with no row gave undefined from the array and gives
//   undefined from the Map, so a missing row stops both forms alike: Node throws reading its slot, and Adamic stops
//   on the `!` the readers already carry (adaptation 33).
// - core.ts: arrayToNumericMap fills the table with undefined through each key before it writes the key, the ruling's
//   explicit dense fill, so every write lands on an index that exists (makeReverseMap's form, Apple's adapter) and
//   none leans on :318's amendment for a write at `length`. Its one caller in src/ is classifier.ts's noRegexTable,
//   13 keys, the largest SyntaxKind.TrueKeyword at 112, so 113 slots. It's read only by index, as
//   `!noRegexTable[lastNonTriviaToken]`, where undefined and a hole read alike. Its return types and noRegexTable's
//   annotation now say undefined, which the filled slots hold.
// JavaScript-changing, value-identical at every read (the lane count is in the README). Each owner is found through the
// stock AST and pinned whole, as it arrives (adaptations 30 and 33 assert the indexed reads in arrayToNumericMap and the
// two bracket readers) or as this leaves it; any other text, and a tree partly adapted, is refused before any write,
// and a second run writes nothing. `rewrite(tree, { restore: true })` is the mutant: every owner back as it arrived.
// This directory sorts after every adapter that touches these owners (30, 33) and after 99-correctness-core-*, and no
// directory after it touches them: 99-correctness-receiver-printer pins createPrinter whole, which calls the two
// readers and holds neither them nor the table.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');

const emitter = 'src/compiler/emitter.ts';
const core = 'src/compiler/core.ts';
const classifier = 'src/services/classifier.ts';

const bracketsBefore = `function createBracketsMap() {
    const brackets: string[][] = [];
    brackets[ListFormat.Braces] = ["{", "}"];
    brackets[ListFormat.Parenthesis] = ["(", ")"];
    brackets[ListFormat.AngleBrackets] = ["<", ">"];
    brackets[ListFormat.SquareBrackets] = ["[", "]"];
    return brackets;
}`;
const bracketsAfter = `function createBracketsMap() {
    // A Map, not an array: the four keys are 1024 to 8192, and an array would hold holes below each.
    const brackets = new Map<number, readonly [string, string]>();
    brackets.set(ListFormat.Braces, ["{", "}"]);
    brackets.set(ListFormat.Parenthesis, ["(", ")"]);
    brackets.set(ListFormat.AngleBrackets, ["<", ">"]);
    brackets.set(ListFormat.SquareBrackets, ["[", "]"]);
    return brackets;
}`;

const reader = (name, slot) => ({
    before: `function ${name}(format: ListFormat) {
    return brackets[format & ListFormat.BracketsMask]![${slot}]!;
}`,
    after: `function ${name}(format: ListFormat) {
    return brackets.get(format & ListFormat.BracketsMask)![${slot}];
}`,
});
const opening = reader('getOpeningBracket', 0);
const closing = reader('getClosingBracket', 1);

const numericMapBefore = `export function arrayToNumericMap<T>(array: readonly T[], makeKey: (value: T) => number): T[];
/** @internal */
export function arrayToNumericMap<T, U>(array: readonly T[], makeKey: (value: T) => number, makeValue: (value: T) => U): U[];
/** @internal */
export function arrayToNumericMap<T, U>(array: readonly T[], makeKey: (value: T) => number, makeValue: (value: T) => T | U = identity): (T | U)[] {
    const result: (T | U)[] = [];
    for (let i = 0; i < array.length; i++) {
        const value = array[i]!;
        result[makeKey(value)] = makeValue(value);
    }
    return result;
}`;
const numericMapAfter = `export function arrayToNumericMap<T>(array: readonly T[], makeKey: (value: T) => number): (T | undefined)[];
/** @internal */
export function arrayToNumericMap<T, U>(array: readonly T[], makeKey: (value: T) => number, makeValue: (value: T) => U): (U | undefined)[];
/** @internal */
export function arrayToNumericMap<T, U>(array: readonly T[], makeKey: (value: T) => number, makeValue: (value: T) => T | U = identity): (T | U | undefined)[] {
    const result: (T | U | undefined)[] = [];
    for (let i = 0; i < array.length; i++) {
        const value = array[i]!;
        const key = makeKey(value);
        // The table is filled with undefined through the key before the write, so it has no holes, the write lands on
        // a slot that exists, and a read by index gives what a hole gave.
        while (result.length <= key) result.push(undefined);
        result[key] = makeValue(value);
    }
    return result;
}`;

const noRegexBefore = `const noRegexTable: true[] = arrayToNumericMap<SyntaxKind, true>(
    [
        SyntaxKind.Identifier,
        SyntaxKind.StringLiteral,
        SyntaxKind.NumericLiteral,
        SyntaxKind.BigIntLiteral,
        SyntaxKind.RegularExpressionLiteral,
        SyntaxKind.ThisKeyword,
        SyntaxKind.PlusPlusToken,
        SyntaxKind.MinusMinusToken,
        SyntaxKind.CloseParenToken,
        SyntaxKind.CloseBracketToken,
        SyntaxKind.CloseBraceToken,
        SyntaxKind.TrueKeyword,
        SyntaxKind.FalseKeyword,
    ],
    token => token,
    () => true,
);`;
const noRegexAfter = noRegexBefore.replace('const noRegexTable: true[] =', 'const noRegexTable: (true | undefined)[] =');

// Each owner: its file, how it's found at the file's top level, how many declarations it spans (an overloaded function
// spans all of them), and its text as it arrives and as this leaves it.
const pins = [
    { file: emitter, name: 'createBracketsMap', kind: 'function', declarations: 1, before: bracketsBefore, after: bracketsAfter },
    { file: emitter, name: 'getOpeningBracket', kind: 'function', declarations: 1, ...opening },
    { file: emitter, name: 'getClosingBracket', kind: 'function', declarations: 1, ...closing },
    { file: core, name: 'arrayToNumericMap', kind: 'function', declarations: 3, before: numericMapBefore, after: numericMapAfter },
    { file: classifier, name: 'noRegexTable', kind: 'variable', declarations: 1, before: noRegexBefore, after: noRegexAfter },
];
for (const pin of pins) assert.notEqual(pin.before, pin.after, `${pin.name}: the edit didn't apply to its own pin`);
const files = [...new Set(pins.map(pin => pin.file))];

function parse(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${fileName}: parse error`);
    return source;
}

// The top-level statements that declare the pin's name: function declarations, or a variable statement whose one
// declaration it is.
function owners(source, pin) {
    return source.statements.filter(node => pin.kind === 'function'
        ? ts.isFunctionDeclaration(node) && node.name?.text === pin.name
        : ts.isVariableStatement(node) && node.declarationList.declarations.length === 1
            && ts.isIdentifier(node.declarationList.declarations[0].name) && node.declarationList.declarations[0].name.text === pin.name);
}

// One file's owners, each read as 'before' or 'after', and the edits that bring each to `wanted`. Line endings are
// judged per owner, never per file: emitter.ts arrives here with one ending, but adapters after this one write lines
// into it in the other, so the finished tree's emitter.ts mixes the two outside these owners (J-1hv0x7d), and a second
// run or the restore mutant over that tree must still read it. Each owner is compared with its endings made LF, and
// its replacement is written in the ending the owner arrived in. An owner that mixes them is refused.
function read(file, raw, wanted) {
    const source = parse(file, raw);
    const states = [];
    const edits = [];
    for (const pin of pins.filter(pin => pin.file === file)) {
        const found = owners(source, pin);
        assert.equal(found.length, pin.declarations, `${file}: ${pin.name} declaration count drift (${found.length})`);
        const start = found[0].getStart(source);
        const end = found[found.length - 1].end;
        const span = raw.slice(start, end);
        assert(!(span.includes('\r\n') && /(?<!\r)\n/.test(span)), `${file}: ${pin.name}: mixed line endings`);
        const crlf = span.includes('\r\n');
        const current = span.replaceAll('\r\n', '\n');
        assert([pin.before, pin.after].includes(current), `${file}: ${pin.name} drift or partial adaptation, found ${JSON.stringify(current)}`);
        const state = current === pin.after ? 'after' : 'before';
        states.push({ file, name: pin.name, state });
        if (state !== wanted) edits.push({ start, end, text: crlf ? pin[wanted].replaceAll('\n', '\r\n') : pin[wanted] });
    }
    return { states, edits };
}

// One file as it should be, from its raw text, every byte outside the owners kept.
function planFile(file, raw, wanted) {
    assert(!/\r(?!\n)/.test(raw), `${file}: bare CR`);
    const { states, edits } = read(file, raw, wanted);
    let output = raw;
    edits.sort((a, b) => b.start - a.start);
    for (const edit of edits) output = output.slice(0, edit.start) + edit.text + output.slice(edit.end);
    if (output !== raw) parse(file, output);
    return { states, output, edits: edits.length };
}

// The tree's files as they should be. Every owner must read as it arrived, or every one as adapted.
function plan(raws, wanted = 'after') {
    const plans = Object.entries(raws).map(([file, raw]) => ({ file, raw, ...planFile(file, raw, wanted) }));
    const states = plans.flatMap(entry => entry.states);
    const adapted = states.filter(state => state.state === 'after');
    assert(adapted.length === 0 || adapted.length === states.length,
        `partly adapted: ${states.map(state => `${state.file} ${state.name} ${state.state}`).join(', ')}`);
    return plans;
}

function rewrite(tree, { restore = false } = {}) {
    const wanted = restore ? 'before' : 'after';
    const raws = Object.fromEntries(files.map(file => [file, fs.readFileSync(path.join(tree, file), 'utf8')]));
    const plans = plan(raws, wanted);
    const changed = plans.filter(entry => entry.output !== entry.raw);
    if (!changed.length) return { files: 0, edits: 0, javascriptChanging: true };
    // The planned files must read wholly as wanted to a second plan, which then edits nothing.
    const second = plan(Object.fromEntries(plans.map(entry => [entry.file, entry.output])), wanted);
    assert(second.every(entry => entry.output === entry.raw), 'invalid planned result');
    for (const file of files) assert.equal(fs.readFileSync(path.join(tree, file), 'utf8'), raws[file], `${file}: changed while planning`);
    for (const entry of changed) fs.writeFileSync(path.join(tree, entry.file), entry.output);
    return { files: changed.length, edits: changed.reduce((sum, entry) => sum + entry.edits, 0), javascriptChanging: true };
}

const adapt = tree => rewrite(tree);
module.exports = { adapt, rewrite, plan, pins };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify({ typescript: ts.version, task: '#s59veah', ...adapt(path.resolve(process.argv[2])) }));
}
