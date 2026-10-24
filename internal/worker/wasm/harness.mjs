import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { registerHooks } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';

const [directory, mode = 'all'] = process.argv.slice(2);
// Any accidental dependency on Node WASI, including a transitive one, fails.
registerHooks({ resolve(specifier, context, next) {
  if (specifier === 'node:wasi' || specifier === 'wasi') throw new Error('Node WASI forbidden in host harness');
  return next(specifier, context);
}, load(url, context, next) {
  if (url.endsWith('.wasm')) {
    const bytes = readFileSync(fileURLToPath(url));
    return { format: 'module', shortCircuit: true,
      source: `export default new WebAssembly.Module(Uint8Array.from(${JSON.stringify([...bytes])}));` };
  }
  return next(url, context);
} });
const { createHost } = await import('./host.mjs');
const { createWASI, WASIExit, validateImports } = await import('./wasi.mjs');
const { createWorker } = await import('./bridge.mjs');
const load = name => new WebAssembly.Module(readFileSync(`${directory}/${name}.wasm`));
const quiet = { log() {}, error() {} };
function imports(module, fdstat = false) {
  const entries = WebAssembly.Module.imports(module);
  console.log(JSON.stringify(entries));
  assert.deepEqual(entries.map(entry => `${entry.module}.${entry.name}`).sort(),
    ['fd_close', 'fd_prestat_get', 'fd_prestat_dir_name', 'fd_seek', 'fd_write', 'proc_exit', ...(fdstat ? ['fd_fdstat_get'] : [])]
      .map(name => `wasi_snapshot_preview1.${name}`).sort());
}

if (mode === 'all' || mode === 'compare') {
  const host = createHost(load('driver'), { logger: quiet });
  imports(load('driver'), true);
  const requests = JSON.parse(readFileSync(`${directory}/requests.json`, 'utf8'));
  const expected = JSON.parse(readFileSync(`${directory}/expected.json`, 'utf8'));
  assert.equal(expected.length, 1000);
  for (let index = 0; index < 1000; index++) {
    assert.deepEqual(new TextEncoder().encode(host.call(requests[index])),
      new TextEncoder().encode(expected[index]), `driver response ${index}`);
  }
  const grow = createHost(load('grow'), { logger: quiet });
  imports(load('grow'));
  assert.equal(grow.call('世界🌍'), JSON.parse(readFileSync(`${directory}/grow-expected.json`, 'utf8'))[0], 'large response after memory growth');
  assert.ok(grow.exports.memory.buffer.byteLength > 65536 * 10);
  console.log('compare: 1000 byte-identical responses and growing response passed');
}
if (mode === 'all' || mode === 'lifetime') {
  const host = createHost(load('counted'), { logger: quiet });
  const requests = ['', 'request', '世界 🌍', 'x'.repeat(1024), 'a\0b'];
  for (let index = 0; index < 1000; index++) host.call(requests[index % requests.length]);
  const api = host.exports;
  assert.equal(api.adamic_live(), 0, 'warmup retained counted values');
  const memory = api.memory.buffer.byteLength;
  const regions = api.adamic_regions();
  for (let index = 0; index < 100000; index++) {
    const text = requests[index % requests.length];
    assert.equal(host.call(text), `hello ${text}: 63`);
    assert.equal(host.exports, api, 'lifetime run must keep one instance');
    assert.equal(api.adamic_live(), 0, `live allocation at call ${index}`);
    assert.equal(api.memory.buffer.byteLength, memory, `memory grew at call ${index}`);
  }
  assert.equal(api.adamic_regions() - regions, 6300000);
  console.log(`lifetime: 100000 calls, live=0, memory=${memory}, regionObjects=6300000`);
}
if (mode === 'all' || mode === 'panic') {
  const errors = [];
  const host = createHost(load('panic'), { logger: { log() {}, error(line) { errors.push(line); } } });
  imports(load('panic'));
  const worker = createWorker(host);
  const request = body => new Request('https://example.test/', { method: 'POST', body });
  const bad = await worker.fetch(request('explode'));
  assert.equal(bad.status, 500);
  assert.equal(await bad.text(), 'internal error');
  assert.deepEqual(errors, ['adamic: panic: host fixture exploded'], 'panic stderr lost');
  const good = await worker.fetch(request('good'));
  assert.equal(good.status, 200);
  assert.equal(await good.text(), 'fresh:1', 'terminated instance reused');
  const api = host.exports;
  assert.equal(await (await worker.fetch(request('again'))).text(), 'fresh:2');
  assert.equal(host.exports, api, 'good requests should share the initialized instance');
  console.log('panic: stderr preserved, 500 then fresh:1 passed');
}
if (mode === 'all') {
  const trapped = createHost(load('trap'));
  assert.equal(trapped.call('ok'), '1');
  const previous = trapped.exports;
  assert.throws(() => trapped.call('trap'), WebAssembly.RuntimeError);
  assert.equal(trapped.exports, undefined);
  assert.equal(trapped.call('ok'), '1', 'trap must recreate the instance');
  assert.notEqual(trapped.exports, previous);
  const initialization = [];
  let reentrant;
  reentrant = createHost(load('driver'), { logger: { error() {}, log(line) {
    initialization.push(line);
    assert.throws(() => reentrant.call('nested'), /Concurrent Wasm host.call/);
  } } });
  reentrant.call('first'); reentrant.call('second');
  assert.deepEqual(initialization, ['dependency initialized', 'entry initialized']);
  // Test the actual generated module with a Node loader substituting its Wasm binding.
  const generated = (await import(pathToFileURL(`${directory}/generated/worker.mjs`))).default;
  const bad = await generated.fetch(new Request('https://example.test/', { method: 'POST', body: 'explode' }));
  assert.equal(bad.status, 500);
  assert.equal(await bad.text(), 'internal error');
  assert.equal(await (await generated.fetch(new Request('https://example.test/'))).text(), 'fresh:1');
  let captured;
  const bridge = createWorker({ call(text) {
    captured = JSON.parse(text);
    return JSON.stringify({ status: 201, headers: [{ name: 'x-answer', value: 'yes' }], body: '世界' });
  } });
  const request = new Request('https://example.test/a%2Fb?k=one+two&k=%E4%B8%96', {
    method: 'POST', headers: [['Z-Last', 'z'], ['A-First', 'a']], body: 'body 🌍',
  });
  const response = await bridge.fetch(request);
  assert.deepEqual(captured, { method: 'POST', url: request.url, path: '/a%2Fb',
    query: [{ name: 'k', value: 'one two' }, { name: 'k', value: '世' }],
    headers: Array.from(request.headers, ([name, value]) => ({ name, value })), body: 'body 🌍' });
  assert.equal(response.status, 201);
  assert.equal(response.headers.get('x-answer'), 'yes');
  assert.equal(await response.text(), '世界');
  assert.match(readFileSync(`${directory}/generated/wrangler.toml`, 'utf8'), /type = "CompiledWasm"/);
  for (const name of ['worker', 'bridge', 'host', 'wasi']) {
    assert.doesNotMatch(readFileSync(`${directory}/generated/${name}.mjs`, 'utf8'), /node:|\beval\s*\(|new\s+Function/);
  }
  // Split UTF-8 and lines, invalid descriptors, partial-line flush, explicit exit code.
  const memory = new WebAssembly.Memory({ initial: 1 });
  const lines = [], errors = [];
  const wasi = createWASI(() => memory, { log: line => lines.push(line), error: line => errors.push(line) });
  const write = (fd, bytes) => {
    new Uint8Array(memory.buffer, 64, bytes.length).set(bytes);
    const view = new DataView(memory.buffer);
    view.setUint32(0, 64, true); view.setUint32(4, bytes.length, true);
    assert.equal(wasi.imports.fd_write(fd, 0, 1, 16), 0);
    assert.equal(view.getUint32(16, true), bytes.length);
  };
  write(1, Uint8Array.of(0xe4)); write(1, Uint8Array.of(0xb8, 0x96, 10, 97, 10));
  write(2, new TextEncoder().encode('partial'));
  assert.throws(() => wasi.imports.proc_exit(70), error => error instanceof WASIExit && error.code === 70);
  assert.deepEqual(lines, ['世', 'a']); assert.deepEqual(errors, ['partial']);
  assert.equal(wasi.imports.fd_write(3, 0, 1, 16), 8);
  assert.equal(wasi.imports.fd_prestat_get(3, 0), 8);
  assert.equal(wasi.imports.fd_fdstat_get(1, 32), 0);
  assert.equal(new DataView(memory.buffer).getUint8(32), 2);
  assert.equal(new DataView(memory.buffer).getBigUint64(40, true), 64n);
  assert.equal(wasi.imports.fd_fdstat_get(3, 32), 8);
  // Tiny real module importing an unknown WASI function.
  const unknown = new WebAssembly.Module(Uint8Array.of(0,97,115,109,1,0,0,0,
    1,4,1,96,0,0,2,10,1,3,101,110,118,2,110,111,0,0));
  assert.throws(() => validateImports(unknown, wasi.imports), /env.no/);
  console.log('bridge, generated Worker, import rejection and WASI logging passed');
}
