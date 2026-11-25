'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const { plan } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs TREE');
assert.equal(ts.version, '6.0.3');
const tree = path.resolve(process.argv[2]);
const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'parse-config-host-'));
const ref = 'refs/stage3/99-seats-scout24-parse-config-host';
const handler = 'onUnRecoverableConfigFileDiagnostic';
const constructionReads = ['useCaseSensitiveFileNames', 'directoryExists', 'realpath'];
const forwarded = new Set(['readFile', 'fileExists', 'getCurrentDirectory', 'readDirectory']);

function functionText(source, name) {
    const matches = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source).replace(/^export /, '');
}

function compile(side) {
    const root = path.join(temp, side);
    fs.mkdirSync(root);
    const archive = execFileSync('git', ['-C', tree, 'archive', ref + '-' + side, 'src', 'package.json'],
        { maxBuffer: 100 * 1024 * 1024 });
    execFileSync('tar', ['-xf', '-', '-C', root], { input: archive });
    const api = path.dirname(path.dirname(require.resolve('typescript/package.json')));
    fs.symlinkSync(api, path.join(root, 'node_modules'), 'dir');
    execFileSync(process.execPath, [require.resolve('typescript/lib/tsc.js'),
        '-b', 'src/compiler', '--pretty', 'false', '--emitDeclarationOnly', 'false'],
    { cwd: root, stdio: 'inherit' });

    const watchPath = path.join(root, 'src/compiler/watch.ts');
    const source = ts.createSourceFile(watchPath, fs.readFileSync(watchPath, 'utf8'),
        ts.ScriptTarget.Latest, true);
    assert.deepEqual(source.parseDiagnostics, []);
    const text = [
        functionText(source, 'parseConfigFileWithSystem'),
        functionText(source, 'reportUnrecoverableDiagnostic'),
        'parseConfigFileWithSystem;',
    ].join('\n');
    const { outputText, diagnostics } = ts.transpileModule(text, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
        reportDiagnostics: true,
    });
    assert.deepEqual(diagnostics, []);
    const built = path.join(root, 'built/local/compiler');
    return {
        outputText,
        source,
        parser: require(path.join(built, 'commandLineParser.js')).getParsedCommandLineOfConfigFile,
        ExitStatus: require(path.join(built, 'types.js')).ExitStatus,
    };
}

function answer(value) {
    if (typeof value === 'function') return { kind: 'function' };
    if (value === undefined) return { kind: 'undefined' };
    return { kind: 'value', value };
}

function summarizeDiagnostic(diagnostic) {
    const { code, category, messageText, file, start, length } = diagnostic;
    return { code, category, messageText, file: file?.fileName, start, length };
}

function run(compiled, scenario) {
    const calls = [];
    const systemRecord = [];
    const hostRecord = [];
    const diagnostics = [];
    const files = new Map(Object.entries(scenario.files));
    const stop = Symbol('exit');
    let system;
    const method = (name, body) => function (...args) {
        calls.push([name, args, this === system]);
        systemRecord.push(['call', name]);
        return body(...args);
    };
    const target = {
        useCaseSensitiveFileNames: scenario.caseSensitive ?? true,
        readFile: method('readFile', name => files.get(name)),
        fileExists: method('fileExists', name => files.has(name)),
        getCurrentDirectory: method('getCurrentDirectory', () => '/project'),
        readDirectory: method('readDirectory', root => root === '/project' ? ['/project/index.ts'] : []),
        exit: method('exit', () => { throw stop; }),
    };
    if (scenario.optional) {
        target.directoryExists = method('directoryExists', directory => directory === '/' ||
            [...files.keys()].some(name => name.startsWith(directory + '/')));
        target.realpath = method('realpath', scenario.realpath ?? (name => name));
    }

    system = new Proxy(target, {
        get(object, key, receiver) {
            const value = Reflect.get(object, key, receiver);
            systemRecord.push(['get', key, answer(value)]);
            return value;
        },
        has(object, key) {
            const present = Reflect.has(object, key);
            systemRecord.push(['has', key, present]);
            return present;
        },
    });

    let parserCalls = 0;
    const parse = vm.runInNewContext(compiled.outputText, {
        ExitStatus: compiled.ExitStatus,
        getParsedCommandLineOfConfigFile(config, options, host, cache, watchOptions) {
            parserCalls++;
            const observed = new Proxy(host, {
                get(object, key) {
                    const value = Reflect.get(object, key, object);
                    hostRecord.push(['get', key, answer(value)]);
                    return typeof value === 'function' ? value.bind(object) : value;
                },
                has(object, key) {
                    const present = Reflect.has(object, key);
                    hostRecord.push(['has', key, present]);
                    return present;
                },
            });
            return compiled.parser(config, options, observed, cache, watchOptions);
        },
    });

    let result;
    let exited = false;
    try {
        const parsed = parse('/project/tsconfig.json', {}, new Map(), undefined, system,
            diagnostic => diagnostics.push(summarizeDiagnostic(diagnostic)));
        if (parsed) {
            const { options, fileNames, errors } = parsed;
            result = JSON.parse(JSON.stringify({ options, fileNames,
                errors: errors.map(summarizeDiagnostic) }));
        }
    } catch (error) {
        if (error !== stop) throw error;
        exited = true;
    }
    assert.equal(parserCalls, 1);
    const property = Object.getOwnPropertyDescriptor(target, handler);
    return { calls, systemRecord, hostRecord, result, diagnostics, exited,
        key: property ? answer(property.value) : { kind: 'absent' } };
}

function compare(base, head, scenario) {
    const message = scenario.name + ' ';
    if (scenario.realpath) {
        assert.deepEqual(head.result?.fileNames, base.result?.fileNames,
            message + '(iv) realpath dependent file list');
    }
    assert.deepEqual(head.hostRecord, base.hostRecord, message + '(ii) parser host');
    assert.deepEqual(head.calls, base.calls, message + '(i) System calls and receivers');
    assert(base.calls.every(([, , correct]) => correct), message + '(i) System receiver');
    assert.equal(base.exited, !!scenario.missing, message + '(iv) expected exit');
    assert.equal(head.exited, base.exited, message + '(iv) exit');
    assert.deepEqual(head.result, base.result, message + '(iv) parse result');
    assert.deepEqual(head.diagnostics, base.diagnostics, message + '(iv) diagnostics');
    assert.deepEqual(base.key, { kind: scenario.missing ? 'function' : 'undefined' },
        message + '(v) original key');
    assert.deepEqual(head.key, { kind: 'absent' }, message + '(v) adapted key');

    const baseReads = base.systemRecord.filter(([operation]) => operation !== 'call');
    assert.deepEqual(baseReads.filter(([, member]) => member !== 'exit'), base.hostRecord,
        message + '(iii) stock System reads');
    const init = constructionReads.map(member => ['get', member,
        answer(member === 'useCaseSensitiveFileNames' ? (scenario.caseSensitive ?? true) :
            scenario.optional ? (() => {}) : undefined)]);
    assert.deepEqual(head.systemRecord.slice(0, init.length), init, message + '(iii) construction');
    const runtime = [];
    for (const [name] of head.calls) {
        if (forwarded.has(name) || name === 'exit') {
            runtime.push(['get', name, { kind: 'function' }]);
        }
        runtime.push(['call', name]);
    }
    assert.deepEqual(head.systemRecord.slice(init.length), runtime, message + '(iii) forwarded reads');

    if (scenario.missing) {
        assert.deepEqual(base.calls.at(-1), ['exit', [1], true]);
        assert.deepEqual(base.diagnostics.map(diagnostic => diagnostic.code), [5083]);
    } else {
        assert.deepEqual(base.result.fileNames, scenario.expectedFiles ?? ['/project/index.ts']);
        assert.deepEqual(base.result.errors, []);
        assert.equal(base.diagnostics.length, 0);
        if (scenario.extends) assert.equal(base.result.options.strict, true);
    }
    if (scenario.package) {
        for (const member of ['directoryExists', 'realpath']) {
            assert(base.hostRecord.some(([op, name]) => op === 'get' && name === member),
                message + 'fixture exercises ' + member);
            assert.equal(base.calls.some(([name]) => name === member), scenario.optional,
                message + 'fixture calls ' + member);
        }
    }
}

const file = { '/project/index.ts': 'export const n = 1;' };
const config = options => JSON.stringify({ include: ['*.ts'], ...options });
const packageFiles = {
    ...file,
    '/project/tsconfig.json': config({ extends: 'example-config' }),
    '/project/node_modules/example-config/package.json': JSON.stringify({ tsconfig: 'base.json' }),
    '/project/node_modules/example-config/base.json': JSON.stringify({ compilerOptions: { strict: true } }),
};
const packageRoot = '/project/node_modules/example-config';
const storeRoot = '/store/example-config';
const packageConfig = JSON.stringify({ files: ['settings.ts'], compilerOptions: { strict: true } });
const scenarios = [
    { name: 'valid config', files: { ...file, '/project/tsconfig.json': config({}) } },
    { name: 'missing config', files: file, missing: true },
    { name: 'case insensitive config', caseSensitive: false, files: {
        ...file,
        '/project/tsconfig.json': JSON.stringify({ files: ['INDEX.ts', 'index.ts'] }),
    } },
    { name: 'relative extends', extends: true, optional: true, files: {
        ...file,
        '/project/tsconfig.json': config({ extends: './base.json' }),
        '/project/base.json': JSON.stringify({ compilerOptions: { strict: true } }),
    } },
    { name: 'package extends with optional methods', extends: true, package: true,
        optional: true, files: packageFiles },
    { name: 'package extends without optional methods', extends: true, package: true,
        optional: false, files: packageFiles },
    { name: 'package extends with alternate realpath spelling', extends: true, package: true,
        optional: true, realpath: name => name.replace(packageRoot + '/', storeRoot + '/'),
        expectedFiles: [storeRoot + '/settings.ts'], files: {
            ...packageFiles,
            '/project/tsconfig.json': JSON.stringify({ extends: 'example-config' }),
            [packageRoot + '/base.json']: packageConfig,
            [storeRoot + '/base.json']: packageConfig,
            [packageRoot + '/settings.ts']: 'export const settings = true;',
            [storeRoot + '/settings.ts']: 'export const settings = true;',
        } },
];

function checkDuplicateImport(source) {
    const imports = source.statements.filter(node => ts.isImportDeclaration(node) &&
        ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === './_namespaces/ts.js');
    assert.equal(imports.length, 1);
    const end = imports[0].end;
    const newline = /^\r?\n/.exec(source.text.slice(end))?.[0];
    assert(newline, 'Expected the namespace import to end on its own line');
    const duplicate = source.text.slice(0, end) + newline +
        "import { memoize } from './_namespaces/ts.js';" + source.text.slice(end);
    assert.throws(() => plan(duplicate), /Expected one compiler namespace import/);
    console.log('PASS second namespace import with single quotes');
}

function checkBoundMembers(source) {
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'parseConfigFileWithSystem');
    assert.equal(owners.length, 1, 'Expected one parseConfigFileWithSystem declaration');
    const host = owners[0].body?.statements[0];
    assert(host && ts.isVariableStatement(host), 'Expected host declaration');
    const declarations = host.declarationList.declarations;
    assert.equal(declarations.length, 1, 'Expected one host declaration');
    const declaration = declarations[0];
    assert(ts.isIdentifier(declaration.name) && declaration.name.text === 'host', 'Expected host name');
    const literal = declaration.initializer;
    assert(literal && ts.isObjectLiteralExpression(literal), 'Expected host object literal');

    for (const member of ['directoryExists', 'realpath']) {
        const text = 'system.' + member + '?.bind(system)';
        const properties = literal.properties.filter(property =>
            property.name && ts.isIdentifier(property.name) && property.name.text === member);
        assert.equal(properties.length, 1, 'Expected one ' + member + ' property');
        const property = properties[0];
        assert(ts.isPropertyAssignment(property), 'Expected ' + member + ' property assignment');
        const call = property.initializer;
        assert(ts.isCallExpression(call), 'Expected ' + member + ' bind call');
        const bind = call.expression;
        assert(ts.isPropertyAccessExpression(bind) && bind.name.text === 'bind' && bind.questionDotToken,
            'Expected ' + member + ' optional bind');
        const target = bind.expression;
        assert(ts.isPropertyAccessExpression(target) && target.name.text === member &&
            ts.isIdentifier(target.expression) && target.expression.text === 'system',
            'Expected ' + member + ' on system');
        assert.equal(call.arguments.length, 1, 'Expected one ' + member + ' bind argument');
        const receiver = call.arguments[0];
        assert(ts.isIdentifier(receiver) && receiver.text === 'system', 'Expected system bind receiver');
        assert.equal(call.getText(source), text);
        console.log('PASS ' + member + ' written as ' + text);
    }
}

function countMaybeBind(source) {
    let count = 0;
    function visit(node) {
        if (ts.isIdentifier(node) && node.text === 'maybeBind') count++;
        ts.forEachChild(node, visit);
    }
    visit(source);
    return count;
}

try {
    const before = compile('before');
    const after = compile('after');
    checkBoundMembers(after.source);
    assert.equal(countMaybeBind(before.source), 0, 'Unexpected maybeBind in before watch.ts');
    assert.equal(countMaybeBind(after.source), 0, 'Unexpected maybeBind in after watch.ts');
    console.log('PASS zero maybeBind identifiers in both watch.ts snapshots');
    checkDuplicateImport(before.source);
    for (const scenario of scenarios) {
        const base = run(before, scenario);
        const head = run(after, scenario);
        const controls = {};
        if (scenario.caseSensitive === false) {
            const sensitive = run(before, { ...scenario, caseSensitive: true });
            controls.stockCaseSensitiveFileNames = sensitive.result.fileNames;
            assert.notDeepEqual(controls.stockCaseSensitiveFileNames, base.result.fileNames,
                'Fixture must distinguish the two stock case sensitivity values');
        }
        if (scenario.realpath) {
            const identity = run(before, { ...scenario, realpath: undefined });
            controls.stockIdentityRealpathFileNames = identity.result.fileNames;
            assert.deepEqual(controls.stockIdentityRealpathFileNames, [packageRoot + '/settings.ts']);
            assert.notDeepEqual(controls.stockIdentityRealpathFileNames, base.result.fileNames,
                'Fixture must distinguish the two stock realpath spellings');
        }
        console.log(JSON.stringify({ case: scenario.name, base, head, controls }));
        compare(base, head, scenario);
        console.log('PASS ' + scenario.name);
    }
} finally {
    fs.rmSync(temp, { recursive: true, force: true });
}
