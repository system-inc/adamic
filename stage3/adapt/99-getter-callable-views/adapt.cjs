'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const sites = require('./sites.json');
const snapshot = text => text.replace(/\r\n/g, '\n');

function planFile(tree, file, pins) {
    const filename = path.join(tree, file);
    const raw = fs.readFileSync(filename, 'utf8');
    assert(!/\r(?!\n)/.test(raw), `${file}: unsupported bare CR`);
    const text = raw;
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse diagnostics`);

    const matches = new Map(pins.map(pin => [pin, []]));
    function visit(node) {
        if (ts.isStatement(node) && ts.isBlock(node.parent)) {
            const statement = snapshot(node.getText(source));
            for (const pin of pins) {
                if (statement === pin.before || statement === pin.after) {
                    matches.get(pin).push({ node, adapted: statement === pin.after });
                }
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);

    const edits = [];
    const states = [];
    for (const [pin, candidates] of matches) {
        assert.equal(candidates.length, 1, `${pin.site}: missing, duplicate or drifted statement`);
        const { node, adapted } = candidates[0];
        states.push(adapted);
        const calls = [];
        function findCall(child) {
            if (ts.isCallExpression(child) && ts.isPropertyAccessExpression(child.expression) &&
                child.expression.name.text === pin.member &&
                child.expression.expression.getText(source) === (adapted ? pin.view : pin.receiver)) {
                calls.push(child.expression.expression);
            }
            ts.forEachChild(child, findCall);
        }
        findCall(node);
        assert.equal(calls.length, 1, `${pin.site}: expected one matching call receiver`);

        const start = node.getStart(source);
        const indentation = text.slice(text.lastIndexOf('\n', start - 1) + 1, start);
        assert(/^[ \t]+$/.test(indentation), `${pin.site}: expected an indented standalone statement`);
        if (adapted) {
            const siblings = node.parent.statements;
            const declaration = siblings[siblings.indexOf(node) - 1];
            assert.equal(declaration?.getText(source), pin.declaration, `${pin.site}: adapted view declaration drift`);
            continue;
        }
        const receiver = calls[0];
        const rewritten = text.slice(start, receiver.getStart(source)) + pin.view + text.slice(receiver.end, node.end);
        assert.equal(snapshot(rewritten), pin.after, `${pin.site}: rewritten statement mismatch`);
        const newline = text[text.lastIndexOf('\n', start - 1) - 1] === '\r' ? '\r\n' : '\n';
        edits.push({ start, end: node.end, replacement: pin.declaration + newline + indentation + rewritten });
    }

    let result = text;
    let boundary = text.length;
    for (const { start, end, replacement } of edits.sort((a, b) => b.start - a.start)) {
        assert(end <= boundary, `${file}: overlapping pins`);
        result = result.slice(0, start) + replacement + result.slice(end);
        boundary = start;
    }
    assert.equal(ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    return { filename, raw, result, states };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const files = Map.groupBy(sites, pin => pin.site.replace(/:\d+:\d+$/, ''));
    const plans = [...files].map(([file, pins]) => planFile(tree, file, pins));
    const states = plans.flatMap(plan => plan.states);
    assert(states.length === 13 && states.every(state => state === states[0]), 'half-adapted tree');
    for (const { filename, raw } of plans) {
        assert.equal(fs.readFileSync(filename, 'utf8'), raw, `${filename}: changed during planning`);
    }
    const changed = plans.filter(plan => plan.result !== plan.raw);
    for (const { filename, result } of changed) fs.writeFileSync(filename, result);
    return { files: changed.length, sites: changed.length ? states.length : 0, javascriptChanging: changed.length > 0 };
}

assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
