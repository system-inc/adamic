'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const ts = require('typescript');
const [beforeLane, afterLane, output] = process.argv.slice(2).map(p => path.resolve(p));
const before = path.join(beforeLane, 'adapted-tree');
const after = path.join(afterLane, 'adapted-tree');
fs.mkdirSync(output, {recursive: true});
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const walk = d => fs.readdirSync(d, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(d, e.name)) : [path.join(d, e.name)]).sort();
function artifacts(tree) {
    return Object.fromEntries(walk(path.join(tree, 'built')).filter(f => /\.(js|mjs|cjs)$/.test(f) || /\/typescript\.d\.(ts|mts|cts)$/.test(f)).map(f => [path.relative(tree, f), hash(fs.readFileSync(f))]));
}
const old = artifacts(before), current = artifacts(after);
assert(Object.keys(old).length > 0);
assert.deepEqual(current, old, 'all emitted JavaScript and public API bytes');
const oldLane = JSON.parse(fs.readFileSync(path.join(beforeLane, 'report.json')));
const newLane = JSON.parse(fs.readFileSync(path.join(afterLane, 'report.json')));
for (const key of ['status', 'counts', 'failed_tests', 'baseline_diffs', 'api', 'platform', 'verdict']) assert.deepEqual(newLane[key], oldLane[key], 'lane identical: ' + key);
assert.equal(fs.readFileSync(path.join(beforeLane, 'oracle/baseline.diff'), 'utf8'), fs.readFileSync(path.join(afterLane, 'oracle/baseline.diff'), 'utf8'));
function fixture(tree) {
    const compiler = require(path.join(tree, 'built/local/typescript.js'));
    const parse = text => compiler.parseJsonText('fixture.json', text);
    const errors = [];
    assert.deepEqual(compiler.convertToObject(parse('{"a": [true, false, null, "s", 7, -2], "b": {"c": 1}}'), errors), {a: [true, false, null, 's', 7, -2], b: {c: 1}});
    assert.equal(errors.length, 0);
    assert.deepEqual(compiler.convertToObject(parse('{"bad": invalid, "array": [invalid, 2]}'), []), {bad: undefined, array: [2]});
    assert.equal(compiler.convertToJson(parse('{}'), parse('{}').statements[0].expression, [], false, undefined), undefined);
    assert.equal(compiler.convertToJson(parse('true'), parse('true').statements[0].expression, [], false, undefined), true);
    assert.deepEqual(compiler.parseConfigFileTextToJson('fixture.json', '[{"a":1}]').config, {a: 1});
    assert.deepEqual(compiler.parseConfigFileTextToJson('fixture.json', '7').config, {});
    assert.equal(compiler.tryParseJson('7'), 7);
    assert.equal(compiler.tryParseJson('null'), null);
    assert.equal(compiler.tryParseJson('invalid'), undefined);
    const Node = compiler.objectAllocator.getNodeConstructor();
    const node = new Node(compiler.SyntaxKind.Identifier, 1, 2);
    assert.equal(node.kind, compiler.SyntaxKind.Identifier);
    assert.equal(node.pos, 1); assert.equal(node.end, 2);
    const timer = setTimeout(() => {}, 10000);
    assert.equal(typeof timer.ref, 'function'); clearTimeout(timer);
}
fixture(before); fixture(after);
function returnTypes(tree) {
    const configFile = path.join(tree, 'src/compiler/tsconfig.json');
    const config = ts.parseJsonConfigFileContent(ts.readConfigFile(configFile, ts.sys.readFile).config, ts.sys, path.dirname(configFile), undefined, configFile);
    assert.equal(config.errors.length, 0);
    const program = ts.createProgram(config.fileNames, config.options);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length, 0, 'all compiler consumers type check');
    const checker = program.getTypeChecker(), result = {};
    const names = new Set(['convertConfigFileToObject', 'convertToObject', 'convertToJson', 'convertObjectLiteralExpressionToJson', 'convertPropertyValueToJson', 'tryParseJson']);
    for (const file of ['commandLineParser.ts', 'utilities.ts', 'sys.ts']) {
        const source = program.getSourceFile(path.join(tree, 'src/compiler', file));
        function visit(node) {
            let name;
            if (ts.isFunctionDeclaration(node) && names.has(node.name?.text)) name = node.name.text;
            if (ts.isArrowFunction(node) && ts.isPropertyAssignment(node.parent) && node.parent.name.getText(source) === 'getNodeConstructor') name = 'getNodeConstructor';
            if (name) {
                const type = checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(node));
                result[name] = {type: checker.typeToString(type), any: Boolean(type.flags & ts.TypeFlags.Any)};
            }
            ts.forEachChild(node, visit);
        }
        visit(source);
    }
    assert.equal(Object.keys(result).length, 7);
    return result;
}
const originalReturns = returnTypes(before), adaptedReturns = returnTypes(after);
for (const [name, original] of Object.entries(originalReturns)) {
    assert.equal(original.any, true, 'baseline any: ' + name);
    assert.equal(adaptedReturns[name].any, ['convertToObject', 'tryParseJson'].includes(name), 'adapted contract: ' + name);
}
const {rules} = require('./rules.json');
const files = [...new Set(rules.map(r => r.file))];
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'real-any-mutants-'));
function reset() {
    for (const f of files) { fs.mkdirSync(path.dirname(path.join(scratch, f)), {recursive: true}); fs.copyFileSync(path.join(before, f), path.join(scratch, f)); }
}
function apply() { return spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {encoding: 'utf8'}); }
reset(); assert.equal(apply().status, 0);
const first = Object.fromEntries(files.map(f => [f, fs.readFileSync(path.join(scratch, f), 'utf8')]));
assert.equal(apply().status, 0);
for (const f of files) assert.equal(fs.readFileSync(path.join(scratch, f), 'utf8'), first[f], 'idempotent: ' + f);
const sites = [
    ['convertConfigFileToObject', 'src/compiler/commandLineParser.ts', '): any {', 'function convertConfigFileToObject('],
    ['convertToObject', 'src/compiler/commandLineParser.ts', '): any {', 'export function convertToObject('],
    ['convertToJson', 'src/compiler/commandLineParser.ts', '): any {', 'export function convertToJson('],
    ['convertObjectLiteralExpressionToJson', 'src/compiler/commandLineParser.ts', '): any {', 'function convertObjectLiteralExpressionToJson('],
    ['convertPropertyValueToJson', 'src/compiler/commandLineParser.ts', '): any {', 'function convertPropertyValueToJson('],
    ['tryParseJson', 'src/compiler/utilities.ts', '): any {', 'export function tryParseJson('],
    ['getNodeConstructor', 'src/compiler/utilities.ts', 'Node as any', 'getNodeConstructor:'],
];
const mutants = [];
for (const [name, file, token, anchor] of sites) {
    reset(); const f = path.join(scratch, file); const text = fs.readFileSync(f, 'utf8');
    const start = text.indexOf(anchor); assert(start >= 0); const pos = text.indexOf(token, start); assert(pos >= start);
    fs.writeFileSync(f, text.slice(0, pos) + token.replace('any', 'string') + text.slice(pos + token.length));
    const result = apply();
    assert.notEqual(result.status, 0, 'site mutant rejected: ' + name);
    assert.match(result.stderr, /unreviewed site:/);
    fs.writeFileSync(path.join(output, name + '-mutant.log'), result.stderr);
    mutants.push({site: name, mutation: 'any changed to string', check: 'reviewed source contract guard', exit: result.status});
}
const changed = {...current}; changed[Object.keys(changed).find(f => f.endsWith('.js'))] = hash('mutated output');
assert.throws(() => assert.deepEqual(changed, old));
const changedAPI = {...current}; const api = Object.keys(changedAPI).find(f => f.endsWith('/typescript.d.ts')); assert(api);
changedAPI[api] = hash('changed public declaration'); assert.throws(() => assert.deepEqual(changedAPI, old));
const changedLane = {...newLane.counts, passing: newLane.counts.passing - 1}; assert.throws(() => assert.deepEqual(changedLane, oldLane.counts));
const report = {base: oldLane.execution.commit, node: process.version, typescript: ts.version, returnTypes: {before: originalReturns, after: adaptedReturns}, adaptedSites: 5, declinedSites: 2, adaptedHiddenBytes: 6947, jsFiles: Object.keys(current).filter(f => /\.(js|mjs|cjs)$/.test(f)).length, publicAPIFiles: Object.keys(current).filter(f => /typescript\.d\.(ts|mts|cts)$/.test(f)), outputs: current, lane: {status: newLane.status, counts: newLane.counts, failed_tests: newLane.failed_tests, baseline_diffs: newLane.baseline_diffs}, fixtures: 'recovery scalars, arrays, objects, invalid values, root recovery, disabled construction, allocator, Node timer, parse primitives', idempotent: true, mutants, comparisonMutants: ['JavaScript hash mutant', 'public declaration hash mutant', 'lane count mutant']};
fs.writeFileSync(path.join(output, 'proof.json'), JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify({adaptedSites: 5, declinedSites: 2, jsFiles: report.jsFiles, publicAPIFiles: report.publicAPIFiles, mutants: mutants.length, lane: report.lane}));
