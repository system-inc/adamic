'use strict';
const assert = require('node:assert/strict'), fs = require('node:fs'), path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto'), os = require('node:os'), {spawnSync} = require('node:child_process'), ts = require('typescript');
const [beforeLane, afterLane, output] = process.argv.slice(2).map(p => path.resolve(p));
const before = path.join(beforeLane, 'adapted-tree'), after = path.join(afterLane, 'adapted-tree');
fs.mkdirSync(output, {recursive: true});
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const walk = d => fs.readdirSync(d, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(d, e.name)) : [path.join(d, e.name)]).sort();
const rules = require('./rules.json');
const file = rules[0].file, original = fs.readFileSync(path.join(before, file), 'utf8'), adapted = fs.readFileSync(path.join(after, file), 'utf8');
const nl = text => text.replace(/\r?\n/g, original.includes('\r\n') ? '\r\n' : '\n');
let expected = original;
for (const r of rules) { assert.equal(expected.split(nl(r.before)).length, 2, r.id); expected = expected.replace(nl(r.before), nl(r.after)); }
assert.equal(adapted, expected, 'only reviewed type edits in converter file');
const sourceFiles = walk(path.join(before, 'src')).filter(f => f.endsWith('.ts'));
assert.deepEqual(walk(path.join(after, 'src')).filter(f => f.endsWith('.ts')).map(f => path.relative(after, f)), sourceFiles.map(f => path.relative(before, f)), 'exact source file set');
for (const f of sourceFiles) {
    const relative = path.relative(before, f);
    if (relative !== file) assert.deepEqual(fs.readFileSync(path.join(after, relative)), fs.readFileSync(f), 'untouched source: ' + relative);
}
function functions(text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true), result = {};
    for (const node of source.statements) if (ts.isFunctionDeclaration(node) && node.name) result[node.name.text] = node;
    return {source, result};
}
const oldFunctions = functions(original), newFunctions = functions(adapted);
assert.equal(newFunctions.result.isCompilerOptionsValue.getText(newFunctions.source), oldFunctions.result.isCompilerOptionsValue.getText(oldFunctions.source), 'predicate remains exactly written');
for (const name of ['parseJsonConfigFileContent', 'convertCompilerOptionsFromJson', 'convertTypeAcquisitionFromJson', 'convertToObject']) {
    assert.equal(newFunctions.result[name].getText(newFunctions.source), oldFunctions.result[name].getText(oldFunctions.source), 'public owner unchanged: ' + name);
}
const options = {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext};
assert.equal(ts.transpileModule(adapted, {compilerOptions: options}).outputText, ts.transpileModule(original, {compilerOptions: options}).outputText, 'whole source JS bytes');
function artifacts(tree) {
    return Object.fromEntries(walk(path.join(tree, 'built')).filter(f => /\.(js|mjs|cjs)$/.test(f) || /\/typescript\.d\.(ts|mts|cts)$/.test(f)).map(f => [path.relative(tree, f), hash(fs.readFileSync(f))]));
}
const old = artifacts(before), current = artifacts(after);
assert(Object.keys(old).length > 0); assert.deepEqual(current, old, 'actual emitted JavaScript and public declaration bytes');
const oldLane = JSON.parse(fs.readFileSync(path.join(beforeLane, 'report.json'))), newLane = JSON.parse(fs.readFileSync(path.join(afterLane, 'report.json')));
assert.equal(oldLane.status, 'pass'); assert.equal(newLane.status, 'pass');
for (const key of ['status', 'counts', 'failed_tests', 'baseline_diffs', 'api', 'platform', 'verdict']) assert.deepEqual(newLane[key], oldLane[key], 'lane identical: ' + key);
assert.deepEqual(fs.readFileSync(path.join(beforeLane, 'oracle/baseline.diff')), fs.readFileSync(path.join(afterLane, 'oracle/baseline.diff')));
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'config-generics-mutants-'));
const target = path.join(scratch, file); fs.mkdirSync(path.dirname(target), {recursive: true});
const apply = () => spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {encoding: 'utf8'});
const mutants = [];
for (const r of rules) {
    // Drift a reviewed source token, rather than altering an expected result.
    const token = nl(r.before), mutant = token.replace('any', 'unknown') === token ? token.replace('map(values,', 'map([...values],') : token.replace('any', 'unknown');
    assert.notEqual(mutant, token);
    const changed = original.replace(token, mutant); fs.writeFileSync(target, changed);
    const result = apply(); assert.notEqual(result.status, 0, 'source site mutant: ' + r.id);
    assert(result.stderr.includes('unreviewed site: ' + r.id), result.stderr);
    assert.equal(fs.readFileSync(target, 'utf8'), changed, 'reject before writing: ' + r.id);
    fs.writeFileSync(path.join(output, r.id + '-mutant.log'), result.stderr);
    mutants.push({site: r.id, mutation: token.includes('any') ? 'raw any to unreviewed unknown' : 'copy list before mapping', caughtBy: 'exact reviewed source guard', exit: result.status});
}
const predicateToken = require('./predicate.json').text.replaceAll('\n', original.includes('\r\n') ? '\r\n' : '\n');
fs.writeFileSync(target, original.replace(predicateToken, predicateToken.replace('value: any', 'value: unknown')));
const predicateMutant = apply(); assert.notEqual(predicateMutant.status, 0); assert(predicateMutant.stderr.includes('protected predicate changed: S02'));
fs.writeFileSync(path.join(output, 'S02-mutant.log'), predicateMutant.stderr);
mutants.push({site: 'S02', mutation: 'predicate input any to unknown', caughtBy: 'protected predicate exact text guard', exit: predicateMutant.status});
fs.writeFileSync(target, original); assert.equal(apply().status, 0); assert.equal(fs.readFileSync(target, 'utf8'), adapted);
const second = apply(); assert.equal(second.status, 0); assert(second.stdout.includes('"sites":0')); assert.equal(fs.readFileSync(target, 'utf8'), adapted);
fs.writeFileSync(target, original.replaceAll('\r\n', '\n')); assert.equal(apply().status, 0); assert.equal(fs.readFileSync(target, 'utf8'), adapted.replaceAll('\r\n', '\n'));
const artifactMutants = [];
for (const suffix of ['.js', '/typescript.d.ts']) {
    const artifact = Object.keys(old).find(f => f.endsWith(suffix)); assert(artifact);
    const mutated = path.join(scratch, path.basename(artifact) + '.mutant');
    fs.writeFileSync(mutated, Buffer.concat([fs.readFileSync(path.join(after, artifact)), Buffer.from('\n')]));
    assert.throws(() => assert.equal(hash(fs.readFileSync(mutated)), old[artifact]), assert.AssertionError);
    artifactMutants.push({artifact, mutation: 'one real emitted artifact byte added', caughtBy: 'artifact SHA256 equality'});
}
assert.throws(() => assert.deepEqual({...newLane.counts, passing: newLane.counts.passing - 1}, oldLane.counts), assert.AssertionError);
const fixtures = [];
function runtime(tree) {
    const compiler = require(path.join(tree, 'built/local/typescript.js'));
    const results = [];
    const text = fs.readFileSync(path.join(tree, file), 'utf8'), parsed = functions(text);
    const converterText = parsed.result.convertOptionsFromJson.getText(parsed.source);
    const loadConverter = text => vm.runInNewContext(ts.transpileModule(text, {compilerOptions: options}).outputText + '\nconvertOptionsFromJson', {convertJsonOption: compiler.convertJsonOption});
    const converter = loadConverter(converterText);
    for (const raw of [undefined, null, false, 0, '']) {
        const defaults = {strict: true};
        assert.equal(converter(new Map(), raw, '/', defaults, undefined, []), undefined, 'falsy raw returns undefined even with defaults');
    }
    const defaults = {strict: true}, strict = compiler.optionDeclarations.find(o => o.name === 'strict');
    assert(strict);
    assert.equal(converter(new Map([['strict', strict]]), {strict: false}, '/', defaults, undefined, []), defaults);
    assert.equal(defaults.strict, false, 'normalization invalidates narrower input literals');
    const mutant = loadConverter(converterText.replace('return;', 'return defaultOptions;'));
    assert.throws(() => assert.equal(mutant(new Map(), false, '/', {strict: true}, undefined, []), undefined), assert.AssertionError);

    for (const raw of [undefined, null, false, 0, '', true, 7, 'raw', [], {strict: true, target: 'es2022'}, {strict: 'bad', lib: [false, {}, 'es2022'], unknownOption: 1}, {plugins: [{name: 'p'}, false, 7]}, {rootDir: './src'}, {customConditions: ['a', false, '']}]) {
        results.push(compiler.convertCompilerOptionsFromJson(raw, '/fixture', 'tsconfig.json'));
        results.push(compiler.convertTypeAcquisitionFromJson(raw, '/fixture', 'jsconfig.json'));
    }
    let reads = 0; const accessor = {get strict() { reads++; return true; }};
    results.push(compiler.convertCompilerOptionsFromJson(accessor, '/fixture')); assert.equal(reads, 1);
    results.push(compiler.convertCompilerOptionsFromJson(Object.create({strict: true}), '/fixture'));
    const cyclic = {}; cyclic.paths = cyclic; const cyclicResult = compiler.convertCompilerOptionsFromJson(cyclic, '/fixture'); assert.equal(cyclicResult.options.paths, cyclic); results.push({cyclicAccepted: true, errors: cyclicResult.errors});
    const host = {useCaseSensitiveFileNames: true, readDirectory: () => [], fileExists: () => false, readFile: () => undefined};
    for (const save of [undefined, null, false, true, 7, 'bad', {}, () => 1]) {
        const raw = {files: [], compileOnSave: save, watchOptions: {watchFile: 'usefsevents', excludeFiles: [7, './generated.ts'], unknownOption: true}, typeAcquisition: {include: ['node', false]}};
        const parsed = compiler.parseJsonConfigFileContent(raw, host, '/fixture', undefined, 'tsconfig.json');
        assert.equal(typeof raw.compileOnSave, 'boolean'); assert.equal(raw.compileOnSave, save === true);
        results.push({options: parsed.options, watchOptions: parsed.watchOptions, typeAcquisition: parsed.typeAcquisition, errors: parsed.errors, compileOnSave: raw.compileOnSave});
    }
    return JSON.stringify(results);
}
const oldRuntime = runtime(before), newRuntime = runtime(after); assert.equal(newRuntime, oldRuntime, 'actual API entrypoints through internal converters');
fs.writeFileSync(path.join(output, 'runtime.json'), oldRuntime + '\n');
const report = {base: oldLane.execution.commit, node: process.version, typescript: ts.version, reviewedSites: rules.map(r => r.id), sourceFiles: sourceFiles.length, sourceJavaScriptIdentical: true, idempotent: true, LFReconstruction: true, predicateExact: true, runtime: {fixtures: 'falsy and primitive raw inputs, bad list elements, normalized options, defaults, inherited keys, one-read accessor, cyclic object, compileOnSave mutation, watch/typeAcquisition diagnostics', sha256: hash(oldRuntime)}, outputs: current, jsFiles: Object.keys(current).filter(f => /\.(js|mjs|cjs)$/.test(f)).length, publicAPIFiles: Object.keys(current).filter(f => /typescript\.d\.(ts|mts|cts)$/.test(f)), lane: {status: newLane.status, counts: newLane.counts, failed_tests: newLane.failed_tests, baseline_diffs: newLane.baseline_diffs}, mutants, artifactMutants, laneCountMutant: 'caught', runtimeMutant: {site: 'convertOptionsFromJson', mutation: 'return defaults for falsy input', caughtBy: 'actual extracted converter return equality'}};
fs.writeFileSync(path.join(output, 'proof.json'), JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify({sites: rules.length, jsFiles: report.jsFiles, publicAPIFiles: report.publicAPIFiles, mutants: mutants.length, lane: report.lane}));
