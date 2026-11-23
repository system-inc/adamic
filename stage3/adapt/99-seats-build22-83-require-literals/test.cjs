'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');
const ts = require('typescript');

const relative = 'src/compiler/utilities.ts';
const snapshot = 'refs/stage3/99-seats-build22-83-require-literals-before';
const adapter = path.join(__dirname, 'adapt.cjs');
const collectorName = 'getRequiresAtTopOfFile';

function namedFunction(statements, name) {
    const matches = statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(matches.length, 1, name);
    assert(matches[0].body);
    return matches[0];
}

function parseUtilities(text) {
    const file = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0);
    return file;
}

function requireBranch(file) {
    const preference = namedFunction(file.statements, 'getModuleSpecifierEndingPreference');
    const infer = namedFunction(preference.body.statements, 'inferPreference');
    const declarations = infer.body.statements.filter(ts.isVariableStatement)
        .flatMap(node => [...node.declarationList.declarations]);
    const specifiers = declarations.filter(node => ts.isIdentifier(node.name) && node.name.text === 'specifiers');
    assert.equal(specifiers.length, 1);
    const initializer = specifiers[0].initializer;
    assert(initializer && ts.isConditionalExpression(initializer));
    assert(ts.isConditionalExpression(initializer.whenFalse));
    return initializer.whenFalse.whenTrue.getText(file);
}

function evaluateCollector(file, bindings) {
    const collector = namedFunction(file.statements, collectorName);
    const emitted = ts.transpileModule(collector.getText(file), {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
        reportDiagnostics: true,
    });
    assert.deepEqual(emitted.diagnostics, []);
    const result = vm.runInNewContext(emitted.outputText + '\n' + collectorName, { ...bindings });
    assert.equal(typeof result, 'function');
    return result;
}

function runAdapter(tree, expected) {
    const stdout = execFileSync(process.execPath, [adapter, tree], { encoding: 'utf8' }).trim();
    console.log(stdout);
    assert.equal(stdout, JSON.stringify(expected));
}

function sha256(file) {
    return createHash('sha256').update(fs.readFileSync(file)).digest('hex');
}

function loadCollectors(adaptedTree) {
    const originalBytes = execFileSync('git', ['-C', adaptedTree, 'show', snapshot + ':' + relative]);
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'require-literals-test-'));
    try {
        const file = path.join(scratch, relative);
        fs.mkdirSync(path.dirname(file), { recursive: true });
        fs.writeFileSync(file, originalBytes);
        runAdapter(scratch, { files: 1, edits: 3, javascriptChanging: true });
        const digest = sha256(file);
        runAdapter(scratch, { files: 0, edits: 0 });
        assert.equal(sha256(file), digest);
        console.log('Replay SHA-256: ' + digest);

        const original = parseUtilities(originalBytes.toString('utf8'));
        const adapted = parseUtilities(fs.readFileSync(file, 'utf8'));
        assert.equal(requireBranch(original), collectorName + '(sourceFile).map(r => r.arguments[0])');
        assert.equal(requireBranch(adapted), collectorName + '(sourceFile)');

        const bindings = {};
        for (const name of ['isRequireVariableStatement', 'isExpressionStatement', 'isRequireCall', 'concatenate', 'append']) {
            assert.equal(typeof ts[name], 'function', name);
            bindings[name] = ts[name];
        }
        assert(Array.isArray(ts.emptyArray));
        bindings.emptyArray = ts.emptyArray;
        return [evaluateCollector(original, bindings), evaluateCollector(adapted, bindings)];
    }
    finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

assert.equal(process.argv.length, 3, 'usage: node test.cjs <tree produced by stage3/apply.sh>');
assert.equal(ts.version, '6.0.3');
const [original, changed] = loadCollectors(path.resolve(process.argv[2]));

function source(text) {
    return ts.createSourceFile('review.js', text, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
}

function verify(input, expected) {
    const before = Array.from(original(input), node => node.arguments[0]);
    assert.deepEqual(before.map(node => node.text), expected);
    const after = changed(input);
    assert.deepEqual(Array.from(after, node => node.text), expected);
    after.forEach((node, i) => assert.strictEqual(node, before[i]));
}

test('Keeps declaration and statement order and original literal identity', () => {
    verify(source('const a = require("a"), b = require("b"); require(`c`); const x = require("ignored"), y = 1; require("d");'),
        ['a','b','c','d']);
});
test('StopsAfterFourNonRequireStatements', () => {
    verify(source('0; require("a"); 1; 2; require("b"); 3; require("too-late");'), ['a','b']);
});
test('Handles empty input and rejects malformed require shapes', () => {
    verify(source(''), []);
    verify(source('require(); require(a); require("x", "y"); other("z");'), []);
});
