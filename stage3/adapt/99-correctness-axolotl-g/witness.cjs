'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const parser = require('typescript');
assert.equal(parser.version, '6.0.3');
function checkAfter(ts) {
    for (const option of ['outFile', 'strict', 'target', 'rootDir']) {
        const parsed = ts.parseCommandLine(['--init', '--' + option, 'null']);
        assert.equal(parsed.errors.length, 0);
        const config = ts.generateTSConfig(parsed.options, '\n');
        assert(config.includes('"' + option + '": null,'), `${option} did not retain null`);
        assert(!ts.parseConfigFileTextToJson('tsconfig.json', config).error, 'generated config has syntax errors');
    }
}
if (process.argv[2] === '--check-after') {
    checkAfter(require(path.resolve(process.argv[3])));
    process.exit(0);
}
assert.equal(process.argv.length, 6, 'usage: node witness.cjs <before library> <after library> <pinned Go CLI> <output>');
const [beforeFile, afterFile, goFile, out] = process.argv.slice(2).map(p => path.resolve(p));
fs.mkdirSync(out, {recursive: true});
const before = require(beforeFile), after = require(afterFile);
checkAfter(after);
const rows = [];
const cases = [
    {name: 'ordinary', args: ['--init'], unchanged: true},
    {name: 'strict-false', args: ['--init', '--strict', 'false'], unchanged: true},
    ...['outFile', 'strict', 'target', 'rootDir'].map(option => ({name: option + '-null', args: ['--init', '--' + option, 'null'], option})),
];
function cli(label, name, binary, args) {
    const cwd = path.join(out, name, label);
    fs.mkdirSync(cwd, {recursive: true});
    fs.writeFileSync(path.join(cwd, 'index.ts'), 'export {};\n');
    const run = cp.spawnSync(binary[0], [...binary.slice(1), ...args], {cwd, encoding: 'utf8'});
    fs.writeFileSync(path.join(cwd, 'stdout.log'), run.stdout || '');
    fs.writeFileSync(path.join(cwd, 'stderr.log'), run.stderr || '');
    return {cwd, exit: run.status};
}
function nodeCLI(file) { return [process.execPath, '-e', 'const ts=require(process.argv[1]); ts.executeCommandLine(ts.sys,()=>{},process.argv.slice(2));', file]; }
for (const item of cases) {
    const row = {name: item.name, args: item.args};
    const configs = {};
    for (const [label, library] of [['before', beforeFile], ['after', afterFile]]) {
        const run = cli(label, item.name, nodeCLI(library), item.args);
        assert.equal(run.exit, 0, `${item.name} ${label} init failed`);
        const config = fs.readFileSync(path.join(run.cwd, 'tsconfig.json'), 'utf8');
        configs[label] = config;
        const read = cp.spawnSync(process.execPath, nodeCLI(library).slice(1).concat('--showConfig'), {cwd: run.cwd, encoding: 'utf8'});
        fs.writeFileSync(path.join(run.cwd, 'readback.stdout.log'), read.stdout || '');
        fs.writeFileSync(path.join(run.cwd, 'readback.stderr.log'), read.stderr || '');
        row[label] = {initExit: run.exit, readbackExit: read.status, optionLines: item.option ? config.split('\n').filter(s => s.includes('"' + item.option + '"')) : [], readbackDiagnostics: (read.stdout || '').split('\n').filter(s => /error TS/.test(s))};
        if (label === 'after' || item.unchanged) assert.equal(read.status, 0, `${item.name} ${label} readback failed: ${read.stdout}`);
    }
    if (item.unchanged) assert.equal(configs.after, configs.before, `${item.name} changed`);
    if (item.name === 'outFile-null') {
        assert(configs.before.includes('"outFile": undefined,'));
        assert.notEqual(row.before.readbackExit, 0);
        assert(row.before.readbackDiagnostics.length > 0);
    }
    const go = cli('go', item.name, [goFile], item.args);
    row.go = {initExit: go.exit};
    if (go.exit === 0) {
        const config = fs.readFileSync(path.join(go.cwd, 'tsconfig.json'), 'utf8');
        row.go.byteIdentical = config === configs.after;
        const differences = [];
        const a = config.split('\n'), b = configs.after.split('\n');
        for (let i = 0; i < Math.max(a.length, b.length); i++) if (a[i] !== b[i]) differences.push({line: i + 1, go: a[i], after: b[i]});
        row.go.differences = differences;
        assert.equal(differences.length, 0, `${item.name}: Go output differs`);
    }
    else {
        assert.equal(item.name, 'target-null', 'unexpected Go failure');
        assert(fs.readFileSync(path.join(go.cwd, 'stderr.log'), 'utf8').includes('panic: No matching value of <nil>'));
        row.go.difference = 'pinned Go panics for cleared target and writes no config; 99g writes target: null';
    }
    rows.push(row);
}
// API controls also cover absent versus inherited properties, falsy defined
// values, and own null. Only own undefined values are changed by this patch.
for (const options of [{}, {strict:false}, {strict:null}, {outFile:''}, {maxNodeModuleJsDepth:0}, {types:[]}, Object.assign(Object.create({strict:undefined}), {}), Object.assign(Object.create({strict:false}), {})]) {
    assert.equal(after.generateTSConfig(options, '\n'), before.generateTSConfig(options, '\n'));
}
// Restore generateTSConfig's original compiled function in the actual after
// library. The same after checker must fail on that one-function reversion.
function generator(file) {
    const text = fs.readFileSync(file, 'utf8');
    const source = parser.createSourceFile(file, text, parser.ScriptTarget.Latest, true, parser.ScriptKind.JS);
    const found = [];
    (function visit(n) { if (parser.isFunctionDeclaration(n) && n.name?.text === 'generateTSConfig') found.push(n); parser.forEachChild(n, visit); })(source);
    assert.equal(found.length, 1);
    return {text, start: found[0].getStart(source), end: found[0].end, code: found[0].getText(source)};
}
const original = generator(beforeFile), repaired = generator(afterFile);
const mutantFile = path.join(out, 'mutant.cjs');
const mutant = repaired.text.slice(0, repaired.start) + original.code + repaired.text.slice(repaired.end);
fs.writeFileSync(mutantFile, mutant);
const red = cp.spawnSync(process.execPath, [__filename, '--check-after', mutantFile], {encoding:'utf8', env: process.env});
fs.writeFileSync(path.join(out, 'mutant.stdout.log'), red.stdout || '');
fs.writeFileSync(path.join(out, 'mutant.stderr.log'), red.stderr || '');
assert.notEqual(red.status, 0, 'old-line mutant did not turn after witness red');
assert((red.stderr || '').includes('outFile did not retain null'), red.stderr);
const result = {cases: rows, apiControlsUnchanged: true, mutant: {exit: red.status, failure: 'outFile did not retain null'}};
fs.writeFileSync(path.join(out, 'result.json'), JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result, null, 2));
