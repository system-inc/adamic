'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

// The after text of the deduplicateSorted owner in 99-correctness-axolotl-r.
const before = [
    "function deduplicateSorted<T>(array: SortedReadonlyArray<T>, comparer: EqualityComparer<T> | Comparer<T>): SortedReadonlyArray<T> {",
    "    if (array.length === 0) return emptyArray as unknown as SortedReadonlyArray<T>;",
    "",
    "    let last = array[0]!;",
    "    const deduplicated: T[] = [last];",
    "    for (let i = 1; i < array.length; i++) {",
    "        const next = array[i]!;",
    "        switch (comparer(next, last)) {",
    "            // equality comparison",
    "            case true:",
    "",
    "            // relational comparison",
    "            // falls through",
    "            case Comparison.EqualTo:",
    "                continue;",
    "",
    "            case Comparison.LessThan:",
    "                // If `array` is sorted, `next` should **never** be less than `last`.",
    "                return Debug.fail(\"Array is unsorted.\");",
    "        }",
    "",
    "        deduplicated.push(last = next);",
    "    }",
    "",
    "    return deduplicated as unknown as SortedReadonlyArray<T>;",
    "}",
].join('\n');

const replacement = [
    'const comparison = comparer(next, last);',
    '        if (typeof comparison === "boolean") {',
    '            // equality comparison',
    '            if (comparison) continue;',
    '        }',
    '        else {',
    '            // relational comparison',
    '            if (comparison === Comparison.EqualTo) continue;',
    '            if (comparison === Comparison.LessThan) {',
    '                // If `array` is sorted, `next` should **never** be less than `last`.',
    '                return Debug.fail("Array is unsorted.");',
    '            }',
    '        }',
].join('\n');

function ownerOf(text) {
    const source = ts.createSourceFile('core.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'core.ts has syntax errors');
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'deduplicateSorted');
    assert(owners.length > 0, 'missing top-level deduplicateSorted declaration');
    assert(owners.length === 1, 'duplicate top-level deduplicateSorted declarations');
    const owner = owners[0];
    assert(owner.body, 'deduplicateSorted has no body');
    return owner;
}

function loopOf(owner) {
    const loops = owner.body.statements.filter(ts.isForStatement);
    assert.equal(loops.length, 1, 'expected exactly one loop in deduplicateSorted');
    const loop = loops[0];
    assert(ts.isBlock(loop.statement), 'deduplicateSorted loop has no block');
    return loop;
}

function span(statements) {
    return {start: statements[0].getStart(), end: statements.at(-1).end};
}

function switchSpan(owner) {
    const switches = loopOf(owner).statement.statements.filter(ts.isSwitchStatement);
    assert.equal(switches.length, 1, 'expected exactly one switch in the deduplicateSorted loop body');
    return span(switches);
}

function adaptedSpan(owner) {
    const statements = loopOf(owner).statement.statements;
    const first = statements.findIndex(statement => ts.isVariableStatement(statement) &&
        statement.declarationList.declarations.length === 1 &&
        statement.declarationList.declarations[0].name.getText() === 'comparison');
    assert(first >= 0 && statements[first + 1] && ts.isIfStatement(statements[first + 1]),
        'expected the comparison declaration and if in the adapted loop body');
    return span(statements.slice(first, first + 2));
}

function unchanged(text, cut) {
    return [text.slice(0, cut.start), text.slice(cut.end)];
}

function transform(text) {
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    const original = before.replaceAll('\n', newline);
    const inserted = replacement.replaceAll('\n', newline);
    const pinnedCut = switchSpan(ownerOf(original));
    const pinnedAfter = original.slice(0, pinnedCut.start) + inserted + original.slice(pinnedCut.end);
    const owner = ownerOf(text);
    const actual = owner.getText();
    if (actual === pinnedAfter) return text;
    if (actual.includes('const comparison =')) {
        throw new Error('partly adapted deduplicateSorted differs from the pinned after text');
    }
    assert(actual === original, 'deduplicateSorted differs from the pinned before text');
    const cut = switchSpan(owner);
    const after = text.slice(0, cut.start) + inserted + text.slice(cut.end);
    const adapted = ownerOf(after);
    assert.equal(adapted.getText(), pinnedAfter, 'adapted deduplicateSorted differs from its pinned form');
    assert.deepEqual(unchanged(after, adaptedSpan(adapted)), unchanged(text, cut),
        'text outside the deduplicateSorted switch changed');
    return after;
}

if (require.main === module) {
    try {
        assert.equal(ts.version, '6.0.3', 'expected locked TypeScript 6.0.3');
        assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
        const file = path.join(process.argv[2], 'src/compiler/core.ts');
        const text = fs.readFileSync(file, 'utf8');
        const after = transform(text);
        const edits = Number(after !== text);
        if (edits) fs.writeFileSync(file, after);
        console.log(JSON.stringify({files: edits, edits}));
    } catch (error) {
        console.error('99-seats-family36-comparer-once: ' + error.message.split('\n')[0]);
        process.exitCode = 1;
    }
}

module.exports = {ownerOf, switchSpan, adaptedSpan, unchanged};
