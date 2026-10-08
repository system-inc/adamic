import { createWASI, validateImports } from './wasi.mjs';
import { encodeWTF8, decodeWTF8 } from './wtf8.mjs';

// Primitive operations only. Generated functions supply every schema operation.
export class Encoder {
  constructor() { this.buffer = new Uint8Array(64); this.offset = 0; }
  reserve(length) {
    const needed = this.offset + length;
    if (!Number.isSafeInteger(needed) || needed > 0xffffffff) throw new RangeError('ABI buffer too large');
    if (needed > this.buffer.length) {
      const larger = new Uint8Array(Math.max(needed, this.buffer.length * 2));
      larger.set(this.buffer);
      this.buffer = larger;
    }
  }
  align() {
    const padding = (8 - this.offset % 8) % 8;
    this.reserve(padding);
    this.buffer.fill(0, this.offset, this.offset + padding);
    this.offset += padding;
  }
  number(value) {
    this.align(); this.reserve(8);
    new DataView(this.buffer.buffer).setFloat64(this.offset, value, true);
    this.offset += 8;
  }
  boolean(value) { this.reserve(1); this.buffer[this.offset++] = value ? 1 : 0; }
  count(value) {
    this.reserve(4);
    new DataView(this.buffer.buffer).setUint32(this.offset, value, true);
    this.offset += 4;
  }
  string(value) {
    const bytes = encodeWTF8(value);
    this.count(bytes.length); this.reserve(bytes.length);
    this.buffer.set(bytes, this.offset); this.offset += bytes.length;
  }
  finish() { return this.buffer.subarray(0, this.offset); }
}
export class Decoder {
  constructor(bytes) {
    this.bytes = bytes;
    this.view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
    this.offset = 0;
  }
  take(length) {
    if (length > this.bytes.length - this.offset) throw new Error('Truncated ABI result');
    const start = this.offset; this.offset += length; return start;
  }
  align() { this.take((8 - this.offset % 8) % 8); }
  number() { this.align(); return this.view.getFloat64(this.take(8), true); }
  boolean() {
    const byte = this.bytes[this.take(1)];
    if (byte > 1) throw new Error('Invalid ABI boolean');
    return byte === 1;
  }
  count() { return this.view.getUint32(this.take(4), true); }
  string() {
    const length = this.count(); const start = this.take(length);
    return decodeWTF8(this.bytes.subarray(start, start + length));
  }
  finish() { if (this.offset !== this.bytes.length) throw new Error('Trailing ABI result bytes'); }
}
export function directNumbers(values) {
  const bytes = new Uint8Array(values.length * 8);
  const view = new DataView(bytes.buffer);
  for (let index = 0; index < values.length; index++) view.setFloat64(index * 8, values[index], true);
  return bytes;
}

export function createLifecycle(module, names, logger = console) {
  let instance, busy = false;
  function initialize() {
    let next;
    const wasi = createWASI(() => next.exports.memory, logger);
    validateImports(module, wasi.imports);
    next = new WebAssembly.Instance(module, { wasi_snapshot_preview1: wasi.imports });
    for (const name of ['_initialize', 'adamic_alloc', 'adamic_free', 'adamic_result_bytes',
      'adamic_result_length', 'adamic_result_release', ...names]) {
      if (typeof next.exports[name] !== 'function') throw new Error(`Missing Wasm export: ${name}`);
    }
    if (!(next.exports.memory instanceof WebAssembly.Memory)) throw new Error('Missing Wasm export: memory');
    next.exports._initialize();
    instance = next;
  }
  return {
    run(body) {
      if (busy) throw new Error('Concurrent Wasm crossing');
      busy = true;
      try {
        if (!instance) initialize();
        const context = {
          api: instance.exports, alive: true,
          invoke(name, ...args) {
            try { return this.api[name](...args); }
            catch (error) { this.alive = false; instance = undefined; throw error; }
          },
        };
        return body(context);
      } finally { busy = false; }
    },
  };
}
