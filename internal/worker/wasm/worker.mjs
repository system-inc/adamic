import module from './handler.wasm';
import { createHost } from './host.mjs';
import { createWorker } from './bridge.mjs';

const host = createHost(module);
export default createWorker(host);
