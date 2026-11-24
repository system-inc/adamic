'use strict';

const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
const { execFileSync, spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

const input = process.argv[2];
assert(input && process.argv.length === 3, 'usage: node test.cjs <applied tree>');
assert.equal(ts.version, '6.0.3');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'checker-links-'));
const compiler = path.join(scratch, 'src/compiler');
const adapter = path.join(__dirname, 'adapt.cjs');
const ref = 'refs/stage3/99-seats-build22-98-links-classes';
function snapshot(side, file) {
    return execFileSync('git', ['-C', path.resolve(input), 'show', `${ref}-${side}:src/compiler/${file}`],
        { encoding: 'utf8', maxBuffer: 8 * 1024 * 1024 });
}

function declaration(text, name, isKind) {
    const source = ts.createSourceFile('input.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const matches = source.statements.filter(node => isKind(node) && node.name?.text === name);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source);
}

function symbolLinksDrive(original, adapted) {
    const parse = text => ts.createSourceFile('checker.ts', text, ts.ScriptTarget.Latest, true);
    const base = parse(original);
    const head = parse(adapted);
    const [statement] = base.statements.filter(node => ts.isVariableStatement(node) &&
        node.declarationList.declarations.some(binding =>
            binding.name.getText(base) === 'SymbolLinks' && binding.initializer && ts.isClassExpression(binding.initializer)));
    assert(statement);
    const changed = declaration(adapted, 'SymbolLinks', ts.isClassDeclaration) + '\n' +
        declaration(adapted, 'SymbolLinks', ts.isInterfaceDeclaration);

    const allocations = [];
    function visit(node) {
        if (ts.isNewExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'SymbolLinks') {
            let owner = node.parent;
            while (owner && !(ts.isFunctionLike(owner) && owner.name)) owner = owner.parent;
            assert(owner);
            allocations.push({ owner: owner.name.getText(head), text: node.getText(head) });
        }
        ts.forEachChild(node, visit);
    }
    visit(head);
    assert.deepEqual(allocations.map(entry => entry.owner), ['createSymbol', 'getSymbolLinks']);

    function load(text) {
        const emitted = ts.transpileModule(text, {
            compilerOptions: { target: ts.ScriptTarget.ES2020 },
            reportDiagnostics: true,
        });
        assert.deepEqual(emitted.diagnostics, []);
        const context = vm.createContext({});
        vm.runInContext(emitted.outputText, context);
        return context;
    }
    function observe(context, allocation) {
        const first = vm.runInContext(allocation, context);
        const second = vm.runInContext(allocation, context);
        const ctor = vm.runInContext('SymbolLinks', context);
        const names = Object.getOwnPropertyNames(first);
        assert.deepEqual(names, []);
        assert.notEqual(first, second);
        assert(first instanceof ctor);
        const prototype = Object.getPrototypeOf(first);
        assert.equal(prototype, ctor.prototype);
        const parent = Object.getPrototypeOf(prototype);
        assert.equal(parent, vm.runInContext('Object.prototype', context));
        first.witness = 42;
        assert.equal(first.witness, 42);
        return {
            name: ctor.name, kind: typeof ctor, names, prototypeNames: Object.getOwnPropertyNames(prototype),
            parentNames: Object.getOwnPropertyNames(parent), distinct: first !== second,
            instance: first instanceof ctor, field: first.witness,
        };
    }
    const before = load(statement.getText(base));
    const after = load(changed);
    for (const { owner, text } of allocations) {
        assert.deepEqual(observe(after, text), observe(before, text), owner);
        console.log(`PASS SymbolLinks ${owner}: ${text}`);
    }
}

const observations = [
    ['computed template', ['const flagKey = `calculated${"Flags"}` as const;', 'if (links[flagKey] === 0) throw new Error("unreviewed fresh links observation");'], 'computed access', 2],
    ['computed concatenation', ['if (links["calculated" + "Flags"] === 0) throw new Error("unreviewed fresh links observation");'], 'computed access', 1],
    ['computed local', ['const key = ["flags", "calculated" + "Flags"][1] as keyof typeof links;', 'if (links[key] === 0) throw new Error("unreviewed fresh links observation");'], 'computed access', 2],
    ['Object.keys', ['if (Object.keys(links).length !== 1) throw new Error("unreviewed fresh links observation");'], 'CallExpression', 1],
    ['spread copy', ['const copy = { ...links };'], 'SpreadAssignment', 1],
    ['for...in', ['for (const key in links) if (key.length > 15) throw new Error("unreviewed fresh links observation");'], 'ForInStatement', 1],
    ['Object.assign', ['Object.assign(Object.create(null), links);'], 'CallExpression', 1],
    ['P1 arrow return', ['const leak = () => { return links; };', 'if (Object.keys(leak()).length !== 1) throw new Error("unreviewed fresh links observation");'], 'ReturnStatement', 1],
    ['P2 function expression return', ['const leak = function () { return links; };', 'if (Object.keys(leak()).length !== 1) throw new Error("unreviewed fresh links observation");'], 'ReturnStatement', 1],
    ['accessor alias', ['const alias = getNodeLinks;'], 'accessor escape', 1],
];

function observeLinks(text, adapted, middle) {
    const eol = text.includes('\r\n') ? '\r\n' : '\n';
    const allocation = adapted ? 'new NodeLinks()' : 'new (NodeLinks as any)()';
    const expression = `nodeLinks[nodeId] || (nodeLinks[nodeId] = ${allocation})`;
    const anchor = `        return ${expression};`;
    assert.equal(text.split(anchor).length, 2, 'one getNodeLinks return anchor');
    const replacement = [`const links: typeof nodeLinks[number] = ${expression};`, ...middle, 'return links;']
        .map(line => '        ' + line).join(eol);
    return { text: text.replace(anchor, replacement), line: text.slice(0, text.indexOf(anchor)).split(eol).length };
}

function observeImportArguments(text) {
    const source = ts.createSourceFile('checker.ts', text, ts.ScriptTarget.Latest, true);
    const checker = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'createTypeChecker');
    const imports = checker.body.statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'resolveImportSymbolType');
    assert.equal(imports.length, 1);
    const start = imports[0].body.statements[0].getStart(source);
    const { line } = source.getLineAndCharacterOfPosition(start);
    const eol = text.includes('\r\n') ? '\r\n' : '\n';
    const indent = text.slice(source.getPositionOfLineAndCharacter(line, 0), start);
    const observation = 'if (Object.keys(arguments[1]).length !== 1) throw new Error("unreviewed fresh links observation");';
    return { text: text.slice(0, start) + observation + eol + indent + text.slice(start), line: line + 1 };
}

try {
    fs.mkdirSync(compiler, { recursive: true });
    const checker = path.join(compiler, 'checker.ts');
    const definitions = path.join(compiler, 'types.ts');
    const original = snapshot('before', 'checker.ts');
    const adapted = snapshot('after', 'checker.ts');
    const types = snapshot('before', 'types.ts');
    const afterTypes = snapshot('after', 'types.ts');
    fs.writeFileSync(checker, original);
    fs.writeFileSync(definitions, types);
    const run = () => JSON.parse(execFileSync(process.execPath, [adapter, scratch], { encoding: 'utf8' }));
    assert.deepEqual(run(), { files: 1, edits: 11, javascriptChanging: true });
    assert(fs.readFileSync(checker).equals(Buffer.from(adapted)), 'checker.ts differs from after snapshot');
    assert.equal(types, afterTypes, 'types.ts snapshots differ');
    assert(fs.readFileSync(definitions).equals(Buffer.from(types)), 'adapter wrote types.ts');
    console.log('PASS adapter output matches its after snapshot, 11 edits');
    assert.equal(adapted.split('\n').length, original.split('\n').length);
    const enums = ['NodeCheckFlags', 'JsxFlags'].map(name => declaration(types, name, ts.isEnumDeclaration)).join('\n');
    function load(constructor) {
        const result = ts.transpileModule(`${enums}\n${constructor}`, {
            compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
            reportDiagnostics: true,
        });
        assert.deepEqual(result.diagnostics, []);
        return vm.runInNewContext(result.outputText + '\n({ NodeLinks, NodeCheckFlags, JsxFlags });', { exports: {} });
    }
    const before = load(declaration(original, 'NodeLinks', ts.isFunctionDeclaration));
    const after = load(declaration(adapted, 'NodeLinks', ts.isClassDeclaration));
    assert.equal(new before.NodeLinks().flags, new after.NodeLinks().flags);
    for (const [field, flags, names] of [
        ['calculatedFlags', 'NodeCheckFlags', ['SuperInstance', 'CaptureArguments', 'ConstructorReference', 'CapturedBlockScopedBinding']],
        ['jsxFlags', 'JsxFlags', ['IntrinsicNamedElement', 'IntrinsicIndexedElement']],
    ]) {
        for (const name of names) {
            const mask = before[flags][name];
            assert(mask);
            const old = new before.NodeLinks();
            const current = new after.NodeLinks();
            assert.equal(current[field] & mask, old[field] & mask, `${field} fresh & ${name}`);
            assert.equal(current[field] |= mask, old[field] |= mask, `${field} |= ${name}`);
            assert.equal(current[field] & mask, old[field] & mask, `${field} after |= ${name}`);
            console.log(`PASS ${field} ${name}: fresh & 0, |= ${mask}, subsequent & ${mask}`);
        }
    }
    symbolLinksDrive(original, adapted);
    const hashes = () => [checker, definitions].map(file =>
        createHash('sha256').update(fs.readFileSync(file)).digest('hex'));
    const prepared = { files: 1, edits: 11, javascriptChanging: true };
    const eol = original.includes('\r\n') ? '\r\n' : '\n';

    const replay = hashes();
    assert.deepEqual(run(), { files: 0, edits: 0, javascriptChanging: true });
    assert.deepEqual(hashes(), replay);
    console.log(`PASS replay 0 edits: ${replay.join(' ')} -> ${hashes().join(' ')}`);
    function reject(label, checkerText, typesText, diagnostic) {
        fs.writeFileSync(checker, checkerText);
        fs.writeFileSync(definitions, typesText);
        const before = hashes();
        const result = spawnSync(process.execPath, [adapter, scratch], { encoding: 'utf8' });
        assert.ifError(result.error);
        const after = hashes();
        assert.equal(result.status, 1, `${label} must refuse, got exit ${result.status}: ${result.stdout}before ${before.join(' ')} -> after ${after.join(' ')}`);
        assert.match(result.stderr, diagnostic);
        assert.deepEqual(after, before);
        console.log(`PASS ${label} unchanged: ${before.join(' ')} -> ${after.join(' ')}`);
        console.log(result.stderr.split('\n').find(line => line.startsWith('AssertionError')));
    }
    const line = original.split('\n').length;
    reject('appended calculatedFlags read', original +
        `function seatsPlantedRead(links: NodeLinks) { return links.calculatedFlags & NodeCheckFlags.SuperInstance; }${eol}`,
        types, new RegExp(`unaudited flag access at checker.ts:${line}: calculatedFlags`));
    reject('appended jsxFlags read', original + `function extraRead(links: NodeLinks) { return links.jsxFlags; }${eol}`,
        types, new RegExp(`unaudited flag access at checker.ts:${line}: jsxFlags`));
    reject('string-keyed jsxFlags read', original + `function extraRead(links: NodeLinks) { return links["jsxFlags"]; }${eol}`,
        types, new RegExp(`unaudited string flag key at checker.ts:${line}: jsxFlags`));
    reject('changed calculatedFlags type', original, types.replace('calculatedFlags: NodeCheckFlags;', 'calculatedFlags: number;'),
        /calculatedFlags: number;/);
    reject('half-adapted allocation', adapted.replace('new NodeLinks()', 'new (NodeLinks as any)()'),
        types, /new \(NodeLinks as any\)\(\)/);
    for (const [side, inputText, typeText] of [['before', original, types], ['after', adapted, afterTypes]]) {
        for (const [label, middle, role, offset] of observations) {
            const drift = observeLinks(inputText, side === 'after', middle);
            reject(`${side} ${label}`, drift.text, typeText,
                new RegExp(`unaudited NodeLinks ${role} at checker\\.ts:${drift.line + offset} \\(getNodeLinks\\)`));
        }
        const drift = observeImportArguments(inputText);
        reject(`${side} P3 arguments`, drift.text, typeText,
            new RegExp(`unaudited NodeLinks arguments at checker\\.ts:${drift.line} \\(resolveImportSymbolType\\)`));
    }
    const blank = text => text.replace(eol, eol + eol);
    const lf = text => text.replaceAll('\r\n', '\n');
    for (const [label, inputText, typeText, expected] of [
        ['line relocation', blank(original), types, blank(adapted)],
        ['LF input', lf(original), lf(types), lf(adapted)],
    ]) {
        fs.writeFileSync(checker, inputText);
        fs.writeFileSync(definitions, typeText);
        const before = hashes();
        assert.deepEqual(run(), prepared);
        assert.equal(fs.readFileSync(checker, 'utf8'), expected);
        const after = hashes();
        assert.equal(after[1], before[1]);
        console.log(`PASS ${label} 11 edits: ${before.join(' ')} -> ${after.join(' ')}`);
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
