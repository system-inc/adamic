// ABI version 1 reference host. WASI initialization belongs to the caller.
export function encodeWTF8(text) {
 const bytes = [];
 for (let i = 0; i < text.length; i++) {
  let point = text.charCodeAt(i);
  if (point >= 0xd800 && point <= 0xdbff && i + 1 < text.length) {
   const low = text.charCodeAt(i + 1);
   if (low >= 0xdc00 && low <= 0xdfff) { point = 0x10000 + (point - 0xd800) * 1024 + low - 0xdc00; i++; }
  }
  if (point < 128) bytes.push(point);
  else if (point < 2048) bytes.push(0xc0 | point >> 6, 0x80 | point & 63);
  else if (point < 65536) bytes.push(0xe0 | point >> 12, 0x80 | point >> 6 & 63, 0x80 | point & 63);
  else bytes.push(0xf0 | point >> 18, 0x80 | point >> 12 & 63, 0x80 | point >> 6 & 63, 0x80 | point & 63);
 }
 return Uint8Array.from(bytes);
}
export function decodeWTF8(bytes) {
 const parts = [];
 for (let i = 0; i < bytes.length;) {
  const lead = bytes[i++];
  let point, count;
  if (lead < 128) { point = lead; count = 0; }
  else if (lead >= 0xc2 && lead <= 0xdf) { point = lead & 31; count = 1; }
  else if (lead >= 0xe0 && lead <= 0xef) { point = lead & 15; count = 2; }
  else if (lead >= 0xf0 && lead <= 0xf4) { point = lead & 7; count = 3; }
  else throw new Error('Invalid WTF-8 lead');
  if (i + count > bytes.length) throw new Error('Truncated WTF-8');
  for (let j = 0; j < count; j++) { const next = bytes[i++]; if ((next & 0xc0) !== 0x80) throw new Error('Invalid WTF-8 continuation'); point = point * 64 + (next & 63); }
  if ((count === 1 && point < 128) || (count === 2 && point < 2048) || (count === 3 && point < 65536) || point > 0x10ffff) throw new Error('Invalid WTF-8 point');
  parts.push(String.fromCodePoint(point));
 }
 return parts.join('');
}
class Writer {
 constructor() { this.bytes = new Uint8Array(128); this.offset = 0; }
 reserve(count) {
  if (this.offset + count > 0xffffffff) throw new RangeError('ABI buffer too large');
  if (this.offset + count > this.bytes.length) { const next = new Uint8Array(Math.max(this.bytes.length * 2, this.offset + count)); next.set(this.bytes); this.bytes = next; }
 }
 align() { const padding = (8 - this.offset % 8) % 8; this.reserve(padding); this.offset += padding; }
 u32(value) { this.reserve(4); new DataView(this.bytes.buffer).setUint32(this.offset, value, true); this.offset += 4; }
 number(value) { this.align(); this.reserve(8); new DataView(this.bytes.buffer).setFloat64(this.offset, value, true); this.offset += 8; }
 raw(bytes) { this.reserve(bytes.length); this.bytes.set(bytes, this.offset); this.offset += bytes.length; }
 value(type, value) {
  switch (type.kind) {
   case 'number': this.number(value); break;
   case 'boolean': this.raw(Uint8Array.of(value ? 1 : 0)); break;
   case 'string': { const bytes = encodeWTF8(value); this.u32(bytes.length); this.raw(bytes); break; }
   case 'number[]': this.u32(value.length); this.align(); for (const item of value) this.number(item); break;
   case 'string[]': this.u32(value.length); for (const item of value) this.value({kind: 'string'}, item); break;
   case 'record': for (const field of type.fields ?? []) this.value(field.type, value[field.name]); break;
   default: throw new Error(`Unknown ABI type ${type.kind}`);
  }
 }
 finish() { return this.bytes.slice(0, this.offset); }
}
class Reader {
 constructor(bytes) { this.bytes = bytes; this.view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength); this.offset = 0; }
 take(count) { if (count > this.bytes.length - this.offset) throw new Error('Truncated ABI result'); const start = this.offset; this.offset += count; return start; }
 align() { this.take((8 - this.offset % 8) % 8); }
 u32() { return this.view.getUint32(this.take(4), true); }
 number() { this.align(); return this.view.getFloat64(this.take(8), true); }
 value(type) {
  switch (type.kind) {
   case 'number': return this.number();
   case 'boolean': { const value = this.bytes[this.take(1)]; if (value > 1) throw new Error('Invalid ABI boolean'); return value === 1; }
   case 'string': { const count = this.u32(); const start = this.take(count); return decodeWTF8(this.bytes.subarray(start, start + count)); }
   case 'number[]': { const count = this.u32(); this.align(); const result = []; for (let i = 0; i < count; i++) result.push(this.number()); return result; }
   case 'string[]': { const count = this.u32(); const result = []; for (let i = 0; i < count; i++) result.push(this.value({kind: 'string'})); return result; }
   case 'record': { const result = {}; for (const field of type.fields ?? []) Object.defineProperty(result, field.name, {value: this.value(field.type), enumerable: true, writable: true, configurable: true}); return result; }
   default: throw new Error(`Unknown ABI type ${type.kind}`);
  }
 }
}
export function encodeSchema(type, value) { const writer = new Writer(); writer.value(type, value); return writer.finish(); }
export function decodeSchema(type, bytes) { const reader = new Reader(bytes); const result = reader.value(type); if (reader.offset !== bytes.length) throw new Error('Trailing ABI result bytes'); return result; }
export function wrapExports(instance, table) {
 if (table.version !== 1) throw new Error('Unsupported Adamic ABI version');
 const wasm = instance.exports;
 const result = Object.create(null);
 let failed = false;
 for (const signature of table.exports) {
  const call = wasm[`adamic_export_${signature.name}`];
  if (typeof call !== 'function') throw new Error(`Missing export ${signature.name}`);
  result[signature.name] = (...values) => {
   if (failed) throw new Error('Discard terminated Adamic instance');
   if (values.length !== signature.parameters.length) throw new TypeError('ABI argument count mismatch');
   const owned = [];
   let handle;
   try {
    const arguments_ = [];
    signature.parameters.forEach((parameter, index) => {
     const type = parameter.type, value = values[index];
     if (type.kind === 'number') { arguments_.push(value); return; }
     if (type.kind === 'boolean') { arguments_.push(value ? 1 : 0); return; }
     let bytes, length;
     if (type.kind === 'string') { bytes = encodeWTF8(value); length = bytes.length; }
     else if (type.kind === 'number[]') { bytes = new Uint8Array(value.length * 8); const view = new DataView(bytes.buffer); value.forEach((item, i) => view.setFloat64(i * 8, item, true)); length = value.length; }
     else { bytes = encodeSchema(type, value); length = bytes.length; }
     const pointer = wasm.adamic_alloc(bytes.length) >>> 0; owned.push(pointer);
     new Uint8Array(wasm.memory.buffer, pointer, bytes.length).set(bytes);
     arguments_.push(pointer, length);
    });
    let returned;
    try { returned = call(...arguments_); } catch (error) { failed = true; throw error; }
    switch (signature.returns.kind) {
     case 'void': return undefined;
     case 'number': return returned;
     case 'boolean': return returned !== 0;
     default:
      handle = returned >>> 0;
      const pointer = wasm.adamic_result_bytes(handle) >>> 0;
      const length = wasm.adamic_result_length(handle) >>> 0;
      const bytes = new Uint8Array(wasm.memory.buffer, pointer, length);
      return signature.returns.kind === 'string' ? decodeWTF8(bytes) : decodeSchema(signature.returns, bytes);
    }
   } finally {
    if (!failed) {
     if (handle !== undefined) wasm.adamic_result_release(handle);
     for (const pointer of owned) wasm.adamic_free(pointer);
    }
   }
  };
 }
 return result;
}
export async function instantiateExports(module, table, imports, initialize = instance => instance.exports._initialize()) {
 const instance = await WebAssembly.instantiate(module, imports);
 initialize(instance);
 return {instance, functions: wrapExports(instance, table)};
}
