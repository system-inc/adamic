#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const {execFileSync, spawnSync} = require('node:child_process');
const {createHash} = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node test.cjs <fully-adapted-tree>');
const tree = path.resolve(process.argv[2]);
const adapter = path.join(__dirname, 'adapt.cjs');
const ref = 'refs/stage3/99-z-seats-build22-88-ordinal-comparator-';
const paths = ['src/compiler/sys.ts', 'src/compiler/watchUtilities.ts'];
const support = ['src/compiler/core.ts', 'src/compiler/_namespaces/ts.ts'];
function readSnapshot(side, files) {
    return files.map(file => execFileSync('git', ['-C', tree, 'show', ref + side + ':' + file]));
}
const input = readSnapshot('before', [...paths, ...support]).map(bytes => bytes.toString('utf8'));
const expected = readSnapshot('after', paths);
const comparator = '(a, b) => a < b ? -1 : a > b ? 1 : 0';
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'family88-'));
const files = paths.map(file => path.join(scratch, file));
const digest = file => createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const hashes = () => files.map(digest);
const outputBytes = () => files.map(file => fs.readFileSync(file));

function reset(sources) {
    for (const [index, name] of [...paths, ...support].entries()) {
        const file = path.join(scratch, name);
        fs.mkdirSync(path.dirname(file), {recursive: true});
        fs.writeFileSync(file, sources[index] ?? input[index]);
    }
}

function substitute(text, before, after) {
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    before = before.replaceAll('\n', newline);
    assert.equal(text.split(before).length, 2, 'expected one plant anchor: ' + before);
    return text.replace(before, after.replaceAll('\n', newline));
}

function runAdapter() {
    const run = spawnSync(process.execPath, [adapter, scratch], {encoding: 'utf8'});
    if (run.error) throw run.error;
    assert.equal(run.signal, null, 'adapter terminated by signal');
    return run;
}

function reject(label, sources, message) {
    reset(sources);
    const before = hashes();
    const run = runAdapter();
    const after = hashes();
    console.log(label + ': adapter exit ' + run.status);
    for (const [index, file] of paths.entries()) {
        console.log('  ' + file + ' ' + before[index] + ' -> ' + after[index]);
    }
    assert.notEqual(run.status, 0, label + ': adapter accepted the plant and changed ' +
        paths.filter((_, index) => before[index] !== after[index]).join(', '));
    assert.deepEqual(after, before, label + ': wrote source before rejecting');
    if (message) assert(run.stderr.includes(message), run.stderr);
    console.log('  ' + (run.stderr.split('\n').find(line => line.startsWith('AssertionError')) || run.stderr.trim()));
}

function accept() {
    const run = runAdapter();
    assert.equal(run.status, 0, run.stderr);
}

const samples = [
    ['empty', []],
    ['one', ['only']],
    ['case and UTF-16', ['zeta', 'alpha', 'Alpha', '\uE000', '\u{10000}']],
    ['duplicates', ['b', 'a', 'b', 'a']],
    ['empty name', ['b', '', 'a']],
    ['tail after surrogate pair', ['a\u{10000}z', 'a\u{10000}b']],
];

function compileStatement(source, prefix) {
    const lines = source.split(/\r?\n/).map(line => line.trim()).filter(line => line.startsWith(prefix));
    assert.equal(lines.length, 1, prefix);
    const output = ts.transpileModule(lines[0], {
        compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext},
        reportDiagnostics: true,
    });
    assert.deepEqual(output.diagnostics, []);
    return new vm.Script(output.outputText);
}

function execute(script, key, values, canonicalize) {
    const array = [...values];
    if (!canonicalize) {
        const result = script.runInNewContext({[key]: array});
        assert.equal(result, array, key + ': sort must return the input array');
        return array;
    }

    let mapped;
    Object.defineProperty(array, 'map', {
        value(callback) {
            mapped = Array.prototype.map.call(this, callback);
            return mapped;
        },
    });
    const entries = {[key]: array};
    const result = script.runInNewContext({entries, getCanonicalFileName: canonicalize});
    assert(mapped, key + ': expected map to run');
    assert.equal(result, mapped, key + ': sort must return the mapped array');
    const field = 'sortedAndCanonicalized' + key[0].toUpperCase() + key.slice(1);
    assert.equal(entries[field], mapped, key + ': cache must store the mapped array');
    assert.deepEqual([...array], values, key + ': map must leave its input unchanged');
    return [...mapped];
}

function compareOutput() {
    let count = 0;
    for (const key of ['files', 'directories']) {
        const field = 'sortedAndCanonicalized' + key[0].toUpperCase() + key.slice(1);
        for (const [index, prefix, modes] of [
            [0, key + '.sort(', [['in place', undefined]]],
            [1, 'entries.' + field + ' = entries.' + key + '.map(', [
                ['identity', name => name], ['lowercase', name => name.toLowerCase()],
            ]],
        ]) {
            const original = compileStatement(input[index], prefix);
            const adapted = compileStatement(fs.readFileSync(files[index], 'utf8'), prefix);
            for (const [mode, canonicalize] of modes) {
                for (const [label, names] of samples) {
                    const before = execute(original, key, names, canonicalize);
                    const after = execute(adapted, key, names, canonicalize);
                    assert.deepEqual(after, before, 'order row: ' + key + ' / ' + mode + ' / ' + label);
                    count++;
                }
            }
        }
    }
    console.log('output: ' + count + ' order and array-identity rows passed');
}

try {
    const [sys, watch] = input;
    const fileSort = '                files.sort();';
    reject('scope', [substitute(sys, fileSort, [
        '                const sortedFileNames = (() => {',
        '                    const files: number[] = [10, 9, 1];',
        '                    files.sort();',
        '                    return files.map(String);',
        '                })();',
        '                files.push(...sortedFileNames);',
    ].join('\n')), watch]);
    reject('brand', [sys, substitute(watch,
        'type Canonicalized = string & { __canonicalized: void; };',
        'type Canonicalized = number & { __canonicalized: void; };')]);
    const callback = '    const getCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames) as ((name: string) => Canonicalized);';
    const numeric = '    const createGetCanonicalFileName = (_: boolean) => (name: string): string | number => name === "10" || name === "9" || name === "1" ? Number(name) : name;';
    reject('numeric producer', [sys, substitute(watch, callback, numeric + '\n' + callback)],
        'canonicalizer must be the imported core factory');
    const namespace = input[3];
    const newline = namespace.includes('\r\n') ? '\r\n' : '\n';
    assert(namespace.endsWith(newline), 'expected a final newline in _namespaces/ts.ts');
    reject('namespace factory', [sys, watch, input[2], namespace + 'export ' + numeric.trimStart() + newline],
        'canonicalizer must be the imported core factory');
    const inner = '    function getCachedFileSystemEntriesForBaseDir(path: Path) {';
    const next = '    function getBaseNameOfFileName(fileName: string) {';
    function scopedCallback(declarations) {
        return substitute(substitute(watch, inner, '    {\n' + declarations + '\n' + inner),
            next, '    }\n\n' + next);
    }
    reject('core factory return', [sys, watch, substitute(input[2],
        'export type GetCanonicalFileName = (fileName: string) => string;',
        'export type GetCanonicalFileName = (fileName: string) => string | number;')],
        'canonicalizer result must be a string callback');
    reject('watch entries shadow', [sys, substitute(substitute(watch,
        '        const entries = getCachedFileSystemEntries(getDirectoryPath(path));',
        '        {\n        const entries = getCachedFileSystemEntries(getDirectoryPath(path));'),
        '        return entries as SortedAndCanonicalizedMutableFileSystemEntries;',
        '        return entries as SortedAndCanonicalizedMutableFileSystemEntries;\n        }')],
        'entries must be declared directly in the owner');
    reject('watch callback shadow', [sys, scopedCallback(callback)],
        'callback must be declared directly in the host');
    reject('nested numeric producer', [sys, scopedCallback(numeric + '\n' + callback)],
        'callback must be declared directly in the host');
    reject('watch map member', [sys, substitute(watch,
        'interface MutableFileSystemEntries {\n    readonly files: string[];',
        'interface MutableFileSystemEntries {\n    readonly files: number[];')],
        'cache member must contain strings');
    reject('watch cache lookup', [sys, substitute(watch,
        '        const entries = getCachedFileSystemEntries(getDirectoryPath(path));',
        '        const getCachedFileSystemEntries = (_: Path) => ({ files: [], directories: [] });\n        const entries = getCachedFileSystemEntries(getDirectoryPath(path));')],
        'expected the host cache lookup');
    reject('binding', [substitute(sys, fileSort, [
        '                {',
        '                    const files: number[] = [10, 9, 1];',
        '                    files.sort();',
        '                }',
    ].join('\n')), watch]);
    reject('same-text shadow', [substitute(sys, fileSort, [
        '                {',
        '                    const files: string[] = [];',
        '                    files.sort();',
        '                }',
    ].join('\n')), watch]);
    reject('statement drift', [substitute(sys, fileSort, '                void files.sort();'), watch]);
    reject('duplicate', [substitute(sys, fileSort, fileSort + '\n' + fileSort), watch]);
    reject('parse error', [sys + '\nconst = ;\n', watch]);
    reject('second-file drift', [sys, substitute(watch,
        'entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>',
        '(entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>)')]);

    reset(input);
    accept();
    compareOutput();
    assert.deepEqual(outputBytes(), expected, 'clean output must match the adapter after snapshot');
    const sites = [
        ['files.sort();', 'directories.sort();'],
        [
            'entries.sortedAndCanonicalizedFiles = entries.files.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>;',
            'entries.sortedAndCanonicalizedDirectories = entries.directories.map(getCanonicalFileName).sort() as SortedArray<Canonicalized>;',
        ],
    ];
    const independent = sites.map((statements, index) => Buffer.from(statements.reduce(
        (text, statement) => substitute(text, statement, statement.replace('.sort()', '.sort(' + comparator + ')')),
        input[index])));
    assert.deepEqual(outputBytes(), independent, 'clean output must contain exactly four insertions');
    accept();
    assert.deepEqual(outputBytes(), expected, 'repeat output must match the adapter after snapshot');
    reset([substitute(sys, fileSort, '                files.sort(' + comparator + ');'), watch]);
    accept();
    assert.deepEqual(outputBytes(), expected, 'partial output must match the adapter after snapshot');
    console.log('clean, repeat and partial applies matched the after snapshot; clean output contains exactly four insertions');

} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}
