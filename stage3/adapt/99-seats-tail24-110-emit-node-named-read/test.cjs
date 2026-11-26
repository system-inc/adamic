'use strict';
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const {prepare, files} = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
assert.equal(ts.version, '6.0.3');
const tree = process.argv[2];
const ref = 'refs/stage3/' + path.basename(__dirname);
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'tail24-'));
const read = (side, file) => execFileSync('git', ['-C', tree, 'show', `${ref}-${side}:${file}`]);

function fixture(side) {
    const root = path.join(tmp, side);
    for (const file of files) {
        const output = path.join(root, file);
        fs.mkdirSync(path.dirname(output), {recursive: true});
        fs.writeFileSync(output, read(side, file));
    }
    return root;
}

function js(source) {
    return ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ESNext}}).outputText;
}

try {
    const before = fixture('before');
    const after = fixture('after');
    const original = prepare(before).flatMap(file => file.snapshots);
    const adapted = prepare(after).flatMap(file => file.snapshots);
    for (let run = 1; run <= 2; run++) {
        execFileSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), before]);
        for (const file of files) {
            assert.deepEqual(fs.readFileSync(path.join(before, file)), read('after', file), file);
        }
        console.log(`PASS apply ${run}: matches recorded after bytes`);
    }
    witness(original, adapted);
} finally {
    fs.rmSync(tmp, {recursive: true, force: true});
}

function witness(before, after) {
    const keys = ['flags', 'internalFlags', 'leadingComments', 'trailingComments', 'commentRange', 'sourceMapRange',
        'tokenSourceMapRanges', 'constantValue', 'helpers', 'startsOnNewLine', 'snippetElement', 'classThis', 'assignedName'];
    for (const [label, present, value] of [['absent', false], ['undefined', true, undefined], ['false', true, false], ['true', true, true]]) {
        function run(statements) {
            const reads = [];
            const source = Object.fromEntries(keys.filter(key => key !== 'startsOnNewLine').map(key => [key, key]));
            if (present) source.startsOnNewLine = value;
            const sourceEmitNode = new Proxy(source, {get(obj, key) { reads.push(key); return obj[key]; }});
            const values = vm.runInNewContext(js(statements.join('\n') + '\n[' + keys.join(',') + ']'), {sourceEmitNode});
            return {reads, values: Array.from(values)};
        }
        const expected = run(before);
        const actual = run(after);
        assert.deepEqual(actual.reads, keys, 'field read order');
        assert.deepEqual(actual, expected, 'bound values including false');
        assert.equal(actual.values[9], value);
        console.log(`PASS ${label}: read order and bound values agree`);
    }
}
