// A deliberately small Preview 1 host. There is no filesystem or preopened fd.
export class WASIExit extends Error {
  constructor(code) {
    super(`WASI proc_exit(${code})`);
    this.name = 'WASIExit';
    this.code = code >>> 0;
  }
}

export function createWASI(getMemory, logger = console) {
  const streams = new Map([1, 2].map(fd => [fd, {
    decoder: new TextDecoder('utf-8', { ignoreBOM: true }), pending: '',
  }]));
  function emit(fd, text) {
    const stream = streams.get(fd);
    stream.pending += text;
    let newline;
    while ((newline = stream.pending.indexOf('\n')) !== -1) {
      const line = stream.pending.slice(0, newline);
      stream.pending = stream.pending.slice(newline + 1);
      if (fd === 1) logger.log(line);
      else logger.error(line);
    }
  }
  function flush() {
    for (const [fd, stream] of streams) {
      emit(fd, stream.decoder.decode());
      if (stream.pending !== '') {
        const line = stream.pending;
        stream.pending = '';
        if (fd === 1) logger.log(line);
        else logger.error(line);
      }
    }
  }
  const imports = {
    fd_close() { return 8; }, // EBADF: the logging descriptors cannot be closed.
    fd_fdstat_get(fd, pointer) {
      if (!streams.has(fd)) return 8;
      const memory = getMemory().buffer;
      new Uint8Array(memory, pointer >>> 0, 24).fill(0);
      const view = new DataView(memory);
      view.setUint8(pointer >>> 0, 2); // CHARACTER_DEVICE
      view.setBigUint64((pointer >>> 0) + 8, 64n, true); // FD_WRITE
      return 0;
    },
    fd_prestat_get() { return 8; },
    fd_prestat_dir_name() { return 8; },
    fd_seek(fd) { return streams.has(fd) ? 70 : 8; }, // ESPIPE
    fd_write(fd, vectors, count, written) {
      const stream = streams.get(fd);
      if (!stream) return 8;
      const memory = getMemory().buffer;
      const view = new DataView(memory);
      let total = 0;
      for (let index = 0; index < (count >>> 0); index++) {
        const offset = (vectors >>> 0) + index * 8;
        const pointer = view.getUint32(offset, true);
        const length = view.getUint32(offset + 4, true);
        emit(fd, stream.decoder.decode(new Uint8Array(memory, pointer, length), { stream: true }));
        total += length;
      }
      view.setUint32(written >>> 0, total, true);
      return 0;
    },
    proc_exit(code) {
      flush();
      throw new WASIExit(code);
    },
  };
  return { imports, flush };
}

export function validateImports(module, imports) {
  for (const entry of WebAssembly.Module.imports(module)) {
    if (entry.module !== 'wasi_snapshot_preview1' || entry.kind !== 'function' ||
        !Object.hasOwn(imports, entry.name)) {
      throw new Error(`Unsupported Wasm import: ${entry.module}.${entry.name} (${entry.kind})`);
    }
  }
}
