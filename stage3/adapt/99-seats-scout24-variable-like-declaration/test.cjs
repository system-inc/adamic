'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');

const globals = 'interface Array<T> {} interface Boolean {} interface Function {} interface CallableFunction {} interface NewableFunction {} interface IArguments {} interface Number {} interface Object {} interface RegExp {} interface String {}';

const cases = [
    {
        name: 'variable-like reporting',
        files: { 'ambient.d.ts': 'declare class C { private p; q; }', 'main.ts': 'function f(x) { return x; }' },
        names: { 'ambient.d.ts': ['p', 'q'], 'main.ts': ['x'] },
    },
    {
        name: 'JSDoc property reporting',
        files: { 'main.js': '/**\n * @typedef {object} Shape\n * @property prop\n */\n/** @type {Shape} */\nconst value = { prop: 1 };\nconst read = value.prop;' },
        names: { 'main.js': ['read'] },
        js: true,
    },
    {
        name: 'JSDoc callback parameter reporting',
        files: { 'main.js': '/**\n * @callback Callback\n * @param arg\n */\n/** @type {Callback} */\nconst f = arg => arg;\nconst read = f(1);' },
        names: { 'main.js': ['arg', 'read'] },
        js: true,
    },
    {
        name: 'export default widening',
        files: { 'value.ts': 'export default { count: 1 };', 'main.ts': 'import value from "./value.js"; const read = value;' },
        names: { 'main.ts': ['value', 'read'] },
    },
    {
        name: 'export equals widening',
        files: { 'value.ts': 'export = { count: 1 };', 'main.ts': 'import value = require("./value"); const read = value;' },
        names: { 'main.ts': ['value', 'read'] },
        commonjs: true,
    },
];

function answers(compiler, root, fixture) {
    const sources = new Map(Object.entries({ 'globals.d.ts': globals, ...fixture.files }));
    const options = {
        target: compiler.ScriptTarget.ES2024,
        module: fixture.commonjs ? compiler.ModuleKind.CommonJS : compiler.ModuleKind.ESNext,
        moduleResolution: fixture.commonjs ? compiler.ModuleResolutionKind.Node10 : compiler.ModuleResolutionKind.Bundler,
        ignoreDeprecations: '6.0',
        allowJs: !!fixture.js, checkJs: !!fixture.js, noImplicitAny: true, noLib: true, noEmit: true, types: [],
    };
    const readFile = filename => sources.get(path.relative(root, filename));
    const host = {
        getSourceFile(filename, target) {
            const text = readFile(filename);
            return text === undefined ? undefined : compiler.createSourceFile(filename, text, target, true);
        },
        getCurrentDirectory: () => root, getCanonicalFileName: filename => filename,
        useCaseSensitiveFileNames: () => true, getNewLine: () => '\n',
        getDefaultLibFileName: () => path.join(root, 'globals.d.ts'),
        fileExists: filename => readFile(filename) !== undefined, readFile,
        directoryExists: filename => filename === root,
        writeFile() { throw new Error('unexpected emit from witness'); },
    };
    const program = compiler.createProgram([...sources.keys()].map(file => path.join(root, file)), options, host);
    const checker = program.getTypeChecker();
    const diagnostics = compiler.getPreEmitDiagnostics(program).map(d => ({
        code: d.code, file: d.file ? path.relative(root, d.file.fileName) : null,
        start: d.start ?? null, length: d.length ?? null,
        message: compiler.flattenDiagnosticMessageText(d.messageText, '\n'),
    }));
    const types = {};
    for (const [filename, names] of Object.entries(fixture.names)) {
        const source = program.getSourceFile(path.join(root, filename));
        function visit(node) {
            const name = node.name;
            if (name && compiler.isIdentifier(name) && names.includes(compiler.idText(name))) {
                const symbol = checker.getSymbolAtLocation(name);
                assert(symbol, compiler.idText(name));
                const key = filename + ':' + compiler.idText(name);
                assert(!(key in types), key);
                types[key] = checker.typeToString(checker.getTypeOfSymbolAtLocation(symbol, name));
            }
            compiler.forEachChild(node, visit);
        }
        visit(source);
        for (const name of names) assert(filename + ':' + name in types, name);
    }
    return { diagnostics, types };
}

const fs = require('node:fs');
const os = require('node:os');
const { execFileSync } = require('node:child_process');
const { pathToFileURL } = require('node:url');

const snapshot = 'refs/stage3/99-seats-scout24-variable-like-declaration';
const checkerSource = 'src/compiler/checker.ts';
const checkerJavaScript = 'src/compiler/checker.js';
const loaderOptions = {
    strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
    erasableSyntaxOnly: false, verbatimModuleSyntax: true,
    allowImportingTsExtensions: true, noEmit: true, module: 'ESNext',
    moduleDetection: 'force', moduleResolution: 'bundler', target: 'ES2024',
    lib: ['es2024'], types: [],
};

function fromSnapshot(applied, side, directory) {
    fs.mkdirSync(directory);
    const archive = directory + '.tar';
    execFileSync('git', ['-C', applied, 'archive', '--output', archive, snapshot + '-' + side, 'src']);
    execFileSync('tar', ['-xf', archive, '-C', directory]);
    fs.unlinkSync(archive);
}

function readChecker(directory) {
    return fs.readFileSync(path.join(directory, checkerSource));
}

function apply(directory) {
    const report = execFileSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), directory], { encoding: 'utf8' });
    return JSON.parse(report);
}

function compilerProgram(ts, root, options) {
    const converted = ts.convertCompilerOptionsFromJson(options, root);
    assert.deepEqual(converted.errors, []);
    return ts.createProgram([path.join(root, 'src/tsc/tsc.ts')], converted.options);
}

function checkTypes(ts, root) {
    const typeRoots = [path.resolve(__dirname, '../../api/node_modules/@types')];
    const program = compilerProgram(ts, root, { ...loaderOptions, types: ['node'], typeRoots });
    const diagnostics = ts.getPreEmitDiagnostics(program);
    const formatted = ts.formatDiagnostics(diagnostics, {
        getCurrentDirectory: () => root,
        getCanonicalFileName: filename => filename,
        getNewLine: () => '\n',
    });
    assert.equal(diagnostics.length, 0, formatted);
    console.log('PASS snapshot type-check: 0 diagnostics');
}

function emit(ts, root, output) {
    const program = compilerProgram(ts, root, {
        ...loaderOptions, noEmit: false, noCheck: true, rewriteRelativeImportExtensions: true,
        sourceMap: true, rootDir: root, outDir: output,
    });
    const result = program.emit();
    assert.equal(result.emitSkipped, false);
    assert.deepEqual(result.diagnostics, []);
    fs.writeFileSync(path.join(output, 'package.json'), '{"type":"module"}\n');
}

async function load(output) {
    const filename = path.join(output, 'src/compiler/sys.js');
    const globals = { __filename: filename, __dirname: path.dirname(filename) };
    const previous = Object.fromEntries(Object.keys(globals).map(key => [key, Object.getOwnPropertyDescriptor(globalThis, key)]));
    try {
        Object.assign(globalThis, globals);
        return await import(pathToFileURL(path.join(output, 'src/compiler/_namespaces/ts.js')).href);
    } finally {
        for (const [key, descriptor] of Object.entries(previous)) {
            if (descriptor) Object.defineProperty(globalThis, key, descriptor);
            else delete globalThis[key];
        }
    }
}

function checkJavaScript(before, after) {
    function files(directory) {
        return fs.readdirSync(directory, { recursive: true }).filter(file => file.endsWith('.js')).sort();
    }
    const names = files(before);
    assert.deepEqual(files(after), names);
    const read = (directory, file) => fs.readFileSync(path.join(directory, file));
    const changed = names.filter(file => !read(before, file).equals(read(after, file)));
    assert.deepEqual(changed, [checkerJavaScript]);
    const stock = read(before, checkerJavaScript).toString('utf8');
    const original = 'if (!declarationBelongsToPrivateAmbientMember(declaration)) {';
    const guarded = 'if (!(isVariableLike(declaration) && declarationBelongsToPrivateAmbientMember(declaration))) {';
    assert.equal(stock.split(original).length, 2);
    assert.equal(read(after, checkerJavaScript).toString('utf8'), stock.replace(original, guarded));
    console.log('PASS emitted JavaScript: only the checker guard line changed');
}

async function run(applied) {
    const ts = require('typescript');
    assert.equal(ts.version, '6.0.3');
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'family3-output-'));
    try {
        const before = path.join(scratch, 'before');
        const after = path.join(scratch, 'after');
        const replay = path.join(scratch, 'replay');
        fromSnapshot(applied, 'before', before);
        fromSnapshot(applied, 'after', after);
        fs.mkdirSync(path.join(replay, 'src/compiler'), { recursive: true });
        fs.writeFileSync(path.join(replay, checkerSource), readChecker(before));
        assert(apply(replay).edits > 0, "adapter replay must edit the before snapshot");
        assert.deepEqual(readChecker(replay), readChecker(after));
        console.log('PASS adapter replay: before became the exact after bytes');
        assert.deepEqual(apply(replay), { edits: 0 });
        assert.deepEqual(readChecker(replay), readChecker(after));
        console.log('PASS adapter idempotence: 0 edits');
        checkTypes(ts, after);

        const stockEmit = path.join(scratch, 'stock-emit');
        const adaptedEmit = path.join(scratch, 'adapted-emit');
        emit(ts, before, stockEmit);
        emit(ts, after, adaptedEmit);
        const stock = await load(stockEmit);
        const adapted = await load(adaptedEmit);
        for (const fixture of cases) {
            const fixtureRoot = path.join(scratch, 'inputs');
            const expected = answers(stock, fixtureRoot, fixture);
            const actual = answers(adapted, fixtureRoot, fixture);
            assert.deepEqual(actual, expected, JSON.stringify({ fixture, stock: expected, adapted: actual }, null, 2));
            const codes = expected.diagnostics.map(d => d.code);
            const expectedCodes = fixture.js ? [7005] : fixture.name === 'variable-like reporting' ? [7008, 7006] : [];
            assert.deepEqual(codes, expectedCodes, fixture.name);
            const type = fixture.name.startsWith('export ') ? '{ count: number; }' : 'any';
            assert(Object.values(expected.types).every(value => value === type), JSON.stringify(expected.types));
            console.log(`PASS ${fixture.name}: ${JSON.stringify(actual)}`);
        }
        checkJavaScript(stockEmit, adaptedEmit);
    } finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

module.exports = { cases, answers, fromSnapshot, emit, load };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
    run(path.resolve(process.argv[2])).catch(error => { console.error(error); process.exitCode = 1; });
}
