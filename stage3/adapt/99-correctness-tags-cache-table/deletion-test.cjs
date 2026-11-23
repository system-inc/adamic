'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');

const [tree, adapter = path.join(__dirname, 'adapt.cjs')] = process.argv.slice(2);
assert(tree, 'pass a full applied tree containing the adapter checkpoint refs');
assert.equal(ts.version, '6.0.3');
const { adapt } = require(path.resolve(adapter));
const files = [
    'src/compiler/types.ts', 'src/compiler/utilities.ts', 'src/compiler/utilitiesPublic.ts',
    'src/compiler/parser.ts', 'src/compiler/jsDocCache.ts',
];
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'jsdoc-deletion-test-'));
const prefix = 'refs/stage3/99-correctness-tags-cache-table-';
const types = path.join(scratch, files[0]);

function load(side, normalize = false) {
    for (const file of files) {
        const filename = path.join(scratch, file);
        fs.mkdirSync(path.dirname(filename), { recursive: true });
        if (side === 'before' && file === files.at(-1)) {
            fs.rmSync(filename, { force: true });
            continue;
        }
        const bytes = execFileSync('git', ['-C', tree, 'show', prefix + side + ':' + file], { encoding: 'utf8' });
        fs.writeFileSync(filename, normalize ? bytes.replace(/\r\n/g, '\n') : bytes);
    }
}

function snapshot() {
    return files.map(file => fs.readFileSync(path.join(scratch, file)));
}

function keepsLocalsContainer() {
    const { outputText, diagnostics } = ts.transpileDeclaration(fs.readFileSync(types, 'utf8'), {
        fileName: 'types.ts', compilerOptions: { stripInternal: true },
    });
    assert.equal(diagnostics.length, 0, 'declaration stripping diagnostics');
    return /^export interface LocalsContainer extends Node\b/m.test(outputText);
}

try {
    load('after');
    const clean = snapshot();
    assert.equal(adapt(scratch).files, 0);
    assert.deepEqual(snapshot(), clean);
    assert(keepsLocalsContainer(), 'clean declaration stripping must keep LocalsContainer');
    console.log('Clean after-checkpoint: five unchanged files; declaration stripping keeps LocalsContainer.');

    load('before');
    assert.equal(adapt(scratch).files, 5);
    assert.deepEqual(snapshot(), clean);
    console.log('Before-checkpoint application reproduces all five after-checkpoint files.');

    load('after', true);
    execFileSync('git', ['apply', path.join(__dirname, 'orphan-comment.patch')], { cwd: scratch });
    assert(!keepsLocalsContainer(), 'orphan control must hide LocalsContainer from declaration stripping');
    const orphan = snapshot();
    const writeFileSync = fs.writeFileSync;
    let writes = 0;
    fs.writeFileSync = (...args) => {
        writes++;
        return writeFileSync(...args);
    };
    try {
        assert.throws(() => adapt(scratch), /JSDocArray: drift or half-adapted deletion span/);
    } finally {
        fs.writeFileSync = writeFileSync;
    }
    assert.equal(writes, 0);
    assert.deepEqual(snapshot(), orphan);
    console.log('Orphaned comment hides LocalsContainer and the adapter refuses before writing.');
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
