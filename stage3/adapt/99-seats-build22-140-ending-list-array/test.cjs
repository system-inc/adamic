'use strict';

const assert = require('node:assert/strict');
const { execFileSync, spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <tree stage3/apply.sh wrote>');
assert.equal(ts.version, '6.0.3');
const tree = process.argv[2];
const adapter = path.join(__dirname, 'adapt.cjs');
const ref = 'refs/stage3/99-seats-build22-140-ending-list-array';
const relative = 'src/compiler/moduleSpecifiers.ts';
const git = (...args) => execFileSync('git', ['-C', tree, ...args]);

assert.equal(git('diff', '--name-only', ref + '-before', ref + '-after').toString().trim(), relative);
const before = git('show', ref + '-before:' + relative);
const after = git('show', ref + '-after:' + relative);
const parse = bytes => ts.createSourceFile(relative, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
const original = parse(before);
const adapted = parse(after);

function one(items, label) {
    assert.equal(items.length, 1, label);
    return items[0];
}

// Match adapt.cjs's top-level function-with-body locator.
function owner(source, name) {
    return one(source.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === name && node.body), name);
}

const worker = owner(original, 'processEnding');
const factory = owner(original, 'getModuleSpecifierPreferences');
const parameter = one(worker.parameters.filter(node => node.name.getText(original) === 'allowedEndings'), 'allowedEndings');

function replace(node, text) {
    return original.text.slice(0, node.getStart(original)) + text + original.text.slice(node.end);
}

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-ending-list-'));
const target = path.join(scratch, relative);
const stamp = () => fs.statSync(target, { bigint: true }).mtimeNs;
function apply() {
    const result = spawnSync(process.execPath, [adapter, scratch], { encoding: 'utf8' });
    if (result.error) throw result.error;
    return result;
}

function refused(label, source, message, locatorCount) {
    fs.writeFileSync(target, source);
    const time = stamp();
    const result = apply();
    const headline = result.stderr.split(/\r?\n/).find(line => /^(AssertionError \[ERR_ASSERTION\]|Error): /.test(line));
    assert.equal(result.status, 1, label);
    assert.equal(headline, message, label);
    if (locatorCount) assert.ok(result.stderr.includes(locatorCount), label + ': locator count');
    assert.equal(fs.readFileSync(target).equals(Buffer.from(source)), true, label + ': source bytes');
    assert.equal(stamp(), time, label + ': write time');
    console.log('PASS ' + label + ' | ' + headline + (locatorCount ? ' | ' + locatorCount : ''));
}

try {
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, before);
    const first = apply();
    assert.equal(first.status, 0, first.stderr.trim().slice(-300));
    assert.equal(first.stdout.trim(), 'moduleSpecifiers.ts preference list: 7 edits');
    assert.equal(fs.readFileSync(target).equals(after), true, 'replay matches after bytes');
    console.log('PASS before replay | ' + first.stdout.trim());

    fs.writeFileSync(target, after);
    const time = stamp();
    const second = apply();
    assert.equal(second.status, 0, second.stderr.trim().slice(-300));
    assert.equal(second.stdout.trim(), 'moduleSpecifiers.ts preference list: 0 edits');
    assert.equal(fs.readFileSync(target).equals(after), true, 'replay keeps after bytes');
    assert.equal(stamp(), time, 'replay did not write');
    console.log('PASS after replay | ' + second.stdout.trim());

    const drift = 'AssertionError [ERR_ASSERTION]: moduleSpecifiers.ts preference-list snapshot drift';
    refused('one byte of parameter type', replace(parameter.type,
        parameter.type.getText(original).replace('ModuleSpecifierEnding', 'ModuleSpecifierEndinh')), drift);
    refused('only one edit applied', replace(parameter.type, 'readonly ModuleSpecifierEnding[]'), drift);
    refused('duplicate processEnding', original.text + '\n' + worker.getText(original) + '\n',
        'AssertionError [ERR_ASSERTION]: processEnding', '2 !== 1');
    const body = factory.getText(original);
    const changed = body.replace('syntaxImpliedNodeFormat !== impliedNodeFormat', 'syntaxImpliedNodeFormat === impliedNodeFormat');
    assert.notEqual(changed, body, 'factory comparison pin');
    refused('one byte in the factory body', replace(factory, changed),
        'Error: nonempty ending invariant drift: getModuleSpecifierPreferences');
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}

const endings = ts.ModuleSpecifierEnding;
const factoryResult = one(factory.body.statements.filter(ts.isReturnStatement), 'preference factory return').expression;
assert.ok(ts.isObjectLiteralExpression(factoryResult), 'preference factory result');
const listFactory = one(factoryResult.properties.filter(node =>
    ts.isPropertyAssignment(node) && node.name.getText(original) === 'getAllowedEndingsInPreferredOrder'), 'ending factory').initializer;
assert.ok(ts.isArrowFunction(listFactory) && ts.isBlock(listFactory.body), 'ending factory body');

const lists = [];
function collectResults(expression) {
    if (ts.isConditionalExpression(expression)) {
        collectResults(expression.whenTrue);
        collectResults(expression.whenFalse);
        return;
    }
    assert.ok(ts.isArrayLiteralExpression(expression) && expression.elements.length > 0, 'nonempty factory result');
    const names = expression.elements.map(element => {
        assert.ok(ts.isPropertyAccessExpression(element) && element.expression.getText(original) === 'ModuleSpecifierEnding',
            'ending enum member');
        assert.equal(typeof endings[element.name.text], 'number', element.name.text);
        return element.name.text;
    });
    const { line } = original.getLineAndCharacterOfPosition(expression.getStart(original));
    lists.push({ label: 'factory:' + (line + 1), names, values: names.map(name => endings[name]) });
}
function visit(node) {
    if (ts.isReturnStatement(node)) {
        assert.ok(node.expression, 'ending factory return expression');
        collectResults(node.expression);
        return;
    }
    ts.forEachChild(node, visit);
}
visit(listFactory.body);
assert.equal(lists.length, 11, 'factory result literals');
const singletonResults = { Minimal: 'pkg', Index: 'pkg/index', JsExtension: 'pkg/index.js', TsExtension: 'pkg/index.ts' };
const singletonNames = Object.keys(endings).filter(name => typeof endings[name] === 'number');
assert.equal(singletonNames.join(','), Object.keys(singletonResults).join(','), 'ending enum members');
for (const name of singletonNames) lists.push({ label: 'singleton:' + name, names: [name], values: [endings[name]] });

class InvalidEnding extends Error {}

function loadWorker(source) {
    const result = ts.transpileModule(owner(source, 'processEnding').getText(source) + '\nprocessEnding;', {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
        reportDiagnostics: true,
    });
    assert.equal(result.diagnostics.length, 0, 'worker transpile diagnostics');
    return vm.runInNewContext(result.outputText, {
        Extension: ts.Extension,
        ModuleSpecifierEnding: endings,
        fileExtensionIsOneOf: ts.fileExtensionIsOneOf,
        removeFileExtension: ts.removeFileExtension,
        removeSuffix: ts.removeSuffix,
        isDeclarationFileName: ts.isDeclarationFileName,
        tryGetRealFileNameForNonJsDeclarationFileName: ts.moduleSpecifiers.tryGetRealFileNameForNonJsDeclarationFileName,
        getJSExtensionForFile(file, options) {
            const extension = ts.moduleSpecifiers.tryGetJSExtensionForFile(file, options);
            assert.equal(typeof extension, 'string', file);
            return extension;
        },
        tryGetAnyFileFromPath(host, file) {
            return file === 'pkg' && host.indexFileExists ? 'pkg.ts' : undefined;
        },
        Debug: {
            assertNever(value) {
                try {
                    ts.Debug.assertNever(value);
                } catch (error) {
                    throw new InvalidEnding(error.message);
                }
            },
        },
    });
}

const originalWorker = loadWorker(original);
const adaptedWorker = loadWorker(adapted);
const files = [
    { file: 'pkg/index.ts' },
    { file: 'pkg/index.ts', found: true },
    { file: 'pkg/index.ts', found: false },
    { file: 'pkg/index.d.ts' },
    { file: 'pkg/data.json' },
    { file: 'pkg/main.mjs' },
    { file: 'pkg/main.cjs' },
    { file: 'pkg/main' },
    { file: 'pkg/main.mts' },
    { file: 'pkg/main.cts' },
    { file: 'pkg/types.d.mts' },
    { file: 'pkg/types.d.cts' },
    { file: 'pkg/data.d.json.ts' },
];

function outcome(run, values, { file, found }) {
    const host = found === undefined ? undefined : Object.freeze({ indexFileExists: found });
    try {
        const value = run(file, Object.freeze([...values]), Object.freeze({}), host);
        assert.equal(typeof value, 'string', file);
        return 'return ' + JSON.stringify(value);
    } catch (error) {
        if (!(error instanceof InvalidEnding)) throw error;
        return 'throw ' + JSON.stringify(error.message);
    }
}

function compare(list, row) {
    const label = list.label + ' [' + list.names.join(',') + '] ' + row.file +
        ' host=' + (row.found === undefined ? 'none' : row.found ? 'found' : 'missing');
    const expected = outcome(originalWorker, list.values, row);
    const actual = outcome(adaptedWorker, list.values, row);
    assert.equal(actual, expected, label);
    console.log('PASS ' + label + ' | before ' + expected + ' | after ' + actual);
    return expected;
}

for (const list of lists) {
    for (const row of files) {
        const expected = compare(list, row);
        if (list.label.startsWith('singleton:') && row.file === 'pkg/index.ts') {
            const result = list.names[0] === 'Minimal' && row.found ? 'pkg/index' : singletonResults[list.names[0]];
            assert.equal(expected, 'return ' + JSON.stringify(result), list.label + ': switch arm');
        }
    }
}
assert.equal(compare({ label: 'invalid', names: ['99'], values: [99] }, files[0]),
    'throw "Debug Failure. Illegal value: 99"', 'default switch arm');
console.log('PASS ' + (lists.length * files.length + 1) + ' side-by-side rows');
