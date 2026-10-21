import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { registerHooks } from 'node:module';
import { pathToFileURL } from 'node:url';

const [directory, referencePath, generatedPath = `${directory}/crossing.mjs`, mode = 'all'] = process.argv.slice(2);
registerHooks({ resolve(specifier, context, next) {
  if (specifier === 'node:wasi' || specifier === 'wasi') throw new Error('Node WASI forbidden in crossing harness');
  return next(specifier, context);
} });
const reference = await import(pathToFileURL(referencePath));
const { createWASI, WASIExit } = await import(new URL('./wasi.mjs', pathToFileURL(generatedPath)));
const { encodeWTF8, decodeWTF8 } = await import(new URL('./wtf8.mjs', pathToFileURL(generatedPath)));
const table = JSON.parse(readFileSync(`${directory}/abi.json`, 'utf8'));
const module = new WebAssembly.Module(readFileSync(`${directory}/crossing.wasm`));
const Instance = WebAssembly.Instance;
const quiet = { log() {}, error() {} };
const generatedInstances = [];
let expectedWire;

// Observe actual allocation/free calls and the wire bytes independently of either host.
function instrument(real) {
  const api = { ...real.exports };
  const state = { initializations: 0, real, api, inputs: [], output: undefined, owned: new Set(), frees: 0, allocations: 0 };
  api._initialize = () => { state.initializations++; return real.exports._initialize(); };
  api.adamic_alloc = size => {
    const pointer = real.exports.adamic_alloc(size) >>> 0;
    assert.ok(!state.owned.has(pointer), 'allocator returned a live input pointer');
    state.owned.add(pointer); state.allocations++;
    return pointer;
  };
  api.adamic_free = pointer => {
    assert.ok(state.owned.delete(pointer >>> 0), 'input freed twice');
    state.frees++;
    return real.exports.adamic_free(pointer);
  };
  api.adamic_result_length = handle => {
    const length = real.exports.adamic_result_length(handle) >>> 0;
    const pointer = real.exports.adamic_result_bytes(handle) >>> 0;
    state.output = new Uint8Array(api.memory.buffer, pointer, length).slice();
    return length;
  };
  for (const signature of table.exports) {
    const name = `adamic_export_${signature.name}`;
    api[name] = (...args) => {
      state.inputs = []; state.output = undefined;
      let cursor = 0;
      for (const parameter of signature.parameters) {
        if (parameter.type.kind === 'number' || parameter.type.kind === 'boolean') { cursor++; continue; }
        const pointer = args[cursor++] >>> 0, length = args[cursor++] >>> 0;
        if (parameter.type.kind === 'number[]') assert.equal(pointer % 8, 0, 'unaligned direct f64 buffer');
        state.inputs.push(new Uint8Array(api.memory.buffer, pointer, parameter.type.kind === 'number[]' ? length * 8 : length).slice());
      }
      if (state.generated && expectedWire) assert.deepEqual(state.inputs, expectedWire, `${signature.name} input bytes differ`);
      return real.exports[name](...args);
    };
  }
  return state;
}
WebAssembly.Instance = function(module, imports) {
  const state = instrument(new Instance(module, imports));
  state.generated = true;
  generatedInstances.push(state);
  return { exports: state.api };
};
const { createCrossing } = await import(pathToFileURL(generatedPath));
const errors = [];
const generated = createCrossing(module, { logger: { log() {}, error(line) {
  errors.push(line);
  if (line === 'adamic: panic: crossing panic') assert.throws(() => generated.stringValue('nested'), /Concurrent Wasm crossing/);
} } });
let refInstance;
const refWASI = createWASI(() => refInstance.exports.memory, quiet);
refInstance = new Instance(module, { wasi_snapshot_preview1: refWASI.imports });
refInstance.exports._initialize();
const refState = instrument(refInstance);
const expected = reference.wrapExports({ exports: refState.api }, table);
const active = () => generatedInstances.at(-1);
function compare(name, args) {
  const oracle = expected[name](...args);
  expectedWire = refState.inputs;
  let actual;
  try { actual = generated[name](...args); } finally { expectedWire = undefined; }
  assert.deepEqual(actual, oracle, `${name} result differs`);
  assert.deepEqual(active().inputs, refState.inputs, `${name} input bytes differ`);
  assert.deepEqual(active().output, refState.output, `${name} result bytes differ`);
  assert.equal(active().owned.size, 0, 'input allocation leaked');
  return actual;
}
let seed = 0x61b5a793;
function random() { seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0; return seed; }
const floats = [-0, 0, NaN, Infinity, -Infinity, 5e-324, -5e-324, Number.MAX_VALUE, Math.PI];
const texts = ['', '\ud800', '\udfff', '\ud800\udfff', '\udfff\ud800', '世界🌍', '\0', '\ufeff', 'a\ud800z'];
function text() {
  const fixed = texts[random() % texts.length];
  return random() % 3 === 0 ? fixed + String.fromCharCode(random() & 0xffff, random() & 0xffff) : fixed;
}
function value(type) {
  switch (type.kind) {
    case 'number': return random() % 2 ? floats[random() % floats.length] : (random() - 0x80000000) / 17;
    case 'boolean': return (random() & 1) === 1;
    case 'string': return text();
    case 'number[]': return Array.from({ length: random() % 9 }, () => value({ kind: 'number' }));
    case 'string[]': return Array.from({ length: random() % 6 }, text);
    case 'record': {
      const record = {};
      // Construction order deliberately differs from schema order.
      for (const field of [...(type.fields ?? [])].reverse()) Object.defineProperty(record, field.name, { value: value(field.type), enumerable: true });
      return record;
    }
  }
}
if (mode === 'all' || mode === 'compare') {
  for (const signature of table.exports) {
    for (let index = 0; index < 10000; index++) {
      let args = signature.parameters.map(parameter => value(parameter.type));
      if (signature.name === 'grow') args = [text(), index % 7];
      compare(signature.name, args);
    }
    console.log(`${signature.name}: 10000 results and input/result byte comparisons passed`);
  }
  for (const text of texts) assert.equal(compare('stringValue', [text]), text);
  for (const number of floats) assert.ok(Object.is(compare('numberValue', [number]), number));
  for (let unit = 0; unit < 65536; unit++) {
    const text = String.fromCharCode(unit);
    assert.deepEqual(encodeWTF8(text), reference.encodeWTF8(text));
    assert.equal(decodeWTF8(reference.encodeWTF8(text)), text);
  }
  for (const bytes of [[0xc0, 0xaf], [0xe0, 0x80, 0x80], [0xf4, 0x90, 0x80, 0x80], [0xe2, 0x82], [0x80]]) {
    assert.throws(() => decodeWTF8(Uint8Array.from(bytes)), /WTF-8/);
  }
}
if (mode === 'all' || mode === 'growth') {
  // Both argument allocations and the handler itself detach earlier views.
  const big = Array.from({ length: 262144 }, (_, index) => index === 0 ? -0 : index + 5e-324);
  compare('mixed', ['small', big, 'last\ud800']);
  compare('numbers', [big]);
  const before = active().api.memory.buffer;
  assert.equal(compare('grow', ['世界\ud800', 1048576]), '世界\ud800'.repeat(1048576));
  assert.notEqual(active().api.memory.buffer, before, 'large result must grow memory during the call');
  console.log('growth: large inputs and handler allocation passed');
}
if (mode === 'all' || mode === 'lifetime') {
  const text = 'lifetime\ud800'.repeat(128);
  for (let index = 0; index < 1000; index++) generated.stringValue(text);
  const state = active();
  assert.equal(state.api.adamic_live(), 0);
  const baseline = state.api.memory.buffer.byteLength;
  for (let index = 0; index < 100000; index++) {
    assert.equal(generated.stringValue(text), text);
    assert.equal(active(), state, 'lifetime probe changed instances');
    assert.equal(state.api.adamic_live(), 0, 'live counted value leaked');
    assert.equal(state.api.memory.buffer.byteLength, baseline, 'memory grew after warmup');
    assert.equal(state.owned.size, 0, 'input allocation leaked');
  }
  assert.equal(state.allocations, state.frees, 'input free count mismatch');
  assert.equal(state.initializations, 1, 'initialization replayed');
  console.log(`lifetime: 100000 calls, live=0, memory=${baseline}, every input freed once`);
}
if (mode === 'all' || mode === 'panic') {
  generated.stringValue('before');
  const previous = active();
  const freeCount = previous.frees;
  assert.throws(() => generated.nothing('panic'), error => error instanceof WASIExit && error.code === 70);
  assert.ok(errors.includes('adamic: panic: crossing panic'), 'panic stderr lost');
  assert.equal(previous.frees, freeCount, 'terminated instance ran cleanup');
  assert.equal(generated.stringValue('after\ud800'), 'after\ud800');
  assert.notEqual(active(), previous, 'terminated crossing instance reused');
  assert.equal(active().api.adamic_live(), 0);
  console.log('panic: stderr, no terminated cleanup, fresh instance passed');
}
if (mode === 'all') {
  assert.throws(() => generated.stringValue(), /argument count/);
  const source = readFileSync(generatedPath, 'utf8');
  assert.doesNotMatch(source, /node:|\beval\s*\(|new\s+Function|\.fields|\.parameters|\.kind/);
  for (const name of ['wtf8.mjs', 'crossing-runtime.mjs', 'wasi.mjs']) {
    assert.doesNotMatch(readFileSync(new URL(name, pathToFileURL(generatedPath)), 'utf8'), /node:|\beval\s*\(|new\s+Function/);
  }
  const trapModule = new WebAssembly.Module(readFileSync(`${directory}/trap.wasm`));
  const trapFactory = (await import(pathToFileURL(`${directory}/trap-crossing.mjs`))).createCrossing;
  const trapped = trapFactory(trapModule, { logger: quiet });
  assert.equal(trapped.numberValue(0), 1);
  const previous = active();
  assert.throws(() => trapped.numberValue(-1), WebAssembly.RuntimeError);
  assert.equal(trapped.numberValue(0), 1);
  assert.notEqual(active(), previous, 'trapped instance reused');
  assert.equal(active().initializations, 1);
  console.log('trap recovery, reentrancy, initialization and generated code restrictions passed');
}
