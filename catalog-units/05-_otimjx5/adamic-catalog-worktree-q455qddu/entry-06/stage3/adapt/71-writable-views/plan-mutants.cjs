#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const child = require('node:child_process');
const assert = require('node:assert/strict');
const tree = path.resolve(process.argv[2]);
function hashes(root) {
    const result = {};
    function visit(dir) {
        for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
            const file = path.join(dir, entry.name);
            if (entry.isDirectory()) visit(file);
            else result[path.relative(root, file)] = crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
        }
    }
    visit(path.join(root, 'src'));
    return result;
}
const edit = require('./empty-arrays.json').edits[0];
const cases = [
    { name: 'holder-type', holder: true, expected: 'fresh array holder type changed' },
    { name: 'runtime-copy', after: 'label.antecedent = ([] as FlowNode[]).slice()', expected: 'adaptation changes runtime syntax' },
    { name: 'any-type', after: 'label.antecedent = [] as any', expected: 'adaptation adds any or unknown' },
    { name: 'unknown-type', after: 'label.antecedent = [] as unknown', expected: 'adaptation adds any or unknown' },
];
for (const test of cases) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'unit71-plan-mutant-'));
    try {
        fs.cpSync(path.join(tree, 'src'), path.join(scratch, 'src'), { recursive: true });
        fs.symlinkSync(path.join(tree, 'node_modules'), path.join(scratch, 'node_modules'));
        const file = path.join(scratch, edit.file), text = fs.readFileSync(file, 'utf8');
        assert.equal(text.split(edit.after).length, 2);
        fs.writeFileSync(file, text.replace(edit.after, edit.before));
        if (test.holder) {
            const declaration = path.join(scratch, 'src/compiler/types.ts');
            const input = fs.readFileSync(declaration, 'utf8');
            const needle = 'export interface FlowLabel extends FlowNodeBase {\r\n    node: undefined;\r\n    antecedent: FlowNode[] | undefined;';
            assert.equal(input.split(needle).length, 2);
            fs.writeFileSync(declaration, input.replace(needle, needle.replace('FlowNode[]', 'Type[]')));
        }
        const before = hashes(scratch);
        // Mutate only the loaded plan object in the child, never checked-in metadata.
        const alteration = test.holder ? '' : `require(${JSON.stringify(path.join(__dirname, 'empty-arrays.json'))}).edits[0].after = ${JSON.stringify(test.after)};`;
        const script = `${alteration} process.argv = [process.execPath, ${JSON.stringify(path.join(__dirname, 'adapt.cjs'))}, ${JSON.stringify(scratch)}]; require(process.argv[1]);`;
        const result = child.spawnSync(process.execPath, ['-e', script], { encoding: 'utf8' });
        assert.equal(result.status, 1);
        assert(result.stderr.includes(test.expected), result.stderr);
        assert.deepEqual(hashes(scratch), before);
        console.log(JSON.stringify({ name: test.name, exit: result.status, guard: test.expected, source_unchanged: true }));
    }
    finally { fs.rmSync(scratch, { recursive: true }); }
}
