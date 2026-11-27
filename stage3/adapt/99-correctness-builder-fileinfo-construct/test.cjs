'use strict';
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const {adapt, owner} = require('./adapt.cjs');
const [pin] = require('./owners.json');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
const tree = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-correctness-builder-fileinfo-construct';
const read = side => execFileSync('git', ['-C', tree, 'show', ref + '-' + side + ':' + pin.file], {encoding: 'utf8'});
const before = read('before');
const after = read('after');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-fileinfo-'));
const builder = path.join(scratch, pin.file);
const four = ['version', 'signature', 'affectsGlobalScope', 'impliedFormat'];
let failures = 0;

function row(name, check) {
    try {
        const detail = check();
        console.log('PASS ' + name + (detail === undefined ? '' : ' ' + detail));
    } catch (error) {
        failures++;
        console.error('FAIL ' + name + ': ' + error.message);
    }
}

// The owner as Node runs it: its text from a snapshot, transpiled alone, with tsc's isString.
function compiled(text) {
    const source = owner(text).replace(/^export /, '');
    const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.None}}).outputText;
    const isString = value => typeof value === 'string';
    return new Function('isString', output + '\nreturn toBuilderStateFileInfoForMultiEmit;')(isString);
}

try {
    fs.mkdirSync(path.dirname(builder), {recursive: true});
    row('control writes exact after snapshot and replay does not write', () => {
        fs.writeFileSync(builder, before);
        assert.equal(adapt(scratch).writes, 1, 'expected one write');
        assert(fs.readFileSync(builder, 'utf8') === after, 'unexpected output bytes');
        assert.equal(adapt(scratch).writes, 0, 'expected replay to write nothing');
        assert(fs.readFileSync(builder, 'utf8') === after, 'unexpected replay bytes');
    });

    const drift = /drift or half-adapted owner/;
    const newline = before.includes('\r\n') ? '\r\n' : '\n';
    const cases = [
        ['owner changed away from site', before.replace(pin.before, () => pin.before.replace('export function', 'export  function')), drift],
        ['site expression changed', before.replace('fileInfo as BuilderState.FileInfo :', 'fileInfo as unknown as BuilderState.FileInfo :'), drift],
        ['duplicate toBuilderStateFileInfoForMultiEmit', before + newline + pin.before + newline, /missing or duplicate owner/],
        ['bare carriage return', before + '\r', /bare CR/],
        ['mixed line endings', before + (newline === '\n' ? '\r\n' : '\n'), /mixed line endings/],
    ];
    for (const [name, text, message] of cases) {
        row('refuses ' + name, () => {
            assert(text !== before, 'fixture did not change its input');
            fs.writeFileSync(builder, text);
            assert.throws(() => adapt(scratch), message);
            assert(fs.readFileSync(builder, 'utf8') === text, 'refusal wrote builder.ts');
        });
    }

    row('public declaration emit unchanged', () => {
        const emit = text => ts.transpileDeclaration(text, {
            fileName: pin.file, compilerOptions: {stripInternal: true},
        }).outputText;
        const declarations = emit(before);
        assert(declarations.length > 0 && declarations === emit(after), 'public declaration bytes changed');
    });

    // The witnesses: each branch, and on the cast branch every key order the lane oracle hands it (J-4dma35b:
    // version,signature 5,051; with affectsGlobalScope 433; with impliedFormat 61; all four 3).
    const oldFunction = compiled(before);
    const newFunction = compiled(after);
    const inputs = [
        ['string', () => 'v1', 'string'],
        ['cast, version and signature', () => ({version: 'v1', signature: 's1'}), 'cast'],
        ['cast, with affectsGlobalScope', () => ({version: 'v1', signature: 's1', affectsGlobalScope: true}), 'cast'],
        ['cast, with impliedFormat', () => ({version: 'v1', signature: 's1', impliedFormat: 99}), 'cast'],
        ['cast, all four', () => ({version: 'v1', signature: 's1', affectsGlobalScope: true, impliedFormat: 1}), 'cast'],
        ['built, signature false', () => ({version: 'v1', signature: false}), 'built'],
        ['built, signature absent', () => ({version: 'v1', affectsGlobalScope: true}), 'built'],
    ];
    for (const [name, make, branch] of inputs) {
        row('witness ' + name, () => {
            const oldInput = make();
            const newInput = make();
            const oldResult = oldFunction(oldInput);
            const newResult = newFunction(newInput);
            for (const key of four) assert(Object.is(newResult[key], oldResult[key]), `${key}: ${String(oldResult[key])} before, ${String(newResult[key])} after`);
            assert.equal(JSON.stringify(newResult), JSON.stringify(oldResult), 'serialized result differs');
            assert.deepEqual(newInput, make(), 'the input changed');
            if (branch === 'cast') {
                assert(oldResult === oldInput, 'before: the cast branch should hand the input back');
                assert(newResult !== newInput, 'after: the cast branch should build a fresh object');
                assert.deepEqual(Object.keys(newResult), four, 'after: the fresh object should carry the four keys in order');
                newResult.signature = 'written';
                assert.equal(newInput.signature, 's1', 'a write to the fresh object reached the build info');
            }
            else {
                assert.deepEqual(Object.keys(newResult), Object.keys(oldResult), 'key order differs on an unchanged branch');
            }
            return `${branch}: before keys ${Object.keys(oldResult).join(',')}, after keys ${Object.keys(newResult).join(',')}`;
        });
    }
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}
process.exitCode = Number(failures !== 0);
