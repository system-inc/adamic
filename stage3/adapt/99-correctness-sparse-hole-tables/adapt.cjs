#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const pins = require('./pins.json');

assert.equal(ts.version, '6.0.3');

const lf = text => text.replaceAll('\r\n', '\n');

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file + ': parse diagnostics');
    return source;
}

function owner(node) {
    const names = [];
    for (let parent = node.parent; parent; parent = parent.parent) {
        if (ts.isFunctionLike(parent) && parent.name) names.unshift(parent.name.getText());
    }
    return names;
}

function planFile(file, raw, restore) {
    assert(!/\r(?!\n)/.test(raw), file + ': bare CR');
    const source = parse(file, raw);
    const matches = new Map(pins.filter(pin => pin.file === file).map(pin => [pin, []]));
    function visit(node) {
        for (const [pin, found] of matches) {
            if (pin.kind === 'declaration' && ts.isVariableDeclaration(node) && node.name.getText() === pin.member &&
                JSON.stringify(owner(node)) === JSON.stringify(pin.owner)) {
                found.push(node);
            }
            if (pin.kind === 'writer' && ts.isBinaryExpression(node) &&
                node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isElementAccessExpression(node.left) &&
                node.left.expression.getText() === pin.member && node.left.argumentExpression.getText() === pin.key &&
                JSON.stringify(owner(node)) === JSON.stringify(pin.owner)) {
                let statement = node;
                while (!ts.isStatement(statement)) statement = statement.parent;
                assert(ts.isBlock(statement.parent), file + ': writer is outside a block');
                found.push(statement);
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);

    const edits = [];
    const states = [];
    for (const [pin, found] of matches) {
        const label = [file, ...pin.owner, pin.member].join(':');
        assert.equal(found.length, 1, label + ': missing or duplicate member');
        const node = found[0];
        const start = node.getStart(source);
        const text = raw.slice(start, node.end);
        assert.equal(text.includes('\r\n') && /(?<!\r)\n/.test(text), false, label + ': mixed member endings');
        if (pin.kind === 'declaration') {
            assert([pin.before, pin.after].includes(lf(text)), label + ': declaration drift');
            if (pin.before === pin.after) continue;
            const adapted = lf(text) === pin.after;
            states.push(adapted);
            if (restore === adapted) edits.push({start, end: node.end, text: restore ? pin.before : pin.after});
            continue;
        }

        assert.equal(lf(text), pin.before, label + ': writer statement drift');
        const line = raw.lastIndexOf('\n', start - 1) + 1;
        const indent = raw.slice(line, start);
        assert(/^ *$/.test(indent), label + ': unexpected writer indentation');
        const newline = raw[line - 2] === '\r' ? '\r\n' : '\n';
        const fill = pin.fill.split('\n').map(row => indent + row).join(newline) + newline;
        const previous = node.parent.statements[node.parent.statements.indexOf(node) - 1];
        const adapted = raw.slice(line - fill.length, line) === fill;
        if (previous && !adapted) {
            const priorText = previous.getText(source);
            assert(!priorText.includes(pin.member + '.push(undefined)'), label + ': prior fill drift');
        }
        states.push(adapted);
        if (restore && adapted) edits.push({start: line - fill.length, end: line, text: ''});
        if (!restore && !adapted) edits.push({start: line, end: line, text: fill});
    }
    let output = raw;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        output = output.slice(0, edit.start) + edit.text + output.slice(edit.end);
    }
    if (output !== raw) parse(file, output);
    return {file, raw, output, states, count: edits.length};
}

function plan(raws, restore) {
    const plans = Object.entries(raws).map(([file, raw]) => planFile(file, raw, restore));
    const states = plans.flatMap(row => row.states);
    assert(states.length && states.every(state => state === states[0]), 'partly adapted tree');
    return plans;
}

function rewrite(tree, {restore = false} = {}) {
    const files = [...new Set(pins.map(pin => pin.file))];
    const raws = Object.fromEntries(files.map(file => [file, fs.readFileSync(path.join(tree, file), 'utf8')]));
    const plans = plan(raws, restore);
    const changed = plans.filter(row => row.raw !== row.output);
    const planned = Object.fromEntries(plans.map(row => [row.file, row.output]));
    assert(plan(planned, restore).every(row => row.raw === row.output), 'planned rewrite is not idempotent');
    for (const file of files) assert.equal(fs.readFileSync(path.join(tree, file), 'utf8'), raws[file]);
    for (const row of changed) fs.writeFileSync(path.join(tree, row.file), row.output);
    return {files: changed.length, edits: changed.reduce((n, row) => n + row.count, 0), javascriptChanging: true};
}

module.exports = {rewrite};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify({typescript: ts.version, task: '#s59veah', ...rewrite(path.resolve(process.argv[2]))}));
}
