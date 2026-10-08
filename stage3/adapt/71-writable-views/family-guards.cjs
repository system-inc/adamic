#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const child = require('node:child_process');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const tree = path.resolve(process.argv[2]);
function hashes(root) {
    const result = {};
    function walk(dir) {
        for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
            const file = path.join(dir, entry.name);
            if (entry.isDirectory()) walk(file);
            else result[path.relative(root, file)] = crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
        }
    }
    walk(path.join(root, 'src'));
    return result;
}
const cases = [
    { family: 'diagnostics', owner: 'convertToJson', statement: 'errors.push(undefined);' },
    { family: 'tuples', owner: 'getTypeArguments', statement: 'type.target = type.target;' },
    { family: 'empty-arrays', owner: 'addAntecedent', statement: 'label.antecedent = [label];' },
];
for (const test of cases) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'unit71-family-guard-'));
    try {
        fs.cpSync(path.join(tree, 'src'), path.join(scratch, 'src'), { recursive: true });
        fs.symlinkSync(path.join(tree, 'node_modules'), path.join(scratch, 'node_modules'));
        const owner = require(`./${test.family}.json`).owners.find(o => o.function === test.owner);
        const file = path.join(scratch, owner.file), text = fs.readFileSync(file, 'utf8');
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.ES2024, true);
        const matches = [];
        function visit(n) {
            if (ts.isFunctionDeclaration(n) && n.name?.text === test.owner && n.body) matches.push(n.body);
            ts.forEachChild(n, visit);
        }
        visit(source);
        assert.equal(matches.length, 1);
        const position = matches[0].getStart(source) + 1;
        fs.writeFileSync(file, text.slice(0, position) + test.statement + text.slice(position));
        const before = hashes(scratch);
        const result = child.spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], { encoding: 'utf8' });
        assert.equal(result.status, 1);
        assert(result.stderr.includes(`reviewed runtime body changed: ${test.owner}`), result.stderr);
        assert.deepEqual(hashes(scratch), before);
        console.log(JSON.stringify({ ...test, exit: result.status, source_unchanged_after_rejection: true }));
    }
    finally { fs.rmSync(scratch, { recursive: true }); }
}
