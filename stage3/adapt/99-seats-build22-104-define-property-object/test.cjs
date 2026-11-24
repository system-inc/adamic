'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const { createRequire } = require('node:module');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <tree produced by stage3/apply.sh>');
assert.equal(ts.version, '6.0.3');
const tree = path.resolve(process.argv[2]);
const snapshot = 'refs/stage3/' + path.basename(__dirname);
const adapter = path.join(__dirname, 'adapt.cjs');
const files = ['src/compiler/types.ts', 'src/compiler/utilities.ts', 'src/compiler/binder.ts', 'src/compiler/checker.ts'];

function show(side, file) {
    return execFileSync('git', ['-C', tree, 'show', `${snapshot}-${side}:${file}`]);
}

function parse(file, bytes) {
    const source = ts.createSourceFile(file, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `Cannot parse ${file}`);
    return source;
}

function find(root, predicate) {
    const found = [];
    (function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}

function one(nodes, label) {
    assert.equal(nodes.length, 1, `expected one ${label}, found ${nodes.length}`);
    return nodes[0];
}

function functionNamed(source, name) {
    return one(find(source, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body !== undefined), name);
}

// The arrow makeSerializePropertySymbol hands forEach to find a property's setter among its declarations, in the
// TypeScript of a snapshot and in the stock bundle's JavaScript alike.
function setterFinder(source) {
    const serializer = functionNamed(source, 'makeSerializePropertySymbol');
    const call = one(find(serializer, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) &&
        node.expression.text === 'forEach' && node.arguments.length === 2 &&
        node.arguments[0].getText(source) === 'p.declarations' && ts.isArrowFunction(node.arguments[1])), 'setter finder');
    return call.arguments[1];
}

const before = Object.fromEntries(files.map(file => [file, show('before', file)]));
const after = Object.fromEntries(files.map(file => [file, show('after', file)]));

function testReplay() {
    const changed = execFileSync('git', ['-C', tree, 'diff', '--name-only', `${snapshot}-before`, `${snapshot}-after`],
        { encoding: 'utf8' }).trim().split('\n');
    assert.deepEqual(changed.sort(), [...files].sort(), 'the adapter changes exactly the four files');
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'define-property-object-test-'));
    try {
        for (const file of files) {
            fs.mkdirSync(path.dirname(path.join(scratch, file)), { recursive: true });
            fs.writeFileSync(path.join(scratch, file), before[file]);
        }
        const run = () => execFileSync(process.execPath, [adapter, scratch], { encoding: 'utf8' }).trim();
        assert.equal(run(), JSON.stringify({ files: 4, edits: 5, javascriptChanging: true }));
        for (const file of files) {
            assert(fs.readFileSync(path.join(scratch, file)).equals(after[file]), `${file} differs from the adapter's -after snapshot`);
        }
        assert.equal(run(), JSON.stringify({ files: 0, edits: 0 }));
        for (const file of files) {
            assert(fs.readFileSync(path.join(scratch, file)).equals(after[file]), `${file} changed on a second application`);
        }
        console.log(`  the adapter on -before writes -after byte for byte in ${files.join(', ')}, then changes nothing`);
    }
    finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

// Each file's -after is its -before with only the retype and the four checks, written out here apart from the adapter.
function testOnlyTheFiveSites() {
    const edits = {
        'src/compiler/types.ts': [['StringLiteralLike | NumericLiteral, Expression]', 'StringLiteralLike | NumericLiteral, ObjectLiteralExpression]', 1]],
        'src/compiler/utilities.ts': [[
            'isObjectLiteralExpression(node.arguments[2]) ? hasExpandoValueProperty(node.arguments[2], node.arguments[1].text === "prototype") : undefined',
            'hasExpandoValueProperty(node.arguments[2], node.arguments[1].text === "prototype")', 1]],
        'src/compiler/binder.ts': [[
            'isObjectLiteralExpression(declaration.arguments[2]) && some(declaration.arguments[2].properties, p => {',
            'some(declaration.arguments[2].properties, p => {', 2]],
        'src/compiler/checker.ts': [
            ['return isObjectLiteralExpression(d.arguments[2]) ? forEach(d.arguments[2].properties, propDecl => {', 'return forEach(d.arguments[2].properties, propDecl => {', 1],
            ['                                            return propDecl;\n                                        }\n                                    }) : undefined;',
                '                                            return propDecl;\n                                        }\n                                    });', 1],
        ],
    };
    for (const file of files) {
        let text = after[file].toString('utf8');
        const newline = text.includes('\r\n') ? '\r\n' : '\n';
        for (const [adapted, stock, count] of edits[file]) {
            assert.equal(text.split(adapted.replaceAll('\n', newline)).length - 1, count, `${file}: ${adapted}`);
            text = text.replaceAll(adapted.replaceAll('\n', newline), stock.replaceAll('\n', newline));
        }
        assert.equal(text, before[file].toString('utf8'), `${file} differs from -before outside the define-property sites`);
    }
}

function testUntouched() {
    const utilities = [parse('before.ts', before['src/compiler/utilities.ts']), parse('after.ts', after['src/compiler/utilities.ts'])];
    const [predicateBefore, predicateAfter] = utilities.map(source => functionNamed(source, 'isBindableObjectDefinePropertyCall'));
    assert.equal(predicateAfter.getText(utilities[1]), predicateBefore.getText(utilities[0]), 'the predicate changed');
    assert.equal(find(predicateAfter, node => ts.isIdentifier(node) && node.text === 'isObjectLiteralExpression').length, 0,
        'isObjectLiteralExpression sits in the predicate');
    const checkers = [parse('before.ts', before['src/compiler/checker.ts']), parse('after.ts', after['src/compiler/checker.ts'])];
    for (const name of ['getInitializerTypeFromAssignmentDeclaration', 'isReadonlyAssignmentDeclaration']) {
        const [stock, adapted] = checkers.map(source => functionNamed(source, name));
        assert.equal(adapted.getText(checkers[1]), stock.getText(checkers[0]), `${name} changed`);
        assert.equal(find(adapted, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) &&
            node.expression.text === 'checkExpressionCached' && /\.arguments\[2\]$/.test(node.arguments[0].getText(checkers[1]))).length, 1, name);
    }
}

// The stock bundle with the units named replaced by the same units of a snapshot, compiled to JavaScript. Only the
// smallest function holding each read is replaced, so the rest of the bundle stays stock tsc 6.0.3: the snapshots
// carry other adaptations elsewhere in these files, and the setter finder's enclosing makeSerializePropertySymbol
// calls a helper an earlier adaptation added, which the stock bundle doesn't hold.
const bundlePath = require.resolve('typescript');
const bundleText = fs.readFileSync(bundlePath, 'utf8');
const bundle = parse(bundlePath, bundleText);
const units = [
    ['utilities.ts', 'getAssignedExpandoInitializer', source => functionNamed(source, 'getAssignedExpandoInitializer')],
    ['binder.ts', 'bindPotentiallyNewExpandoMemberToNamespace', source => functionNamed(source, 'bindPotentiallyNewExpandoMemberToNamespace')],
    ['checker.ts', 'the setter finder in makeSerializePropertySymbol', setterFinder],
];

function javascriptOf(source, node) {
    // An exported function loses its export: the bundle declares it in its own scope.
    const start = node.modifiers?.length ? ts.skipTrivia(source.text, node.modifiers.end) : node.getStart(source);
    const text = source.text.slice(start, node.end);
    const wrapped = ts.isArrowFunction(node) ? `(${text});` : text;
    const output = ts.transpileModule(wrapped, { compilerOptions: { target: ts.ScriptTarget.ES2020 }, reportDiagnostics: true });
    assert.deepEqual(output.diagnostics, []);
    const emitted = ts.createSourceFile('unit.js', output.outputText, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    const statement = one(emitted.statements.filter(node => !ts.isExpressionStatement(node) || !ts.isStringLiteral(node.expression)), 'emitted unit');
    const unit = ts.isArrowFunction(node) ? statement.expression.expression : statement;
    assert(ts.isArrowFunction(node) ? ts.isArrowFunction(unit) : ts.isFunctionDeclaration(unit));
    return { source: emitted, node: unit };
}

// The four checks counted: each isObjectLiteralExpression call on a third argument reports what it answered.
function instrumented(unit, first) {
    const checks = find(unit.node, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) &&
        node.expression.text === 'isObjectLiteralExpression' && /\.arguments\[2\]$/.test(node.arguments[0].getText(unit.source)));
    let text = unit.node.getText(unit.source);
    const offset = unit.node.getStart(unit.source);
    checks.forEach((check, index) => assert(index === 0 || check.pos > checks[index - 1].end));
    for (const [index, check] of [...checks.entries()].reverse()) {
        const start = check.getStart(unit.source) - offset, end = check.end - offset;
        text = text.slice(0, start) + `__reach(${first + index}, ${text.slice(start, end)})` + text.slice(end);
    }
    return { text, checks: checks.length };
}

function spliced(side, sources, instrument) {
    const replacements = [];
    let reads = 1;
    for (const [file, label, locate] of units) {
        const stock = locate(bundle);
        const unit = javascriptOf(sources[file], locate(sources[file]));
        let text = unit.node.getText(unit.source);
        if (instrument) {
            const counted = instrumented(unit, reads);
            assert.equal(counted.checks, file === 'binder.ts' ? 2 : 1, `${label}'s checks`);
            reads += counted.checks;
            text = counted.text;
        }
        replacements.push([stock.getStart(bundle), stock.end, text, `${side} ${label}`]);
    }
    let text = bundleText;
    for (const [start, end, replacement] of replacements.sort((a, b) => b[0] - a[0])) {
        text = text.slice(0, start) + replacement + text.slice(end);
    }
    return text;
}

const reached = {};
function __reach(read, isLiteral) {
    reached[read] ||= { identifier: 0, literal: 0 };
    reached[read][isLiteral ? 'literal' : 'identifier']++;
    return isLiteral;
}

function load(text) {
    const module = { exports: {} };
    vm.compileFunction(text, ['exports', 'require', 'module', '__filename', '__dirname', '__reach'], { filename: bundlePath })(
        module.exports, createRequire(bundlePath), module, bundlePath, path.dirname(bundlePath), __reach);
    assert.equal(module.exports.version, '6.0.3');
    return module.exports;
}

function flagNames(api, flags) {
    return Object.entries(api.SymbolFlags)
        .filter(([, value]) => typeof value === 'number' && value !== 0 && (value & (value - 1)) === 0 && (flags & value))
        .map(([name]) => name).join('|');
}

// One program under allowJs, checkJs and declaration: every diagnostic, every emitted file, and each define-property
// call's property symbol, as one record whose JSON two compilers must share. A throw is part of the record.
function observe(api, files, expand) {
    const options = { allowJs: true, checkJs: true, declaration: true, target: api.ScriptTarget.ES2020, lib: ['lib.es2020.d.ts'], types: [], outDir: '/out' };
    const host = api.createCompilerHost(options, true);
    const getSourceFile = host.getSourceFile.bind(host), fileExists = host.fileExists.bind(host), readFile = host.readFile.bind(host);
    host.getSourceFile = (name, language, onError, create) =>
        name in files ? api.createSourceFile(name, files[name], language) : getSourceFile(name, language, onError, create);
    host.fileExists = name => name in files || fileExists(name);
    host.readFile = name => name in files ? files[name] : readFile(name);
    const emitted = {};
    host.writeFile = (name, text) => { emitted[name] = text; };
    try {
        const program = api.createProgram(Object.keys(files), options, host);
        const emit = program.emit();
        const diagnostics = [...api.getPreEmitDiagnostics(program), ...emit.diagnostics].map(diagnostic => ({
            file: diagnostic.file?.fileName, start: diagnostic.start, length: diagnostic.length, code: diagnostic.code,
            category: diagnostic.category, message: api.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
        }));
        const checker = program.getTypeChecker();
        const symbols = [];
        const printer = api.createPrinter();
        const expanded = [];
        for (const name of Object.keys(files)) {
            const source = program.getSourceFile(name);
            (function visit(node) {
                if (api.isCallExpression(node) && node.expression.getText(source) === 'Object.defineProperty' && node.arguments.length === 3) {
                    const symbol = checker.getSymbolAtLocation(node.arguments[1]);
                    symbols.push({
                        at: `${name}:${source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1}`,
                        flags: symbol && symbol.flags, names: symbol && flagNames(api, symbol.flags),
                        declarations: symbol && symbol.declarations.map(declaration => `${declaration.getSourceFile().fileName}:${declaration.pos}`),
                        type: symbol && checker.typeToString(checker.getTypeOfSymbol(symbol)),
                    });
                }
                if (expand && api.isIdentifier(node) && node.text === expand && api.isFunctionDeclaration(node.parent)) {
                    // The symbol's expanded declaration, through the emit resolver's symbolToDeclarations as a hover
                    // expands a class: for a global merged across script files it serializes the merged members.
                    const symbol = checker.getSymbolAtLocation(node);
                    const declarations = checker.getEmitResolver(undefined, undefined, true).symbolToDeclarations(symbol,
                        api.SymbolFlags.Value | api.SymbolFlags.Type,
                        api.NodeBuilderFlags.MultilineObjectLiterals | api.NodeBuilderFlags.UseAliasDefinedOutsideCurrentScope, undefined, 0, {});
                    expanded.push(declarations.map(declaration => printer.printNode(api.EmitHint.Unspecified, declaration, source)).join('\n'));
                }
                api.forEachChild(node, visit);
            })(source);
        }
        return { diagnostics, emitted, symbols, expanded };
    }
    catch (error) {
        return { threw: String(error && error.message) };
    }
}

const cases = {
    identifier: {
        files: {
            '/src/object.js': 'const o = {};\nconst d = { value: 1, writable: false };\nObject.defineProperty(o, "p", d);\no.p = 2;\nconst n = o.p;\n',
            '/src/function.js': 'function f() {}\nconst g = { get() { return 1; } };\nObject.defineProperty(f, "q", g);\nconst m = f.q;\n',
            '/src/prototype.js': 'function C() { this.a = 1; }\nconst v = { value: 1 };\nObject.defineProperty(C.prototype, "r", v);\nconst k = new C().r;\n',
            '/src/a.js': 'function F() { this.a = 1; }\nvar s = { value: 1 };\nObject.defineProperty(F, "s", s);\nObject.defineProperty(F.prototype, "p", s);\n',
            '/src/b.js': 'var F = {};\nObject.defineProperty(F.prototype, "p", { set(v) {} });\n',
        },
        expand: 'F',
    },
    literal: {
        files: {
            '/src/object.js': 'const o = {};\nObject.defineProperty(o, "p", { value: 1, writable: false });\no.p = 2;\nconst n = o.p;\n',
            '/src/function.js': 'function f() {}\nObject.defineProperty(f, "q", { get() { return 1; } });\nconst m = f.q;\n',
            '/src/prototype.js': 'function C() { this.a = 1; }\nObject.defineProperty(C.prototype, "r", { value: 1 });\nObject.defineProperty(C.prototype, "w", { set(v) {} });\nObject.defineProperty(C.prototype, "m", { value: function () {} });\nconst k = new C().r;\n',
        },
        expand: 'C',
    },
};

// The parts of a record that differ, so a red line says which observation changed.
function differing(expected, actual) {
    if (expected.threw || actual.threw) return `threw: ${expected.threw ?? 'nothing'} versus ${actual.threw ?? 'nothing'}`;
    return ['diagnostics', 'emitted', 'symbols', 'expanded'].filter(key => JSON.stringify(expected[key]) !== JSON.stringify(actual[key])).join(', ');
}

function testStockBeside(label, files, expand, bundles) {
    const expected = observe(require('typescript'), files, expand);
    assert(!expected.threw, `stock tsc threw on the ${label} case: ${expected.threw}`);
    console.log(`  stock on the ${label} case: ${expected.diagnostics.length} diagnostics ` +
        `(${expected.diagnostics.map(diagnostic => 'TS' + diagnostic.code).join(', ') || 'none'}), ` +
        `${Object.keys(expected.emitted).length} emitted files, ${expected.symbols.length} define-property symbols`);
    for (const symbol of expected.symbols) console.log(`  symbol ${symbol.at} ${symbol.names} ${symbol.type} declarations ${symbol.declarations.join(' ')}`);
    for (const [name, text] of Object.entries(expected.emitted).filter(([name]) => name.endsWith('.d.ts'))) {
        console.log(`  ${name}: ${text.trim().replaceAll('\n', ' ')}`);
    }
    for (const text of expected.expanded) console.log(`  expanded ${expand}: ${text.replaceAll('\n', ' ')}`);
    const control = observe(bundles.control, files, expand);
    assert.equal(differing(expected, control), '', `${label}: the -before functions spliced into the stock bundle differ from stock`);
    const adapted = observe(bundles.adapted, files, expand);
    assert.equal(differing(expected, adapted), '', `${label}: the adapted functions differ from stock`);
    for (const key of Object.keys(reached)) delete reached[key];
    const counted = observe(bundles.counted, files, expand);
    assert.equal(differing(expected, counted), '', `${label}: the counted adapted functions differ from stock`);
    const reads = [1, 2, 3, 4].map(read => reached[read] || { identifier: 0, literal: 0 });
    console.log(`  reached ${reads.map((read, index) => `read ${index + 1}: ${read.identifier} non-literal, ${read.literal} literal`).join('; ')}`);
    for (const [index, read] of reads.entries()) {
        assert(label === 'identifier' ? read.identifier > 0 : read.literal > 0, `the ${label} case never reaches read ${index + 1}`);
    }
}

// Every check runs, and each says PASS or FAIL, so one wrong version shows each check it turns red.
let failed = 0;
function check(name, body) {
    try {
        body();
        console.log(`PASS ${name}`);
    }
    catch (error) {
        failed++;
        console.log(`FAIL ${name}: ${String(error && error.message).split('\n')[0]}`);
    }
}

check('replay', testReplay);
check('only the retype and the four checks differ', testOnlyTheFiveSites);
check('the predicate and the checkExpressionCached reads are untouched', testUntouched);
const snapshotUnits = side => Object.fromEntries(['utilities.ts', 'binder.ts', 'checker.ts']
    .map(file => [file, parse(file, (side === 'before' ? before : after)['src/compiler/' + file])]));
const bundles = {};
check('the stock bundle with -before and -after functions spliced in loads', () => {
    bundles.control = load(spliced('before', snapshotUnits('before'), false));
    bundles.adapted = load(spliced('after', snapshotUnits('after'), false));
    bundles.counted = load(spliced('after', snapshotUnits('after'), true));
});
for (const [label, { files, expand }] of Object.entries(cases)) {
    check(`${label} third argument: stock and adapted identical`, () => testStockBeside(label, files, expand, bundles));
}
if (failed) {
    console.log(`FAIL 99-seats-build22-104-define-property-object: ${failed} checks failed`);
    process.exitCode = 1;
}
else {
    console.log('PASS 99-seats-build22-104-define-property-object');
}
