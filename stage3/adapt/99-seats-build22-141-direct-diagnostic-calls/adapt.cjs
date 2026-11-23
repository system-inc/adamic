#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const file = path.join(process.argv[2], 'src/compiler/watch.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, file);

const original = [
    'let messageAndArgs: DiagnosticAndArguments;',
    '    if (errorCount === 1) {',
    '        messageAndArgs = filesInError[0] !== undefined ? [Diagnostics.Found_1_error_in_0, firstFileReference] : [Diagnostics.Found_1_error];',
    '    }',
    '    else {',
    '        messageAndArgs = distinctFileNamesWithLines.length === 0 ? [Diagnostics.Found_0_errors, errorCount] :',
    '            distinctFileNamesWithLines.length === 1 ? [Diagnostics.Found_0_errors_in_the_same_file_starting_at_Colon_1, errorCount, firstFileReference] :',
    '            [Diagnostics.Found_0_errors_in_1_files, errorCount, distinctFileNamesWithLines.length];',
    '    }',
    '',
    '    const d = createCompilerDiagnostic(...messageAndArgs);',
].join('\r\n');
const direct = [
    'let d: Diagnostic;',
    '    if (errorCount === 1) {',
    '        d = filesInError[0] !== undefined',
    '            ? createCompilerDiagnostic(Diagnostics.Found_1_error_in_0, firstFileReference)',
    '            : createCompilerDiagnostic(Diagnostics.Found_1_error);',
    '    }',
    '    else {',
    '        d = distinctFileNamesWithLines.length === 0 ? createCompilerDiagnostic(Diagnostics.Found_0_errors, errorCount) :',
    '            distinctFileNamesWithLines.length === 1 ? createCompilerDiagnostic(Diagnostics.Found_0_errors_in_the_same_file_starting_at_Colon_1, errorCount, firstFileReference) :',
    '            createCompilerDiagnostic(Diagnostics.Found_0_errors_in_1_files, errorCount, distinctFileNamesWithLines.length);',
    '    }',
].join('\r\n');

const owners = source.statements.filter(statement =>
    ts.isFunctionDeclaration(statement) && statement.name?.text === 'getErrorSummaryText');
assert.equal(owners.length, 1, 'expected one getErrorSummaryText');
assert(owners[0].body, 'getErrorSummaryText must have a body');

const statements = owners[0].body.statements;
function bindingIndex(names, flags) {
    const matches = statements.flatMap((statement, index) =>
        ts.isVariableStatement(statement) && statement.declarationList.declarations.length === 1 &&
        (statement.declarationList.flags & flags) !== 0 &&
        names.includes(statement.declarationList.declarations[0].name.getText(source)) ? [index] : []);
    assert.equal(matches.length, 1, `expected one local binding from ${names.join(', ')}`);
    return matches[0];
}

const startIndex = bindingIndex(['messageAndArgs', 'd'], ts.NodeFlags.Let);
const suffixIndex = bindingIndex(['suffix'], ts.NodeFlags.Const);
assert(ts.isIfStatement(statements[startIndex + 1]), 'expected the diagnostic selector after its binding');
const start = statements[startIndex].getStart(source);
const end = statements[suffixIndex - 1].end;
const actual = text.slice(start, end);
assert(actual === original || actual === direct, 'unreviewed getErrorSummaryText diagnostic selector');

// The pinned selector owns every reference to the original tuple local.
function checkTupleReferences(node) {
    if (ts.isIdentifier(node) && node.text === 'messageAndArgs') {
        assert(node.getStart(source) >= start && node.end <= end, 'messageAndArgs escaped the selector');
    }
    ts.forEachChild(node, checkTupleReferences);
}
checkTupleReferences(owners[0].body);

if (actual === original) {
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'watch.ts changed during adaptation');
    fs.writeFileSync(file, text.slice(0, start) + direct + text.slice(end));
}
console.log(JSON.stringify({ file: 'src/compiler/watch.ts', owner: 'getErrorSummaryText', edited: actual === original }));
