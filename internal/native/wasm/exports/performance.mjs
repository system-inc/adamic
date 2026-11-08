import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
import { instantiateExports } from '../exports.mjs';

const [binary, tablePath, mode] = process.argv.slice(2);
const module = await WebAssembly.compile(readFileSync(binary));
const table = JSON.parse(readFileSync(tablePath, 'utf8'));
const wasi = new WASI({ version: 'preview1', args: [], env: {}, preopens: {} });
const { functions } = await instantiateExports(
    module,
    table,
    { wasi_snapshot_preview1: wasi.wasiImport },
    mode === 'skipped' ? () => {} : (instance) => wasi.initialize(instance),
);

if(mode === 'skipped') {
    assert.throws(() => functions.now(), { code: 'ERR_WASI_NOT_STARTED' });
    console.log('first now(): ERR_WASI_NOT_STARTED');
}
else {
    const values = [functions.now(), functions.now(), functions.now()];
    for(const [index, value] of values.entries()) {
        assert.ok(
            Number.isFinite(value) && value >= 0 && value < 60000,
            `now[${index}]=${value} must be finite, >= 0 and < 60000 ms`,
        );
        if(index > 0) assert.ok(value >= values[index - 1], `now[${index}] went backwards`);
    }
    console.log(JSON.stringify({ now: values }));
}
