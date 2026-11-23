'use strict';

const assert = require('node:assert/strict');
const {execFileSync, spawnSync} = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {performance} = require('node:perf_hooks');
const test = require('node:test');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
const applied = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-seats-build22-101-jsx-option-map';
const packageFile = require.resolve('typescript/package.json');
assert.equal(require(packageFile).version, '6.0.3');
const typescript = path.dirname(packageFile);
const adapter = path.join(__dirname, 'adapt.cjs');
const parser = 'src/compiler/commandLineParser.ts';
const roots = [];

function cleanup() {
    for (const root of roots) fs.rmSync(root, {recursive: true, force: true});
}

function snapshot(side, file) {
    return execFileSync('git', ['-C', applied, 'show', ref + '-' + side + ':' + file],
        {maxBuffer: 8 * 1024 * 1024});
}

function logTime(phase, started) {
    console.log(phase + ' wall ms ' + (performance.now() - started).toFixed(1));
}

function extractBefore() {
    const started = performance.now();
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'jsx-table-before-'));
    roots.push(root);
    const archive = path.join(root, 'before.tar');
    const input = path.join(root, 'tree');
    fs.mkdirSync(input);
    execFileSync('git', ['-C', applied, 'archive', '--format=tar', '-o', archive, ref + '-before']);
    execFileSync('tar', ['-x', '-f', archive, '-C', input]);
    logTime('Extract before', started);
    return input;
}

function copyTree(input, label) {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'jsx-table-' + label + '-'));
    roots.push(root);
    fs.cpSync(path.join(input, 'src'), path.join(root, 'src'), {recursive: true});
    fs.copyFileSync(path.join(input, 'package.json'), path.join(root, 'package.json'));
    fs.symlinkSync(path.dirname(typescript), path.join(root, 'node_modules'), 'junction');
    return root;
}

function run(label, root, args) {
    const started = performance.now();
    const result = spawnSync(process.execPath, args, {
        cwd: root,
        encoding: 'utf8',
        env: {...process.env, NODE_DISABLE_COMPILE_CACHE: '1'},
    });
    if (result.error) throw result.error;
    process.stdout.write(result.stdout);
    process.stderr.write(result.stderr);
    assert.equal(result.status, 0, label + ' exit status');
    console.log(label + ' exit 0');
    logTime(label, started);
    return result.stdout.trim();
}

function tables(root) {
    const compiler = require(path.join(root, 'built/local/compiler/_namespaces/ts.js'));
    const descriptor = compiler.optionDeclarations.find(option => option.name === 'jsx');
    return {forward: descriptor.type, inverse: compiler.inverseJsxOptionMap, compiler};
}

function buildTables() {
    try {
        const input = extractBefore();
        const stock = copyTree(input, 'stock');
        const changed = copyTree(input, 'changed');
        const original = fs.readFileSync(path.join(input, parser));
        assert.equal(run('Apply', changed, [adapter, changed]), JSON.stringify({file: parser, edits: 2}));
        const changedParser = path.join(changed, parser);
        const first = fs.readFileSync(changedParser);
        assert.equal(run('Replay', changed, [adapter, changed]), JSON.stringify({file: parser, edits: 0}));
        assert.deepEqual(fs.readFileSync(changedParser), first);
        assert.deepEqual(fs.readFileSync(path.join(stock, parser)), original);

        const build = [path.join(typescript, 'bin', 'tsc'), '-b', 'src/typescript', 'src/tsc',
            '--emitDeclarationOnly', 'false', '--pretty', 'false'];
        run('Stock emit stock', stock, build);
        run('Stock emit changed', changed, build);
        return [tables(stock), tables(changed), first, fs.readFileSync(path.join(stock, parser))];
    } catch (error) {
        cleanup();
        throw error;
    }
}

const [stock, changed, first, stockParser] = buildTables();
const casesStarted = performance.now();
test.after(() => {
    logTime('Cases', casesStarted);
    cleanup();
});

test("adapter output equals the pipeline's after snapshot", () => {
    const after = snapshot('after', parser);
    const before = snapshot('before', parser);
    assert.ok(first.equals(after), `fresh adapter output (${first.length} bytes) differs from after snapshot (${after.length} bytes)`);
    assert.ok(stockParser.equals(before), `stock parser (${stockParser.length} bytes) differs from before snapshot (${before.length} bytes)`);
});

function inspect(map, keys) {
    const callbacks = [];
    map.forEach((value, key, source) => {
        assert.strictEqual(source, map);
        callbacks.push([key, typeof value, value]);
    });
    return {
        size: map.size,
        keys: [...map.keys()],
        values: [...map.values()],
        entries: [...map.entries()],
        iteration: [...map],
        callbacks,
        gets: keys.map(key => [key, map.has(key), typeof map.get(key), map.get(key)]),
    };
}

const expected = {
    forward: [['preserve',1],['react-native',3],['react-jsx',4],['react-jsxdev',5],['react',2]],
    inverse: [['1','preserve'],['3','react-native'],['4','react-jsx'],['5','react-jsxdev'],['2','react']],
};
for (const [name, entries] of Object.entries(expected)) {
    test(name + ' keys, tagged values and insertion order', () => {
        const keys = entries.map(([key]) => key).concat('missing');
        assert.deepEqual([...stock[name]], entries);
        const witness = inspect(stock[name], keys);
        const observed = inspect(changed[name], keys);
        console.log(name + ' stock ' + JSON.stringify(witness));
        console.log(name + ' tested ' + JSON.stringify(observed));
        assert.deepEqual(observed.entries, witness.entries);
        assert.deepEqual(observed, witness);
    });
}
test('CLI and JSON use the same option values and invalid help order', () => {
    for (const key of [...expected.forward.map(([key]) => key), 'not-a-mode']) {
        const cli = ts => ts.parseCommandLine(['--jsx', key]);
        const config = ts => ts.convertCompilerOptionsFromJson({jsx: key}, '/');
        for (const read of [cli, config]) {
            const brief = ({options, errors}) => [options.jsx,
                errors.map(({code, messageText}) => [code, messageText])];
            const before = brief(read(stock.compiler));
            const after = brief(read(changed.compiler));
            console.log(key, JSON.stringify(before), JSON.stringify(after));
            assert.deepEqual(after, before);
        }
    }
});
