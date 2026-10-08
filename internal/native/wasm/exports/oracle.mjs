import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import { WASI } from 'node:wasi';
import { wrapExports } from '../exports.mjs';
const [binary, sourcePath, tablePath] = process.argv.slice(2);
const source = stripTypeScriptTypes(readFileSync(sourcePath, 'utf8'));
const node = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
const table = JSON.parse(readFileSync(tablePath, 'utf8'));
const module = await WebAssembly.compile(readFileSync(binary));
const wasi = new WASI({version: 'preview1', args: [], env: {}, preopens: {}});
const instance = await WebAssembly.instantiate(module, {wasi_snapshot_preview1: wasi.wasiImport});
wasi.initialize(instance);
const wasm = wrapExports(instance, table);
const cases = [];
const add = (name, values) => { for (const value of values) cases.push([name, [value]]); };
add('numberValue', [0, -0, NaN, Infinity, -Infinity, 1.25, Number.MIN_VALUE]);
add('booleanValue', [false, true]);
add('stringValue', ['', 'hello', '🦀漢字', '\ud800', '\udfff', 'x\ud800z', 'nul\0byte', '🦀'.repeat(16000)]);
add('numbers', [[], [0, -0, NaN, Infinity, -Infinity, 1.25], Array.from({length: 16000}, (_, i) => i / 7)]);
add('strings', [[], ['', '🦀', '\ud800', '\udfff'], Array.from({length: 4000}, (_, i) => `item${i}🦀`)]);
add('recordValue', [
 {title: '', values: [], inner: {enabled: false, score: -0, words: []}},
 {title: '🦀\ud800', values: [NaN, -0, Infinity], inner: {enabled: true, score: NaN, words: ['', '\udfff']}},
 {title: 'large'.repeat(3000), values: Array.from({length: 4000}, (_, i) => i), inner: {enabled: false, score: 42, words: ['large'.repeat(4000)]}},
]);
add('noResult', ['', '🦀\ud800']);
add('heavy', [-0, 42]);
add('handleRequest', ['', '🦀\ud800']);
add('emptyRecord', [{}]);
add('bumpRecord', [{title: '🦀', values: [-0, NaN], inner: {enabled: true, score: -0, words: ['\\ud800']}}]);
cases.push(['counter', []]);
cases.push(['many', ['first'.repeat(16000), [0, -0, NaN], 'second'.repeat(16000), true, -0]]);
cases.push(['surrogate', []]);
function check(name, arguments_) {
 const expected = node[name](...arguments_);
 const actual = wasm[name](...arguments_);
 assert.deepStrictEqual(actual, expected, name);
 assert.equal(instance.exports.adamic_live(), 0, `live after ${name}`);
}
for (const [name, arguments_] of cases) check(name, arguments_);
// Warm allocator bins with every large case before measuring reusable capacity.
for (let i = 0; i < 100; i++) for (const [name, arguments_] of cases) check(name, arguments_);
const pages = instance.exports.memory.buffer.byteLength;
const regions = instance.exports.adamic_regions();
// Large values are covered above. Mix short calls to keep the lifetime sweep fast.
const mixed = cases.filter(([, arguments_]) => JSON.stringify(arguments_)?.length < 1000);
for (let i = 0; i < 100000; i++) {
 const [name, arguments_] = i % 256 === 0 ? cases.find(([name, args]) => name === "numbers" && args[0].length === 16000) : mixed[i % mixed.length];
 check(name, arguments_);
 assert.equal(instance.exports.memory.buffer.byteLength, pages, `memory at call ${i}`);
}
assert.ok(instance.exports.adamic_regions() > regions, 'heavy fixture must exercise regions');
console.log(JSON.stringify({calls: 100000, cases: cases.length, memory: pages, live: instance.exports.adamic_live(), regions: instance.exports.adamic_regions(), imports: WebAssembly.Module.imports(module)}));
