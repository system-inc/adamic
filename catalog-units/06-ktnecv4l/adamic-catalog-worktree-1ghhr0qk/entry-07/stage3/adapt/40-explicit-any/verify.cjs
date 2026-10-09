// The stock parser/checker and byte hashes are independent of the adaptation.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [mode, treeArg, evidenceArg] = process.argv.slice(2);
const tree = path.resolve(treeArg), evidence = path.resolve(evidenceArg);
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
function walk(directory) {
    return fs.readdirSync(directory, {withFileTypes: true}).flatMap(entry => entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]).sort();
}
function equalOutputs(before, after) {
    assert.deepEqual(Object.keys(after), Object.keys(before), 'output file set');
    for (const file of Object.keys(before)) assert.equal(after[file], before[file], 'output bytes: ' + file);
}
function expectedDeclaration(text) {
    return text.replace('function length(array: readonly any[]', 'function length<T>(array: readonly T[]')
        .replace('function hasProperty(map: MapLike<any>', 'function hasProperty(map: object');
}
function anyCount(text) {
    const source = ts.createSourceFile('probe.ts', text, ts.ScriptTarget.Latest, true);
    let count = 0;
    function visit(node) { if (node.kind === ts.SyntaxKind.AnyKeyword) count++; ts.forEachChild(node, visit); }
    visit(source); return count;
}
function expectedSource(before, after) {
    const text = before.replace('function length(array: readonly any[]', 'function length<T>(array: readonly T[]')
        .replace('function toOffset(array: readonly any[]', 'function toOffset<T>(array: readonly T[]')
        .replace('function hasProperty(map: MapLike<any>', 'function hasProperty(map: object');
    assert.equal(after, text, 'only three reviewed declaration edits');
}
if (mode === 'mutants') {
    const catches = [];
    function killed(name, fn, match) {
        assert.throws(fn, match); catches.push(name);
    }
    killed('JavaScript byte mutation', () => equalOutputs({'a.js': hash('return 1;')}, {'a.js': hash('return 2;')}), /output bytes/);
    killed('missing emitted JavaScript file', () => equalOutputs({'a.js': 'x'}, {}), /output file set/);
    const before = fs.readFileSync(path.join(evidence, 'before-core.txt'), 'utf8');
    const after = fs.readFileSync(path.join(tree, 'src/compiler/core.ts'), 'utf8');
    killed('unreviewed function statement mutation', () => expectedSource(before, after.replace('array.length : 0', 'array.length : 1')), /only three/);
    killed('added explicit any token', () => assert.equal(anyCount(after + '\nlet mutant: any;'), anyCount(after), 'census count'), /census count/);
    const declarations = JSON.parse(fs.readFileSync(path.join(evidence, 'before-declarations.json')));
    const declaration = declarations['local/typescript.internal.d.ts'];
    const expected = expectedDeclaration(declaration);
    killed('unreviewed internal API declaration mutation', () => assert.equal(expected.replace('function length<T>', 'function length<U>'), expected, 'API declaration'), /API declaration/);
    const {spawnSync} = require('node:child_process');
    const scratch = fs.mkdtempSync(path.join(require('node:os').tmpdir(), 'explicit-any-mutants-'));
    const scratchCore = path.join(scratch, 'src/compiler/core.ts');
    fs.mkdirSync(path.dirname(scratchCore), {recursive: true});
    try {
        fs.writeFileSync(scratchCore, before);
        const first = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {encoding: 'utf8'});
        assert.equal(first.status, 0, first.stderr);
        assert.equal(fs.readFileSync(scratchCore, 'utf8'), after, 'final adapter first pass');
        for (const [name, mutant, diagnostic] of [
            ['changed helper body rejected', before.replace('array.length : 0', 'array.length : 1'), 'unreviewed uses: length'],
            ['unreviewed parameter type rejected', before.replace('function length(array: readonly any[]', 'function length(array: readonly string[]'), 'unreviewed signature: length'],
        ]) {
            fs.writeFileSync(scratchCore, mutant);
            const result = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {encoding: 'utf8'});
            assert.notEqual(result.status, 0, name);
            assert(result.stderr.includes(diagnostic), result.stderr);
            catches.push(name);
        }
    }
    finally { fs.rmSync(scratch, {recursive: true}); }
    const adapted = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), tree], {encoding: 'utf8'});
    assert.equal(adapted.status, 0, adapted.stderr);
    assert.equal(fs.readFileSync(path.join(tree, 'src/compiler/core.ts'), 'utf8'), after, 'idempotence');
    assert.equal(JSON.parse(adapted.stdout).removed, 0);
    killed('second-pass source byte mutation', () => assert.equal(after + '\n', after, 'idempotence'), /idempotence/);
    catches.push('adapter second pass changes zero source bytes');
    const configPath = path.join(tree, 'src/compiler/tsconfig.json');
    const read = ts.readConfigFile(configPath, ts.sys.readFile);
    const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
    const host = ts.createCompilerHost(config.options), original = host.readFile;
    host.readFile = file => { const text = original(file); return file === path.join(tree, 'src/compiler/core.ts') ? text.replace('return array !== undefined ? array.length : 0;', 'return "wrong";') : text; };
    const program = ts.createProgram(config.fileNames, config.options, host);
    assert(ts.getPreEmitDiagnostics(program).some(d => d.code === 2322 && d.file?.fileName.endsWith('/compiler/core.ts')), 'checker must detect erroneous return');
    catches.push('real compiler return type mutation caught by TS2322');
    fs.writeFileSync(path.join(evidence, 'mutants.json'), JSON.stringify(catches, null, 2) + '\n');
    console.log(JSON.stringify(catches)); process.exit(0);
}
assert(['before', 'after'].includes(mode));
fs.mkdirSync(evidence, {recursive: true});
const census = [];
const sources = {};
for (const file of walk(path.join(tree, 'src/compiler')).filter(f => f.endsWith('.ts') && !f.includes('.generated.'))) {
    const text = fs.readFileSync(file, 'utf8');
    sources[path.relative(tree, file)] = hash(text);
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    function visit(node) {
        if (node.kind === ts.SyntaxKind.AnyKeyword) {
            const pos = source.getLineAndCharacterOfPosition(node.getStart(source));
            census.push({file: path.relative(tree, file), line: pos.line + 1, column: pos.character + 1, source: text.split(/\r?\n/)[pos.line]});
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const diagnostics = ts.getPreEmitDiagnostics(program).map(d => {
    const pos = d.file?.getLineAndCharacterOfPosition(d.start);
    return {code: d.code, file: d.file ? path.relative(tree, d.file.fileName) : null, line: pos ? pos.line + 1 : null, column: pos ? pos.character + 1 : null, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')};
});
const outputs = {}, declarations = {};
for (const file of walk(path.join(tree, 'built'))) {
    if (/\.(?:js|mjs|cjs)$/.test(file)) outputs[path.relative(path.join(tree, 'built'), file)] = hash(fs.readFileSync(file));
    if (/\.d\.ts$/.test(file)) declarations[path.relative(path.join(tree, 'built'), file)] = fs.readFileSync(file, 'utf8');
}
const snapshot = {typescript: ts.version, census, sources, diagnostics, outputs};
fs.writeFileSync(path.join(evidence, mode + '.json'), JSON.stringify(snapshot, null, 2) + '\n');
const core = fs.readFileSync(path.join(tree, 'src/compiler/core.ts'), 'utf8');
if (mode === 'before') {
    fs.writeFileSync(path.join(evidence, 'before-core.txt'), core);
    fs.writeFileSync(path.join(evidence, 'before-declarations.json'), JSON.stringify(declarations));
    assert.equal(census.length, 210);
}
else {
    const before = JSON.parse(fs.readFileSync(path.join(evidence, 'before.json')));
    const source = fs.readFileSync(path.join(evidence, 'before-core.txt'), 'utf8');
    expectedSource(source, core);
    for (const file of Object.keys(before.sources)) if (file !== 'src/compiler/core.ts') assert.equal(sources[file], before.sources[file], 'unowned source changed: ' + file);
    assert.equal(census.length, 207);
    equalOutputs(before.outputs, outputs);
    const prior = JSON.parse(fs.readFileSync(path.join(evidence, 'before-declarations.json')));
    assert.deepEqual(Object.keys(prior), Object.keys(declarations));
    const changedDeclarations = [];
    for (const file of Object.keys(prior)) {
        const expected = expectedDeclaration(prior[file]);
        assert.equal(declarations[file], expected, 'unexplained declaration output: ' + file);
        if (prior[file] !== declarations[file]) changedDeclarations.push(file);
    }
    const newlyExposed = diagnostics.filter(d => !before.diagnostics.some(old => JSON.stringify(old) === JSON.stringify(d)));
    fs.writeFileSync(path.join(evidence, 'consumer-errors.json'), JSON.stringify(newlyExposed, null, 2) + '\n');
    fs.writeFileSync(path.join(evidence, 'proof.json'), JSON.stringify({before: before.census.length, after: census.length, jsFiles: Object.keys(outputs).length, declarationFiles: Object.keys(declarations).length, changedDeclarations, newlyExposed}, null, 2) + '\n');
}
console.log(JSON.stringify({mode, tokens: census.length, diagnostics: diagnostics.length, jsFiles: Object.keys(outputs).length}));
