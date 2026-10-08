// One fresh process per observation; disk reads and import setup precede Wasm timers.
import { readFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';
import { performance } from 'node:perf_hooks';
const [variant, artifact, oracle, corpus, wasiPath] = process.argv.slice(2);
if (variant === 'wasm') {
  const bytes = readFileSync(artifact);
  const { createWASI, validateImports } = await import(pathToFileURL(wasiPath));
  let instance;
  const wasi = createWASI(() => instance.exports.memory);
  const begin = performance.now();
  const module = await WebAssembly.compile(bytes);
  const compileMs = performance.now() - begin;
  validateImports(module, wasi.imports);
  const instantiate = performance.now();
  instance = new WebAssembly.Instance(module, { wasi_snapshot_preview1: wasi.imports });
  instance.exports._initialize();
  console.log(JSON.stringify({ compileMs, instantiateInitMs: performance.now() - instantiate }));
} else {
  // The unchanged oracle installs its hooks and loads the K=0 driver and modules.
  process.argv = [process.execPath, oracle, artifact, corpus, '0'];
  const original = console.log;
  const output = [];
  console.log = (...args) => output.push(args.join(' '));
  const begin = performance.now();
  await import(pathToFileURL(oracle));
  const parseModuleLoadMs = performance.now() - begin;
  console.log = original;
  if (output.join('\n') !== '0:0:2166136261:333555777') throw new Error('startup probe ran requests');
  console.log(JSON.stringify({ parseModuleLoadMs }));
}
