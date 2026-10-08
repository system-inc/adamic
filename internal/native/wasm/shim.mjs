// Callbacks synchronously receive owned byte copies, never decoded text.
export class AdamicExit extends Error {
	constructor(code) { super(`Adamic exited with code ${code}`); this.name = 'AdamicExit'; this.code = code; }
}

export function wasiShim({ stdout, stderr }) {
	let memory;
	const zeroSizes = (count, size) => {
		const view = new DataView(memory.buffer);
		view.setUint32(count, 0, true); view.setUint32(size, 0, true); return 0;
	};
	const services = {
		fd_write(fd, iovecs, count, written) {
			if (fd !== 1 && fd !== 2) return 8;
			const view = new DataView(memory.buffer), chunks = [];
			let total = 0;
			for (let index = 0; index < count; index++) {
				const pointer = view.getUint32(iovecs + index * 8, true);
				const length = view.getUint32(iovecs + index * 8 + 4, true);
				chunks.push(new Uint8Array(memory.buffer, pointer, length).slice()); total += length;
			}
			for (const bytes of chunks) (fd === 1 ? stdout : stderr)(bytes);
			new DataView(memory.buffer).setUint32(written, total, true); return 0;
		},
		proc_exit(code) { throw new AdamicExit(code); },
		fd_prestat_get: () => 8,
		args_sizes_get: zeroSizes, environ_sizes_get: zeroSizes,
		args_get: () => 0, environ_get: () => 0,
	};
	const imports = { wasi_snapshot_preview1: new Proxy(services, {
		get(target, name) {
			if (Object.hasOwn(target, name)) return target[name];
			if (typeof name === 'string' && /^(fd_|path_)/.test(name)) return () => 8;
			return () => { const error = new Error(`Unsupported WASI import: ${String(name)}`); error.name = 'AdamicUnsupportedWASI'; throw error; };
		},
	}) };
	return { imports, attach(instance) { memory = instance.exports.memory; } };
}
