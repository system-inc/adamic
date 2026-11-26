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
    const values = Array.from({length: 4}, (_, ordinal) => ({ordinal}));
    function run([decl, write, read]) {
        let calls = 0;
        const factory = {createUniqueName() {return values[calls++];}};
        const init = `const ${decl}; let contextObject;
            function write(id) { ${write} return contextObject; }
            function read(id) { ${read} return contextObject; }
            [write, read]`;
        const [put, get] = vm.runInNewContext(js(init), {factory});
        const seen = [];
        for (const [turn,id] of [7,2,40,7].entries()) {
            const current = put(id);
            assert(Object.is(current, get(id)), 'stored and local identifier identity');
            assert.equal(calls, turn+1, 'one identifier per write');
            assert(Object.is(current, values[turn]), 'factory result identity');
            seen.push(get(id));
        }
        assert.equal(get(8), undefined, 'unwritten numeric key');
        seen.push(get(7),get(2),get(40));
        return seen;
    }
    const expected = run(before);
    const actual = run(after);
    assert.deepEqual(actual.map(v => v.ordinal), [0,1,2,3,3,1,2]);
    actual.forEach((item,i) => assert(Object.is(item, expected[i]), 'sparse lookup identity'));
    console.log('PASS sparse ids, absent key, overwrite, call count and identifier identity');
}
