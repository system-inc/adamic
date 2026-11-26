'use strict';

const assert = require('node:assert/strict');
const {execFileSync, spawnSync} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

const [tree] = process.argv.slice(2);
assert(tree && process.argv.length === 3, 'Usage: node test.cjs ADAPTED_TREE');

const source = 'src/compiler/resolutionCache.ts';
const namespaceSource = 'src/testRunner/unittests/canWatch.ts';
const driftSource = 'src/compiler/zzDrift.ts';
const ref = 'refs/stage3/99-value-of-type-adaptations';
const adapter = path.join(__dirname, 'adapt.cjs');

function snapshot(side, file = source) {
    return execFileSync('git', ['-C', tree, 'show', `${ref}-${side}:${file}`]);
}

function append(input, line) {
    const text = input.toString('utf8');
    const newline = text.match(/\r?\n/)[0];
    return Buffer.concat([input, Buffer.from(`${text.endsWith(newline) ? '' : newline}${line}${newline}`)]);
}

function check(label, input, output, diagnostic, drift, driftPath = driftSource) {
    const scratch = fs.mkdtempSync('/tmp/adamic-watch-exports-');
    try {
        const file = path.join(scratch, source);
        fs.mkdirSync(path.dirname(file), {recursive: true});
        fs.writeFileSync(file, input);
        // A fixed old mtime detects a write even if the candidate produces identical bytes.
        fs.utimesSync(file, 1000000000, 1000000000);
        const mtime = fs.statSync(file, {bigint: true}).mtimeNs;
        const extras = [[path.join(scratch, namespaceSource), namespaceCall]];
        if (drift) extras.push([path.join(scratch, driftPath), Buffer.from(drift)]);
        for (const [extra, bytes] of extras) {
            fs.mkdirSync(path.dirname(extra), {recursive: true});
            fs.writeFileSync(extra, bytes);
            fs.utimesSync(extra, 1000000000, 1000000000);
        }
        const child = spawnSync(process.execPath, [adapter, scratch], {encoding: 'utf8'});
        if (child.error) throw child.error;
        if (diagnostic) {
            assert.equal(child.status, 1, `${label}: expected refusal, got ${child.status}`);
            const matches = Array.isArray(diagnostic)
                ? child.stderr.split('\n').some(line => line.startsWith(diagnostic[0]) && line.endsWith(diagnostic[1]))
                : child.stderr.includes(diagnostic);
            assert(matches, `${label}: ${child.stderr}`);
        } else {
            assert.equal(child.status, 0, `${label}: ${child.stderr}`);
        }
        assert(fs.readFileSync(file).equals(output), `${label}: unexpected file bytes`);
        if (output === input) {
            assert.equal(fs.statSync(file, {bigint: true}).mtimeNs, mtime, `${label}: mtime changed`);
        }
        for (const [extra, bytes] of extras) {
            assert(fs.readFileSync(extra).equals(bytes), label + ': external fixture bytes changed');
            assert.equal(fs.statSync(extra, {bigint: true}).mtimeNs, mtime, label + ': external fixture mtime changed');
        }
        console.log(`PASS ${label}`);
    } finally {
        fs.rmSync(scratch, {recursive: true, force: true});
    }
}

const before = snapshot('before');
const after = snapshot('after');
const namespaceCall = snapshot('before', namespaceSource);
check('a: clean input writes exact after snapshot', before, after);
check('b: replay preserves bytes and mtime', after, after);

for (const [row, owner, text, binding = text] of [
    ['c', 'isInDirectoryPath', 'export { isInDirectoryPath };', 'export { isInDirectoryPath }'],
    ['d', 'perceivedOsRootLengthForWatching', 'export { perceivedOsRootLengthForWatching as rootLength };',
        'export { perceivedOsRootLengthForWatching as rootLength }'],
    ['d', 'canWatchDirectoryOrFile', 'export { canWatchDirectoryOrFile as canWatch };',
        'export { canWatchDirectoryOrFile as canWatch }'],
    ['e', 'isInDirectoryPath', 'export default isInDirectoryPath;'],
    ['f', 'isInDirectoryPath', 'export = isInDirectoryPath;'],
]) {
    for (const [state, snapshot] of [['clean', before], ['adapted', after]]) {
        const input = append(snapshot, text);
        check(`${row}: ${state} ${text}`, input, input, `${owner}: unexpected export binding ${binding}`);
    }
}


function referenceDiagnostic(owner, statement, file = '') {
    return ['Error: ' + owner + ': reference outside a call at ' + file, ' in ' + statement];
}

const states = [['clean', before], ['adapted', after]];
for (const [row, owner, text, statement = text] of [
    ['g', 'perceivedOsRootLengthForWatching', 'export const rootLength = perceivedOsRootLengthForWatching;'],
    ['h', 'isInDirectoryPath', 'export const watchHelpers = { isInDirectoryPath };'],
    ['i', 'perceivedOsRootLengthForWatching', 'export function heldRoot() { return perceivedOsRootLengthForWatching; }',
        'return perceivedOsRootLengthForWatching;'],
    ['j', 'canWatchDirectoryOrFile', 'export type WatchCheck = typeof canWatchDirectoryOrFile;'],
    ['n', 'isInDirectoryPath', 'export const held = Array.of(isInDirectoryPath);'],
]) {
    for (const [state, snapshot] of states) {
        const input = append(snapshot, text);
        check(row + ': ' + state + ' ' + text, input, input, referenceDiagnostic(owner, statement));
    }
}

const prefixed = 'export function prefixed(a: readonly Path[], b: readonly Path[]) { return isInDirectoryPath(a, b); }';
const adaptedPrefix = append(after, prefixed);
check('k: clean direct call', append(before, prefixed), adaptedPrefix);
check('k: adapted direct call', adaptedPrefix, adaptedPrefix);

const namedImport = 'import { canWatchDirectoryOrFile } from "./_namespaces/ts.js";\n';
const namespaceImport = 'import * as ts from "./_namespaces/ts.js";\n';
const externalCall = 'import { canWatchDirectoryOrFile, Path } from "./_namespaces/ts.js";\n' +
    'export function canWatchPath(p: readonly Path[]) { return canWatchDirectoryOrFile(p); }\n';
for (const [row, prefix, statement] of [
    ['l', namedImport, 'export const canWatch = canWatchDirectoryOrFile;'],
    ['o', namespaceImport, 'export const w = ts.canWatchDirectoryOrFile;'],
]) {
    for (const [state, snapshot] of states) {
        check(row + ': ' + state + ' ' + statement, snapshot, snapshot,
            referenceDiagnostic('canWatchDirectoryOrFile', statement, driftSource + ':'), prefix + statement + '\n');
    }
}
check('m: clean cross-file direct call', before, after, undefined, externalCall);
check('m: adapted cross-file direct call', after, after, undefined, externalCall);


const ts = require('typescript');
const ownerPins = require('./owners.json');
const ownerMembers = {
    perceivedOsRootLengthForWatching: ['pathComponents'],
    canWatchDirectoryOrFile: ['pathComponents'],
    isInDirectoryPath: ['dirComponents', 'fileOrDirComponents'],
};
assert.deepEqual(Object.keys(ownerPins), Object.keys(ownerMembers));
for (const [name, members] of Object.entries(ownerMembers)) {
    assert.deepEqual(Object.keys(ownerPins[name]), ['before', 'after']);
    for (const [side, input] of [['before', before], ['after', after]]) {
        const parsed = ts.createSourceFile(source, input.toString('utf8'), ts.ScriptTarget.Latest, true);
        const declarations = parsed.statements.filter(statement => ts.isFunctionDeclaration(statement) &&
            statement.name?.text === name && statement.body);
        assert.equal(declarations.length, 1, name + ' declaration count');
        assert.equal(declarations[0].getText(parsed), ownerPins[name][side], name + ' ' + side + ' text');
    }
    let expected = ownerPins[name].before;
    for (const member of members) {
        const original = member + ': Readonly<PathPathComponents>';
        assert.equal(expected.split(original).length, 2, name + '.' + member);
        expected = expected.replace(original, member + ': readonly Path[]');
    }
    assert.equal(expected, ownerPins[name].after, name + ' annotation-only edit');
}
console.log('PASS p: exact owner text pins in both refs');

const prefixOwner = 'isInDirectoryPath';
const countDiagnostic = ['Error: ' + prefixOwner + ': expected one declaration statement, found 2', ''];
const overload = 'function isInDirectoryPath(dirComponents: readonly Path[], fileOrDirComponents: readonly Path[]): readonly Path[];';
for (const [state, snapshot] of states) {
    const side = state === 'clean' ? 'before' : 'after';
    const pinned = ownerPins[prefixOwner][side];
    const text = snapshot.toString('utf8');
    assert.equal(text.split(pinned).length, 2, prefixOwner + ' snapshot occurrence');
    assert.equal(pinned.split('return true;').length, 2, prefixOwner + ' final return occurrence');
    assert(pinned.endsWith('return true;\r\n}'));
    const changedOwner = pinned.replace('return true;', 'return dirComponents;');
    const bodyDrift = Buffer.from(text.replace(pinned, () => changedOwner));
    check('q: ' + state + ' owner body drift', bodyDrift, bodyDrift,
        ['Error: ' + prefixOwner + ': owner text differs from its pinned ' + side + ' text', '']);

    const newline = text.match(/\r?\n/)[0];
    const overloaded = Buffer.from(text.replace(pinned, () => overload + newline + pinned));
    check('r: ' + state + ' overload', overloaded, overloaded, countDiagnostic);

    const merged = append(snapshot, 'namespace isInDirectoryPath { export const prefix = 1; }');
    check('s: ' + state + ' namespace merge', merged, merged, countDiagnostic);

    for (const suffix of ['mts', 'cts', 'tsx']) {
        const file = 'src/compiler/zzDrift.' + suffix;
        check('t: ' + state + ' ' + file, snapshot, snapshot,
            ['Error: canWatchDirectoryOrFile: unscanned source file ' + file, ''], 'export {};\n', file);
    }
}
