'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const pins = require('./pins.json');
const runtimePins = require('./runtime-pins.json');
const { locate } = require('../52-json-values/adapt.cjs');

function plan(file, input, rules) {
    const source = ts.createSourceFile(file, input, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file);
    const edits = [];
    const states = [];
    for (const rule of rules) {
        const matches = [];
        function visit(node) {
            if (ts.SyntaxKind[node.kind] === rule.kind && ts.isIdentifier(node.name) && node.name.text === rule.binding) {
                let owner = node.parent;
                while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
                if (owner?.name?.text === rule.owner && owner.parent === source) matches.push(node);
            }
            ts.forEachChild(node, visit);
        }
        visit(source);
        assert.equal(matches.length, 1, `${file}: ${rule.owner}.${rule.binding} binding count`);
        const [node] = matches;
        const text = node.getText(source);
        assert(text === rule.before || text === rule.after, `${file}: ${rule.owner}.${rule.binding} snapshot drift`);
        states.push(text === rule.after);
        if (text === rule.before) edits.push({ start: node.getStart(source), end: node.end, value: rule.after });
    }
    let output = input;
    for (const { start, end, value } of edits.sort((a, b) => b.start - a.start)) {
        output = output.slice(0, start) + value + output.slice(end);
    }
    assert.equal(ts.createSourceFile(file, output, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, file);
    if (edits.length) {
        const compilerOptions = { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext };
        const emit = text => ts.transpileModule(text, { fileName: file, compilerOptions }).outputText;
        assert.equal(emit(output), emit(input), `${file}: emitted JavaScript changed`);
    }
    return { output, states, edits: edits.length };
}

function planRuntime(file, input, rules) {
    const edits = [];
    const states = [];
    for (const rule of rules) {
        const found = locate(file, input, rule);
        states.push(found.state === 'after');
        if (found.state === 'before') edits.push({ start: found.start, end: found.start + rule.before.length, value: rule.after });
    }
    let output = input;
    let boundary = input.length;
    for (const { start, end, value } of edits.sort((a, b) => b.start - a.start)) {
        assert(end <= boundary, `${file}: overlapping runtime edits`);
        boundary = start;
        output = output.slice(0, start) + value + output.slice(end);
    }
    for (const rule of rules) assert.equal(locate(file, output, rule).state, 'after');
    return { output, states, edits: edits.length };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const files = [...new Set([...pins, ...runtimePins].map(rule => rule.file))];
    const plans = files.map(file => {
        const location = path.join(tree, file);
        const input = fs.readFileSync(location, 'utf8');
        const types = plan(file, input, pins.filter(rule => rule.file === file));
        const runtime = planRuntime(file, types.output, runtimePins.filter(rule => rule.file === file));
        return { location, input, output: runtime.output, states: [...types.states, ...runtime.states], edits: types.edits + runtime.edits };
    });
    assert.equal(new Set(plans.flatMap(result => result.states)).size, 1, 'partly adapted any values');
    for (const result of plans) {
        assert.equal(fs.readFileSync(result.location, 'utf8'), result.input, 'source changed during planning');
    }
    for (const result of plans) {
        if (result.edits) fs.writeFileSync(result.location, result.output);
    }
    return { files: plans.filter(result => result.edits).length, edits: plans.reduce((count, result) => count + result.edits, 0) };
}

module.exports = { adapt, plan, planRuntime };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
