'use strict';
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const { adapt, inspect, relativeFile } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
const applied = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-seats-build22-136-level-overrides';

function snapshot(side) {
    return execFileSync('git', ['-C', applied, 'show', ref + '-' + side + ':' + relativeFile], { maxBuffer: 8 * 1024 * 1024 });
}

const [beforeBytes, afterBytes] = ['before', 'after'].map(snapshot);
const baseline = beforeBytes.toString('utf8');
const original = inspect(baseline);
assert.equal(original.adapted, false);
const recorded = inspect(afterBytes.toString('utf8'));
assert(recorded.adapted);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-level-overrides-'));
const file = path.join(scratch, relativeFile);
try {
    fs.mkdirSync(path.dirname(file), { recursive: true });
    const store = '[level] = Number(customLevel);';
    const producer = original.producer.getText();
    assert.equal(producer.split(store).length, 2);
    for (const value of ['customLevel', 'undefined']) {
        const mutant = baseline.replace(producer, producer.replace(store, `[level] = ${value};`));
        assert.notEqual(mutant, baseline);
        const bytes = Buffer.from(mutant);
        fs.writeFileSync(file, bytes);
        assert.throws(() => adapt(scratch), /getCustomLevels producer changed/);
        assert.deepEqual(fs.readFileSync(file), bytes);
        console.log(`Producer storing ${value} refused without writes`);
    }

    fs.writeFileSync(file, baseline);
    assert.equal(adapt(scratch).edits, 1);
    const adaptedBytes = fs.readFileSync(file);
    assert.equal(adapt(scratch).edits, 0);
    assert.deepEqual(fs.readFileSync(file), adaptedBytes);
    assert.deepEqual(adaptedBytes, afterBytes, 'scratch output differs from the recorded after snapshot');
    const adapted = inspect(adaptedBytes.toString('utf8'));
    assert(adapted.adapted);
    console.log('One merge edit; replay byte-identical');

    const javascript = ts.transpileModule(`
        ${producer}
        const defaultLevels = ${original.defaults.initializer.getText()};
        const getLevel = (_baseVariable: string, level: string) => env[level];
        const customLevels = getCustomLevels("TSC_WATCH_POLLINGCHUNKSIZE");
        return { customLevels, stock: ${original.conditional.getText()}, adapted: ${recorded.conditional.getText()} };
    `, { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText;
    const evaluate = new Function('env', javascript);
    const cases = [
        ['no variable', {}, undefined, [32, 64, 256]],
        ['zero', { Low: '0' }, ['Low'], [0, 64, 256]],
        ['NaN', { Low: 'abc' }, ['Low'], [NaN, 64, 256]],
        ['empty string', { Low: '' }, undefined, [32, 64, 256]],
        ['all three', { Low: '1', Medium: '2', High: '3' }, ['Low', 'Medium', 'High'], [1, 2, 3]],
    ];
    const keys = ['Low', 'Medium', 'High'];
    for (const [name, env, producedKeys, expected] of cases) {
        const { customLevels, stock, adapted } = evaluate(env);
        if (producedKeys) assert.deepEqual(Object.keys(customLevels), producedKeys, name);
        else assert.equal(customLevels, undefined, name);
        assert.deepEqual(Object.keys(stock), keys, name);
        assert.deepEqual(Object.keys(adapted), Object.keys(stock), name);
        for (const [index, key] of keys.entries()) {
            assert(Object.is(stock[key], expected[index]), `${name}: unexpected stock value at ${key}`);
            assert(Object.is(adapted[key], stock[key]), `${name}: different value at ${key}`);
        }
        console.log(`${name}: ${keys.map(key => `${key}=${String(adapted[key])}`).join(', ')}`);
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
