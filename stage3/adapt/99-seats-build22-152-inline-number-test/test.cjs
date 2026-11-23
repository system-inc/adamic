'use strict';

const assert = require('node:assert/strict');
const {execFileSync, spawnSync} = require('node:child_process');
const {createHash} = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const api = path.resolve(__dirname, '../../api/node_modules');
const ts = require(path.join(api, 'typescript'));

const [tree, adapter = path.join(__dirname, 'adapt.cjs')] = process.argv.slice(2);
assert(tree, 'Usage: node test.cjs ADAPTED_TREE [ADAPTER]');
assert.equal(ts.version, '6.0.3');
const snapshot = 'refs/stage3/99-seats-build22-152-inline-number-test-before';
const compiler = 'src/compiler';
const oldLine = '    const oldEmitKind = oldOptionsOrEmitKind && (isNumber(oldOptionsOrEmitKind) ? oldOptionsOrEmitKind : getBuilderFileEmit(oldOptionsOrEmitKind));';
const currentLine = '    const emitKind = isNumber(optionsOrEmitKind) ? optionsOrEmitKind : getBuilderFileEmit(optionsOrEmitKind);';
const shadow = '    function isNumber(x: unknown): x is number { return typeof x === "number" && x !== 0; }';
const newline = '\r\n';

function substitute(text, before, after) {
    assert.equal(text.split(before).length, 2, 'Expected one match for ' + before);
    return text.replace(before, after);
}

function inline(line, name) {
    return substitute(line, 'isNumber(' + name + ')', 'typeof ' + name + ' === "number"');
}

function insertAfter(text, line, inserted) {
    return substitute(text, line + newline, line + newline + inserted + newline);
}

const nextOld = inline(oldLine, 'oldOptionsOrEmitKind');
const nextCurrent = inline(currentLine, 'optionsOrEmitKind');
const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-inline-number-'));
const input = path.join(temp, 'input');
let failures = 0;
let fixtureNumber = 0;
let output;

function file(root, name) {
    return path.join(root, compiler, name);
}

function receipt(root) {
    return Object.fromEntries(['builder.ts', 'core.ts'].map(name => [
        name, createHash('sha256').update(fs.readFileSync(file(root, name))).digest('hex'),
    ]));
}

function fixture(builder, core) {
    const root = path.join(temp, 'case-' + ++fixtureNumber);
    fs.cpSync(input, root, {recursive: true});
    fs.writeFileSync(file(root, 'builder.ts'), builder);
    fs.writeFileSync(file(root, 'core.ts'), core);
    return root;
}

function run(root) {
    const before = receipt(root);
    const child = spawnSync(process.execPath, [path.resolve(adapter), root], {
        encoding: 'utf8',
        env: {...process.env, NODE_PATH: api},
    });
    if (child.error) throw child.error;
    const after = receipt(root);
    console.log(JSON.stringify({exit: child.status, stdout: child.stdout.trim(),
        stderr: child.stderr.trim(), before, after}));
    return {child, before, after};
}

function check(name, body) {
    try {
        body();
        console.log('PASS ' + name);
    } catch (error) {
        failures++;
        console.error('FAIL ' + name + ': ' + error.message);
    }
}

function rejected(root) {
    const {child, before, after} = run(root);
    assert.equal(typeof child.status, 'number', 'Adapter failed to exit normally');
    assert.deepEqual({exit: child.status, files: after}, {exit: 1, files: before},
        'Expected exit 1 and unchanged builder.ts and core.ts');
}

function extract(text, fileName, name, isDeclaration) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const matches = source.statements.filter(node => isDeclaration(node) && node.name?.text === name);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source);
}

function evaluator(builder, core) {
    const source = [
        extract(builder, 'builder.ts', 'BuilderFileEmit', ts.isEnumDeclaration),
        extract(builder, 'builder.ts', 'getBuilderFileEmit', ts.isFunctionDeclaration),
        extract(builder, 'builder.ts', 'getPendingEmitKind', ts.isFunctionDeclaration),
    ];
    if (core !== undefined) source.push(extract(core, 'core.ts', 'isNumber', ts.isFunctionDeclaration));
    const result = ts.transpileModule(source.join('\n') + '\nexports.run = getPendingEmitKind;', {
        compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020},
        reportDiagnostics: true,
    });
    assert.deepEqual(result.diagnostics, []);
    let reads = 0;
    const exports = {};
    vm.runInNewContext(result.outputText, {
        exports,
        getEmitDeclarations(options) {
            reads++;
            return options.declaration;
        },
    });
    return (current, old) => {
        reads = 0;
        const value = exports.run(current, old);
        return {value, reads};
    };
}

try {
    fs.mkdirSync(input);
    const archive = execFileSync('git', ['-C', tree, 'archive', snapshot, 'src', 'package.json'],
        {maxBuffer: 100 * 1024 * 1024});
    execFileSync('tar', ['-xf', '-', '-C', input], {input: archive});
    const builder = fs.readFileSync(file(input, 'builder.ts'), 'utf8');
    const core = fs.readFileSync(file(input, 'core.ts'), 'utf8');
    assert(builder.includes(newline) && core.includes(newline), 'Expected CRLF source');
    const adapted = substitute(substitute(builder, oldLine, nextOld), currentLine, nextCurrent);

    check('clean two-edit application and byte-identical replay', () => {
        const root = fixture(builder, core);
        const first = run(root);
        output = fs.readFileSync(file(root, 'builder.ts'), 'utf8');
        const replay = run(root);
        assert.equal(first.child.status, 0, first.child.stderr);
        assert.equal(replay.child.status, 0, replay.child.stderr);
        assert.deepEqual(JSON.parse(first.child.stdout), {files: 1, edits: 2, javascriptChanging: true});
        assert.deepEqual(JSON.parse(replay.child.stdout), {files: 0, edits: 0, javascriptChanging: true});
        assert.equal(first.after['core.ts'], first.before['core.ts']);
        assert.deepEqual(replay.after, first.after);
        assert(output === adapted, 'Clean output differs from the two pinned replacements');
    });

    for (const [label, source, predicate] of [
        ['shadow on clean input', insertAfter(builder, currentLine, shadow), core],
        ['shadow on replay', insertAfter(adapted, nextCurrent, shadow), core],
        ['changed predicate', builder, substitute(core,
            '    return typeof x === "number";' + newline,
            '    return typeof x === "number" && x !== 0;' + newline)],
        ['third call', insertAfter(builder, currentLine, '    isNumber(optionsOrEmitKind);'), core],
        ['swapped statements', substitute(builder, oldLine + newline + currentLine,
            currentLine + newline + oldLine), core],
        ['half-adapted', substitute(builder, oldLine, nextOld), core],
    ]) {
        check(label, () => rejected(fixture(source, predicate)));
    }

    check('executing original and adapted emit-kind functions', () => {
        assert(output, 'No adapter output to evaluate');
        const original = evaluator(builder, core);
        const candidate = evaluator(output);
        const withDeclarations = {declaration: true, sourceMap: true};
        const withMap = {sourceMap: true};
        for (const [label, current, old, value, reads] of [
            ['2 and 1', 2, 1, 2, 0],
            ['current zero', 0, 1, 0, 0],
            ['old zero', 1, 0, 1, 0],
            ['old negative zero', 1, -0, 1, 0],
            ['old undefined', 1, undefined, 1, 0],
            ['old NaN', 1, NaN, 1, 0],
            ['current NaN', NaN, 1, NaN, 0],
            ['equal emit kinds', 1, 1, 0, 0],
            ['current options', withDeclarations, 1, 27, 1],
            ['old options', 2, withMap, 2, 1],
            ['both options', withDeclarations, withMap, 24, 2],
        ]) {
            check('VM row ' + label, () => {
                const base = original(current, old);
                const head = candidate(current, old);
                console.log(label + ': original ' + String(base.value) + ', ' + base.reads +
                    ' reads; output ' + String(head.value) + ', ' + head.reads + ' reads');
                assert(Object.is(base.value, value), 'Original value disagrees with the reviewed vector');
                assert.equal(base.reads, reads, 'Original option reads disagree with the reviewed vector');
                assert(Object.is(head.value, base.value), 'Output value disagrees with the original');
                assert.equal(head.reads, base.reads, 'Output option reads disagree with the original');
            });
        }
    });
} finally {
    fs.rmSync(temp, {recursive: true, force: true});
}
if (failures) process.exitCode = 1;
