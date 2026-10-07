import { createWASI, validateImports } from './wasi.mjs';

// Each host belongs to one isolate. Calls are synchronous, including cleanup.
export function createHost(module, { logger = console } = {}) {
  const encoder = new TextEncoder();
  // Preserve a leading BOM in the response, as the UTF-8 bytes require.
  const decoder = new TextDecoder('utf-8', { ignoreBOM: true });
  let instance;
  let busy = false;
  function initialize() {
    let next;
    const wasi = createWASI(() => next.exports.memory, logger);
    validateImports(module, wasi.imports);
    next = new WebAssembly.Instance(module, { wasi_snapshot_preview1: wasi.imports });
    for (const name of ['_initialize', 'malloc', 'free', 'adamic_request',
      'adamic_response_bytes', 'adamic_response_length', 'adamic_release']) {
      if (typeof next.exports[name] !== 'function') throw new Error(`Missing Wasm export: ${name}`);
    }
    if (!(next.exports.memory instanceof WebAssembly.Memory)) throw new Error('Missing Wasm export: memory');
    next.exports._initialize();
    instance = next;
  }
  return {
    call(text) {
      if (typeof text !== 'string') throw new TypeError('Wasm host.call requires a string');
      if (busy) throw new Error('Concurrent Wasm host.call');
      busy = true;
      try {
        if (!instance) initialize();
        const api = instance.exports;
        const bytes = encoder.encode(text);
        const input = api.malloc(Math.max(1, bytes.length)) >>> 0;
        if (input === 0) throw new Error('Wasm malloc failed');
        new Uint8Array(api.memory.buffer, input, bytes.length).set(bytes);
        const response = api.adamic_request(input, bytes.length) >>> 0;
        if (response === 0) throw new Error('Wasm handler returned a null string');
        const pointer = api.adamic_response_bytes(response) >>> 0;
        const length = api.adamic_response_length(response) >>> 0;
        const copy = new Uint8Array(api.memory.buffer, pointer, length).slice();
        api.adamic_release(response);
        api.free(input);
        return decoder.decode(copy);
      } catch (error) {
        // Never run cleanup or another request inside terminated execution.
        instance = undefined;
        throw error;
      } finally {
        busy = false;
      }
    },
    // Read-only diagnostics for counted integration builds; no initialization side effects.
    get exports() { return instance?.exports; },
  };
}
