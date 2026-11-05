'use strict';
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const assert = require('node:assert/strict');
const adapter = require('./adapt.cjs');
const pins = require('./pins.json');
const before = fs.readFileSync(path.join(process.argv[2], 'src/compiler/debug.ts'), 'utf8');
const root = fs.mkdtempSync(path.join(os.tmpdir(), '99n-contract-'));
const results = [];
function fixture(name, text, refused) {
    const tree = path.join(root, name), file = path.join(tree, 'src/compiler/debug.ts');
    fs.mkdirSync(path.dirname(file), {recursive: true});
    fs.writeFileSync(file, text);
    const original = fs.readFileSync(file);
    let writes = 0;
    const write = fs.writeFileSync;
    fs.writeFileSync = (...args) => { writes++; return write(...args); };
    let result, error;
    try { result = adapter.adapt(tree); } catch (e) { error = e; }
    finally { fs.writeFileSync = write; }
    if (refused) {
        assert(error, name + ' was accepted');
        assert.equal(writes, 0, name + ' wrote before refusing');
        assert.deepEqual(fs.readFileSync(file), original);
        results.push({name, refused: true, writes, reason: error.message});
    } else {
        assert.ifError(error);
        assert.equal(writes, 1);
        const after = fs.readFileSync(file), stamp = fs.statSync(file, {bigint: true}).mtimeNs;
        assert.deepEqual(adapter.adapt(tree), {files: 0, edits: 0, javascriptChanging: false});
        assert.deepEqual(fs.readFileSync(file), after);
        assert.equal(fs.statSync(file, {bigint: true}).mtimeNs, stamp);
        if (name === 'foreign') assert(fs.readFileSync(file, 'utf8').startsWith('// unrelated foreign edit\r\n'));
        results.push({name, first: result, replay: {files: 0, edits: 0}, unchangedReplayMtime: true});
    }
}
fixture('apply', before, false);
fixture('foreign', '// unrelated foreign edit\r\n' + before, false);
fixture('changed-owner', before.replace('function hasNode(f: FlowNode)', 'function hasNode(f:  FlowNode)'), true);
fixture('half-adapted', before.replace('function hasNode(f: FlowNode)', 'function hasNode(f: Readonly<FlowNode>)'), true);
const after = adapter.plan(before).after;
fixture('changed-after-owner', after.replace('function hasNode(f: Readonly<FlowNode>)', 'function hasNode(f:  Readonly<FlowNode>)'), true);
console.log(JSON.stringify({results}, null, 2));
fs.rmSync(root, {recursive: true});
